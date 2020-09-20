package cross

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	"github.com/datachainlab/cross/x/ibc/cross/keeper/simple"
	"github.com/datachainlab/cross/x/ibc/cross/keeper/tpc"
	"github.com/datachainlab/cross/x/ibc/cross/types"
	simpletypes "github.com/datachainlab/cross/x/ibc/cross/types/simple"
	tpctypes "github.com/datachainlab/cross/x/ibc/cross/types/tpc"
)

// NewHandler returns a handler
func NewHandler(keeper Keeper, packetMiddleware types.PacketMiddleware, contractHandler ContractHandler) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx, ps, err := packetMiddleware.HandleMsg(ctx, msg, types.NewSimplePacketSender(keeper.ChannelKeeper()))
		if err != nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "failed to handle request: %v", err)
		}
		switch msg := msg.(type) {
		case MsgInitiate:
			return handleMsgInitiate(ctx, keeper, ps, contractHandler, msg)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC message type: %T", msg)
		}
	}
}

type PacketReceiver func(ctx sdk.Context, packet channeltypes.Packet) (*sdk.Result, error)

func NewPacketReceiver(keeper Keeper, packetMiddleware types.PacketMiddleware, contractHandler ContractHandler) PacketReceiver {
	return func(ctx sdk.Context, packet channeltypes.Packet) (*sdk.Result, error) {
		ip, err := types.UnmarshalIncomingPacket(types.ModuleCdc, packet)
		if err != nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC packet type: %T: %v", packet, err)
		}
		ctx, _, as, err := packetMiddleware.HandlePacket(ctx, ip, types.NewSimplePacketSender(keeper.ChannelKeeper()), types.NewSimpleACKSender())
		if err != nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "failed to handle request: %v", err)
		}
		var resData []byte
		var ack types.OutgoingPacketAcknowledgement
		switch payload := ip.Payload().(type) {
		case simpletypes.PacketDataCall:
			resData, ack, err = handlePacketDataCall(ctx, keeper.SimpleKeeper(), contractHandler, packet, payload)
		case tpctypes.PacketDataPrepare:
			resData, ack, err = handlePacketDataPrepare(ctx, keeper.TPCKeeper(), contractHandler, packet, payload)
		case tpctypes.PacketDataCommit:
			resData, ack, err = handlePacketDataCommit(ctx, keeper.TPCKeeper(), contractHandler, packet, payload)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC packet payload type: %T", payload)
		}
		if err != nil {
			return nil, err
		}
		if err = as.SendACK(ctx, ack); err != nil {
			return nil, err
		}
		bz, err := types.MarshalJSONPacketAcknowledgementData(ack.Data())
		if err != nil {
			return nil, err
		}
		if err := keeper.PacketExecuted(ctx, packet, bz); err != nil {
			return nil, err
		}
		return &sdk.Result{Data: resData, Events: ctx.EventManager().ABCIEvents()}, nil
	}
}

type PacketAcknowledgementReceiver func(ctx sdk.Context, packet channeltypes.Packet, ack types.IncomingPacketAcknowledgement) (*sdk.Result, error)

func NewPacketAcknowledgementReceiver(keeper Keeper, packetMiddleware types.PacketMiddleware, contractHandler ContractHandler) PacketAcknowledgementReceiver {
	return func(ctx sdk.Context, packet channeltypes.Packet, ack types.IncomingPacketAcknowledgement) (*sdk.Result, error) {
		pi, err := types.UnmarshalIncomingPacket(types.ModuleCdc, packet)
		if err != nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC packet type: %T: %v", packet, err)
		}
		ctx, ps, err := packetMiddleware.HandleACK(ctx, pi, ack, types.NewSimplePacketSender(keeper.ChannelKeeper()))
		if err != nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "failed to handle request: %v", err)
		}
		var resData []byte
		switch ack := ack.Payload().(type) {
		case simpletypes.PacketCallAcknowledgement:
			resData, err = handlePacketCallAcknowledgement(ctx, keeper.SimpleKeeper(), contractHandler, packet, ack, pi.Payload().(simpletypes.PacketDataCall))
		case tpctypes.PacketPrepareAcknowledgement:
			resData, err = handlePacketPrepareAcknowledgement(ctx, keeper.TPCKeeper(), ps, packet, ack, pi.Payload().(tpctypes.PacketDataPrepare))
		case tpctypes.PacketCommitAcknowledgement:
			resData, err = handlePacketCommitAcknowledgement(ctx, keeper.TPCKeeper(), packet, ack, pi.Payload().(tpctypes.PacketDataCommit))
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC ack type: %T", ack)
		}
		if err != nil {
			return nil, err
		}
		return &sdk.Result{Data: resData, Events: ctx.EventManager().ABCIEvents()}, nil
	}
}

/*
Steps:
- Ensure that all channels in ContractTransactions are correct
- Multicast a Prepare packet to each participants
*/
func handleMsgInitiate(ctx sdk.Context, k Keeper, packetSender types.PacketSender, contractHandler ContractHandler, msg MsgInitiate) (*sdk.Result, error) {
	var data []byte
	switch msg.CommitProtocol {
	case types.COMMIT_PROTOCOL_SIMPLE:
		txID, err := k.SimpleKeeper().SendCall(ctx, packetSender, contractHandler, msg, msg.ContractTransactions)
		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrFailedInitiateTx, err.Error())
		}
		data = txID[:]
	case types.COMMIT_PROTOCOL_TPC:
		txID, err := k.TPCKeeper().MulticastPreparePacket(ctx, packetSender, msg.Sender, msg, msg.ContractTransactions)
		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrFailedInitiateTx, err.Error())
		}
		data = txID[:]
	default:
		return nil, fmt.Errorf("unknown Commit protocol '%v'", msg.CommitProtocol)
	}

	return &sdk.Result{Data: data, Events: ctx.EventManager().ABCIEvents()}, nil
}

