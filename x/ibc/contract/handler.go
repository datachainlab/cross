package contract

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/datachainlab/cross/x/ibc/cross"
)

// NewHandler returns a handler
func NewHandler(contractHandler cross.ContractHandler) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case MsgContractCall:
			// TODO call Handle() after verify signatures
			return handleContractCall(ctx, msg, contractHandler)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized message type: %T", msg)
		}
	}
}

func handleContractCall(ctx sdk.Context, msg MsgContractCall, contractHandler cross.ContractHandler) (*sdk.Result, error) {
	ctx = cross.WithSigners(ctx, msg.GetSigners())
	state, err := contractHandler.Handle(ctx, msg.Contract)
	if err != nil {
		return nil, err
	}
	if err := state.CommitImmediately(); err != nil {
		return nil, err
	}
	return &sdk.Result{}, nil
}
