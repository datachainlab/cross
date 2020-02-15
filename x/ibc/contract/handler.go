package contract

import (
	"github.com/bluele/crossccc/x/ibc/crossccc"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHandler returns a handler
func NewHandler(contractHandler crossccc.ContractHandler) sdk.Handler {
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

func handleContractCall(ctx sdk.Context, msg MsgContractCall, contractHandler crossccc.ContractHandler) (*sdk.Result, error) {
	state, err := contractHandler.Handle(ctx, msg.Contract)
	if err != nil {
		return nil, err
	}
	if err := state.CommitImmediately(); err != nil {
		return nil, err
	}
	return &sdk.Result{}, nil
}
