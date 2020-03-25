package cross

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	"github.com/datachainlab/cross/x/ibc/cross/types"
)

// NewHandler returns a handler
func NewHandler(keeper Keeper, contractHandler ContractHandler) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case MsgInitiate:
			return handleMsgInitiate(ctx, keeper, msg)
		case channeltypes.MsgPacket:
			switch data := msg.Data.(type) {
			case PacketDataPrepare:
				return handlePacketDataPrepare(ctx, keeper, contractHandler, msg, data)
			case PacketDataPrepareResult:
				return handlePacketDataPrepareResult(ctx, keeper, msg, data)
			case PacketDataCommit:
				return handlePacketDataCommit(ctx, keeper, contractHandler, msg, data)
			default:
				return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized packet data type: %T", data)
			}
		case channeltypes.MsgAcknowledgement:
			switch ack := msg.Acknowledgement.(type) {
			case AckDataCommit:
				switch data := msg.Data.(type) {
				case PacketDataCommit:
					return handleAcknowledgePacket(ctx, keeper, msg, ack, data)
				default:
					return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized ack packet data type: %T", data)
				}
			default:
				return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized ack packet type: %T", ack)
			}
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC message type: %T", msg)
		}
	}
}

/*
Steps:
- Ensure that all channels in ContractTransactions are correct
- Multicast a Prepare packet to each participants
*/
func handleMsgInitiate(ctx sdk.Context, k Keeper, msg MsgInitiate) (*sdk.Result, error) {
	err := k.MulticastPreparePacket(ctx, msg.Sender, msg, msg.ContractTransactions)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrFailedInitiateTx, err.Error())
	}
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

/*
Precondition:
- Given proof of packet is valid.
Steps:
- Try to apply given contract transaction to our state.
- If it was success, precommit these changes and get locks for concerned keys. Furthermore, send a Prepacket with status 'OK' to coordinator.
- If it was failed, discard theses changes. Furthermore, send a Prepacket with status 'Failed' to coordinator.
*/
func handlePacketDataPrepare(ctx sdk.Context, k Keeper, contractHandler ContractHandler, msg channeltypes.MsgPacket, data PacketDataPrepare) (*sdk.Result, error) {
	err := k.PrepareTransaction(ctx, contractHandler, msg.SourcePort, msg.SourceChannel, msg.DestinationPort, msg.DestinationChannel, data, msg.Signer)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrFailedPrepare, err.Error())
	}
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
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
func handlePacketDataPrepareResult(ctx sdk.Context, k Keeper, msg channeltypes.MsgPacket, data PacketDataPrepareResult) (*sdk.Result, error) {
	canMulticast, isCommitable, err := k.ReceivePrepareResultPacket(ctx, msg.Packet, data)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrFailedRecievePrepareResult, err.Error())
	}
	if canMulticast {
		if err := k.MulticastCommitPacket(ctx, data.TxID, msg.Signer, isCommitable); err != nil {
			return nil, sdkerrors.Wrap(types.ErrFailedMulticastCommitPacket, err.Error())
		}
		return &sdk.Result{Events: ctx.EventManager().Events()}, nil
	} else {
		return &sdk.Result{Events: ctx.EventManager().Events()}, nil
	}
}

/*
Precondition:
- Given proof of packet is valid.
Steps:
- If PacketDataCommit indicates committable, commit precommitted state and unlock locked keys.
- If PacketDataCommit indicates uncommittable, rollback precommitted state and unlock locked keys.
*/
func handlePacketDataCommit(ctx sdk.Context, k Keeper, contractHandler ContractHandler, msg channeltypes.MsgPacket, data PacketDataCommit) (*sdk.Result, error) {
	err := k.ReceiveCommitPacket(ctx, contractHandler, msg.SourcePort, msg.SourceChannel, msg.DestinationPort, msg.DestinationChannel, data)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrFailedReceiveCommitPacket, err.Error())
	}

	// FIXME set transactionID that is taken from packet or state
	acknowledgement := NewAckDataCommit(0)
	if err := k.PacketExecuted(ctx, msg.Packet, acknowledgement); err != nil {
		return nil, err
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleAcknowledgePacket(ctx sdk.Context, k Keeper, msg channeltypes.MsgAcknowledgement, ack AckDataCommit, data PacketDataCommit) (*sdk.Result, error) {
	if err := k.ReceiveAckPacket(ctx, ack, data.TxID); err != nil {
		return nil, err
	}
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