func handlePacketDataCall(ctx sdk.Context, k simple.Keeper, contractHandler ContractHandler, packet channeltypes.Packet, payload simpletypes.PacketDataCall) (res []byte, ack types.OutgoingPacketAcknowledgement, err error) {
	status, err := k.ReceiveCallPacket(ctx, contractHandler, packet.DestinationPort, packet.DestinationChannel, payload)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(types.ErrFailedPrepare, err.Error())
	}
	return nil, types.NewOutgoingPacketAcknowledgement(nil, simpletypes.NewPacketCallAcknowledgement(status)), nil
}

func handlePacketCallAcknowledgement(ctx sdk.Context, k simple.Keeper, contractHandler ContractHandler, packet channeltypes.Packet, ack simpletypes.PacketCallAcknowledgement, payload simpletypes.PacketDataCall) ([]byte, error) {
	isCommitable, err := k.ReceiveCallAcknowledgement(ctx, packet.SourcePort, packet.SourceChannel, ack, payload.TxID)
	if err != nil {
		return nil, err
	}
	res, err := k.TryCommit(ctx, contractHandler, payload.TxID, isCommitable)
	if err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvents(res.GetEvents())
	return res.GetData(), nil
}

/*
Precondition:
- Given proof of packet is valid.
Steps:
- Try to apply given contract transaction to our state.
- If it was success, precommit these changes and get locks for concerned keys. Furthermore, send a Prepacket with status 'OK' to coordinator.
- If it was failed, discard theses changes. Furthermore, send a Prepacket with status 'Failed' to coordinator.
*/
func handlePacketDataPrepare(ctx sdk.Context, k tpc.Keeper, contractHandler ContractHandler, packet channeltypes.Packet, data tpctypes.PacketDataPrepare) (res []byte, ack types.OutgoingPacketAcknowledgement, err error) {
	status, err := k.Prepare(ctx, contractHandler, packet.DestinationPort, packet.DestinationChannel, data)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(types.ErrFailedPrepare, err.Error())
	}
	return nil, types.NewOutgoingPacketAcknowledgement(nil, tpctypes.NewPacketPrepareAcknowledgement(status)), nil
}

/*
Precondition:
- Given proof of packet is valid.
Steps:
- Verify PrepareResultPacket
- If packet status is 'Failed', we send a CommitPacket with status 'Abort' to all participants.
- If packet status is 'OK' and all packets are confirmed, we send a CommitPacket with status 'Commit' to all participants.
- If packet status is 'OK' and we haven't confirmed all packets yet, we wait for next packet receiving.
*/
func handlePacketPrepareAcknowledgement(ctx sdk.Context, k tpc.Keeper, packetSender types.PacketSender, packet channeltypes.Packet, ack tpctypes.PacketPrepareAcknowledgement, data tpctypes.PacketDataPrepare) ([]byte, error) {
	canMulticast, isCommitable, err := k.ReceivePrepareAcknowledgement(ctx, packet.SourcePort, packet.SourceChannel, ack, data.TxID, data.TxIndex)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrFailedRecievePrepareResult, err.Error())
	}
	if canMulticast {
		if err := k.MulticastCommitPacket(ctx, packetSender, data.TxID, isCommitable); err != nil {
			return nil, sdkerrors.Wrap(types.ErrFailedMulticastCommitPacket, err.Error())
		}
		return nil, nil
	} else {
		return nil, nil
	}
}

/*
Precondition:
- Given proof of packet is valid.
Steps:
- If PacketDataCommit indicates committable, commit precommitted state and unlock locked keys.
- If PacketDataCommit indicates uncommittable, rollback precommitted state and unlock locked keys.
*/
func handlePacketDataCommit(ctx sdk.Context, k tpc.Keeper, contractHandler ContractHandler, packet channeltypes.Packet, data tpctypes.PacketDataCommit) (res []byte, ack types.OutgoingPacketAcknowledgement, err error) {
	r, err := k.ReceiveCommitPacket(ctx, contractHandler, packet.SourcePort, packet.SourceChannel, packet.DestinationPort, packet.DestinationChannel, data)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(types.ErrFailedReceiveCommitPacket, err.Error())
	}
	ctx.EventManager().EmitEvents(r.GetEvents())
	return r.GetData(), types.NewOutgoingPacketAcknowledgement(nil, tpctypes.NewPacketCommitAcknowledgement()), nil
}

func handlePacketCommitAcknowledgement(ctx sdk.Context, k tpc.Keeper, packet channeltypes.Packet, ack tpctypes.PacketCommitAcknowledgement, data tpctypes.PacketDataCommit) ([]byte, error) {
	err := k.PacketCommitAcknowledgement(ctx, data.TxID, data.TxIndex)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrFailedReceiveAckCommitPacket, err.Error())
	}
	return nil, nil
}
