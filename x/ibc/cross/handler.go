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
func NewHandler(keeper Keeper, contractHandler ContractHandler) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case MsgInitiate:
			return handleMsgInitiate(ctx, keeper, contractHandler, msg)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC message type: %T", msg)
		}
	}
}

type PacketReceiver func(ctx sdk.Context, packet channeltypes.Packet) (*sdk.Result, error)

func NewPacketReceiver(keeper Keeper, contractHandler ContractHandler) PacketReceiver {
	return func(ctx sdk.Context, packet channeltypes.Packet) (*sdk.Result, error) {
		var data PacketData
		if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC packet type: %T", packet)
		}
		switch data := data.(type) {
		case simpletypes.PacketDataCall:
			return handlePacketDataCall(ctx, keeper.SimpleKeeper(), contractHandler, packet, data)
		case tpctypes.PacketDataPrepare:
			return handlePacketDataPrepare(ctx, keeper.TPCKeeper(), contractHandler, packet, data)
		case tpctypes.PacketDataCommit:
			return handlePacketDataCommit(ctx, keeper.TPCKeeper(), contractHandler, packet, data)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC packet data type: %T", data)
		}
	}
}

type PacketAcknowledgementReceiver func(ctx sdk.Context, packet channeltypes.Packet, ack PacketAcknowledgement) (*sdk.Result, error)

func NewPacketAcknowledgementReceiver(keeper Keeper, contractHandler ContractHandler) PacketAcknowledgementReceiver {
	return func(ctx sdk.Context, packet channeltypes.Packet, ack PacketAcknowledgement) (*sdk.Result, error) {
		var data PacketData
		if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC packet type: %T", packet)
		}
		switch ack := ack.(type) {
		case simpletypes.PacketCallAcknowledgement:
			return handlePacketCallAcknowledgement(ctx, keeper.SimpleKeeper(), contractHandler, packet, ack, data.(simpletypes.PacketDataCall))
		case tpctypes.PacketPrepareAcknowledgement:
			return handlePacketPrepareAcknowledgement(ctx, keeper.TPCKeeper(), packet, ack, data.(tpctypes.PacketDataPrepare))
		case tpctypes.PacketCommitAcknowledgement:
			return handlePacketCommitAcknowledgement(ctx, keeper.TPCKeeper(), packet, ack, data.(tpctypes.PacketDataCommit))
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC packet data type: %T", data)
		}
	}
}

/*
Steps:
- Ensure that all channels in ContractTransactions are correct
- Multicast a Prepare packet to each participants
*/
func handleMsgInitiate(ctx sdk.Context, k Keeper, contractHandler ContractHandler, msg MsgInitiate) (*sdk.Result, error) {
	var data []byte
	switch msg.CommitProtocol {
	case types.COMMIT_PROTOCOL_SIMPLE:
		txID, err := k.SimpleKeeper().SendCall(ctx, contractHandler, msg, msg.ContractTransactions)
		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrFailedInitiateTx, err.Error())
		}
		data = txID[:]
	case types.COMMIT_PROTOCOL_TPC:
		txID, err := k.TPCKeeper().MulticastPreparePacket(ctx, msg.Sender, msg, msg.ContractTransactions)
		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrFailedInitiateTx, err.Error())
		}
		data = txID[:]
	default:
		return nil, fmt.Errorf("unknown Commit protocol '%v'", msg.CommitProtocol)
	}

	return &sdk.Result{Data: data, Events: ctx.EventManager().ABCIEvents()}, nil
}

