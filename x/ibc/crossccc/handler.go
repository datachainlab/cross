package crossccc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// NewHandler returns a handler
func NewHandler(keeper Keeper, contractHandler ContractHandler) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case MsgInitiate:
			return handleMsgInitiate(ctx, keeper, msg)
		case MsgConfirm:
			return handleMsgConfirm(ctx, keeper, msg)
		case channeltypes.MsgPacket:
			switch data := msg.Data.(type) {
			case PacketDataInitiate:
				return handlePacketDataInitiate(ctx, keeper, contractHandler, msg, data)
			case PacketDataCommit:
				return handlePacketDataCommit(ctx, keeper, contractHandler, msg, data)
			default:
				return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized ics20 packet data type: %T", data)
			}
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC message type: %T", msg)
		}
	}
}

func handleMsgInitiate(ctx sdk.Context, k Keeper, msg MsgInitiate) (*sdk.Result, error) {
	err := k.MulticastInitiatePacket(ctx, msg.Sender, msg, msg.StateTransitions)
	if err != nil {
		return nil, err
	}
	return &sdk.Result{}, nil
}

func handlePacketDataInitiate(ctx sdk.Context, k Keeper, contractHandler ContractHandler, msg channeltypes.MsgPacket, data PacketDataInitiate) (*sdk.Result, error) {
	/*
		1. ensure that verify a given proof -> verified at ante handler
		2. try to apply given transision to our state
		3. If success, precommit that changes and lock them
		4. If failed, discard changes of precommit

		QUESTION: Each participant node should verify transition and update its packet on block store before 1?
	*/
	// FIXME split this method to verify and create-packet method
	err := k.PrepareTransaction(ctx, contractHandler, msg.SourcePort, msg.SourceChannel, msg.DestinationPort, msg.DestinationChannel, data, msg.Signer)
	if err != nil {
		panic(err)
		// TODO: Source chain sent invalid packet, shutdown channel
	}
	return &sdk.Result{}, nil
}

// This method can be called by valid coordinator
// Precondition:
//   - msg.Statuses are verified by ante
// Optional: We can specify any networks(more than 0) as coordinator via Initiate msg?
func handleMsgConfirm(ctx sdk.Context, k Keeper, msg MsgConfirm) (*sdk.Result, error) {
	// ensure that all prepare packet are successful
	err := k.MulticastCommitPacket(ctx, msg.TxID, msg.PreparePackets, msg.Signer, msg.IsCommittable())
	if err != nil {
		return nil, err
	}
	return &sdk.Result{}, nil
}

func handlePacketDataCommit(ctx sdk.Context, k Keeper, contractHandler ContractHandler, msg channeltypes.MsgPacket, data PacketDataCommit) (*sdk.Result, error) {
	err := k.ReceiveCommitPacket(ctx, contractHandler, msg.SourcePort, msg.SourceChannel, msg.DestinationPort, msg.DestinationChannel, data, msg.Signer)
	if err != nil {
		return nil, err
	}
	return &sdk.Result{}, nil
}
