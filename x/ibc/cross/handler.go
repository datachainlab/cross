package cross

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	"github.com/datachainlab/cross/x/ibc/cross/types"
)

// NewHandler returns a handler
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case MsgInitiate:
			return handleMsgInitiate(ctx, keeper, msg)
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
		case PacketDataPrepare:
			return handlePacketDataPrepare(ctx, keeper, contractHandler, packet, data)
		case PacketDataCommit:
			return handlePacketDataCommit(ctx, keeper, contractHandler, packet, data)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC packet data type: %T", data)
		}
	}
}

type PacketAcknowledgementReceiver func(ctx sdk.Context, packet channeltypes.Packet, ack PacketAcknowledgement) (*sdk.Result, error)

func NewPacketAcknowledgementReceiver(keeper Keeper) PacketAcknowledgementReceiver {
	return func(ctx sdk.Context, packet channeltypes.Packet, ack PacketAcknowledgement) (*sdk.Result, error) {
		var data PacketData
		if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC packet type: %T", packet)
		}
		switch ack := ack.(type) {
		case PacketPrepareAcknowledgement:
			return handlePacketPrepareAcknowledgement(ctx, keeper, packet, ack, data.(PacketDataPrepare))
		case PacketCommitAcknowledgement:
			return handlePacketCommitAcknowledgement(ctx, keeper, packet, ack, data.(PacketDataCommit))
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
func handleMsgInitiate(ctx sdk.Context, k Keeper, msg MsgInitiate) (*sdk.Result, error) {
	txID, err := k.MulticastPreparePacket(ctx, msg.Sender, msg, msg.ContractTransactions)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrFailedInitiateTx, err.Error())
	}
	return &sdk.Result{Data: txID[:], Events: ctx.EventManager().ABCIEvents()}, nil
}

/*
Precondition:
- Given proof of packet is valid.
Steps:
- Try to apply given contract transaction to our state.
- If it was success, precommit these changes and get locks for concerned keys. Furthermore, send a Prepacket with status 'OK' to coordinator.
- If it was failed, discard theses changes. Furthermore, send a Prepacket with status 'Failed' to coordinator.
*/
func handlePacketDataPrepare(ctx sdk.Context, k Keeper, contractHandler ContractHandler, packet channeltypes.Packet, data PacketDataPrepare) (*sdk.Result, error) {
	status, err := k.PrepareTransaction(ctx, contractHandler, packet.SourcePort, packet.SourceChannel, packet.DestinationPort, packet.DestinationChannel, data)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrFailedPrepare, err.Error())
	}
	// Send a Prepared Packet to coordinator (reply to source channel)
	ack := types.NewPacketPrepareAcknowledgement(status)
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
func handlePacketPrepareAcknowledgement(ctx sdk.Context, k Keeper, packet channeltypes.Packet, ack PacketPrepareAcknowledgement, data PacketDataPrepare) (*sdk.Result, error) {
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
func handlePacketDataCommit(ctx sdk.Context, k Keeper, contractHandler ContractHandler, packet channeltypes.Packet, data PacketDataCommit) (*sdk.Result, error) {
	res, err := k.ReceiveCommitPacket(ctx, contractHandler, packet.SourcePort, packet.SourceChannel, packet.DestinationPort, packet.DestinationChannel, data)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrFailedReceiveCommitPacket, err.Error())
	}
	ctx.EventManager().EmitEvents(res.GetEvents())
	if err := k.PacketExecuted(ctx, packet, NewPacketCommitAcknowledgement().GetBytes()); err != nil {
		return nil, sdkerrors.Wrap(types.ErrFailedSendAckCommitPacket, err.Error())
	}
	return &sdk.Result{Data: res.GetData(), Events: ctx.EventManager().ABCIEvents()}, nil
}

func handlePacketCommitAcknowledgement(ctx sdk.Context, k Keeper, packet channeltypes.Packet, ack PacketCommitAcknowledgement, data PacketDataCommit) (*sdk.Result, error) {
	err := k.ReceiveAckPacket(ctx, data.TxID, data.TxIndex)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrFailedReceiveAckCommitPacket, err.Error())
	}
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}