func handlePacketDataCall(ctx sdk.Context, k simple.Keeper, contractHandler ContractHandler, packet channeltypes.Packet, data simpletypes.PacketDataCall) (*sdk.Result, error) {
	status, err := k.ReceiveCallPacket(ctx, contractHandler, packet.DestinationPort, packet.DestinationChannel, data)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrFailedPrepare, err.Error())
	}
	ack := simpletypes.NewPacketCallAcknowledgement(status)
	if err := k.PacketExecuted(ctx, packet, ack.GetBytes()); err != nil {
		return nil, sdkerrors.Wrap(types.ErrFailedPrepare, err.Error())
	}
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handlePacketCallAcknowledgement(ctx sdk.Context, k simple.Keeper, contractHandler ContractHandler, packet channeltypes.Packet, ack simpletypes.PacketCallAcknowledgement, data simpletypes.PacketDataCall) (*sdk.Result, error) {
	isCommitable, err := k.ReceiveCallAcknowledgement(ctx, packet.SourcePort, packet.SourceChannel, ack, data.TxID)
	if err != nil {
		return nil, err
	}
	res, err := k.TryCommit(ctx, contractHandler, data.TxID, isCommitable)
	if err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvents(res.GetEvents())
	return &sdk.Result{Data: res.GetData(), Events: ctx.EventManager().ABCIEvents()}, nil
}

/*
Precondition:
- Given proof of packet is valid.
Steps:
- Try to apply given contract transaction to our state.
- If it was success, precommit these changes and get locks for concerned keys. Furthermore, send a Prepacket with status 'OK' to coordinator.
- If it was failed, discard theses changes. Furthermore, send a Prepacket with status 'Failed' to coordinator.
*/
func handlePacketDataPrepare(ctx sdk.Context, k tpc.Keeper, contractHandler ContractHandler, packet channeltypes.Packet, data tpctypes.PacketDataPrepare) (*sdk.Result, error) {
	status, err := k.Prepare(ctx, contractHandler, packet.DestinationPort, packet.DestinationChannel, data)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrFailedPrepare, err.Error())
	}
	ack := tpctypes.NewPacketPrepareAcknowledgement(status)
	if err := k.PacketExecuted(ctx, packet, ack.GetBytes()); err != nil {
		return nil, sdkerrors.Wrap(types.ErrFailedPrepare, err.Error())
	}
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
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
func handlePacketPrepareAcknowledgement(ctx sdk.Context, k tpc.Keeper, packet channeltypes.Packet, ack tpctypes.PacketPrepareAcknowledgement, data tpctypes.PacketDataPrepare) (*sdk.Result, error) {
	canMulticast, isCommitable, err := k.ReceivePrepareAcknowledgement(ctx, packet.SourcePort, packet.SourceChannel, ack, data.TxID, data.TxIndex)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrFailedRecievePrepareResult, err.Error())
	}
	if canMulticast {
		if err := k.MulticastCommitPacket(ctx, data.TxID, isCommitable); err != nil {
			return nil, sdkerrors.Wrap(types.ErrFailedMulticastCommitPacket, err.Error())
		}
		return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
	} else {
		return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
	}
}

/*
Precondition:
- Given proof of packet is valid.
Steps:
- If PacketDataCommit indicates committable, commit precommitted state and unlock locked keys.
- If PacketDataCommit indicates uncommittable, rollback precommitted state and unlock locked keys.
*/
func handlePacketDataCommit(ctx sdk.Context, k tpc.Keeper, contractHandler ContractHandler, packet channeltypes.Packet, data tpctypes.PacketDataCommit) (*sdk.Result, error) {
	res, err := k.ReceiveCommitPacket(ctx, contractHandler, packet.SourcePort, packet.SourceChannel, packet.DestinationPort, packet.DestinationChannel, data)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrFailedReceiveCommitPacket, err.Error())
	}
	ctx.EventManager().EmitEvents(res.GetEvents())
	if err := k.PacketExecuted(ctx, packet, tpctypes.NewPacketCommitAcknowledgement().GetBytes()); err != nil {
		return nil, sdkerrors.Wrap(types.ErrFailedSendAckCommitPacket, err.Error())
	}
	return &sdk.Result{Data: res.GetData(), Events: ctx.EventManager().ABCIEvents()}, nil
}

func handlePacketCommitAcknowledgement(ctx sdk.Context, k tpc.Keeper, packet channeltypes.Packet, ack tpctypes.PacketCommitAcknowledgement, data tpctypes.PacketDataCommit) (*sdk.Result, error) {
	err := k.PacketCommitAcknowledgement(ctx, data.TxID, data.TxIndex)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrFailedReceiveAckCommitPacket, err.Error())
	}
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}
