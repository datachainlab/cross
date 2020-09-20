package contract

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/datachainlab/cross/x/ibc/contract/keeper"
	"github.com/datachainlab/cross/x/ibc/contract/types"
	"github.com/datachainlab/cross/x/ibc/cross"
)

// NewHandler returns a handler
func NewHandler(k Keeper, contractHandler cross.ContractHandler, router types.ServerRouter) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case MsgContractCall:
			return handleContractCall(ctx, msg, k, contractHandler, router)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized message type: %T", msg)
		}
	}
}

func handleContractCall(ctx sdk.Context, msg MsgContractCall, k Keeper, contractHandler cross.ContractHandler, router types.ServerRouter) (*sdk.Result, error) {
	var rs cross.ObjectResolver
	if keeper.IsSimulation(ctx) {
		rs = types.NewHTTPObjectResolver(router)
	} else {
		rs = cross.NewFakeResolver()
	}

	ctx = cross.WithSigners(ctx, msg.GetSigners())
	state, res, err := contractHandler.Handle(ctx, msg.CallInfo, cross.ContractRuntimeInfo{StateConstraintType: msg.StateConstraintType, ExternalObjectResolver: rs})
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrFailedContractHandle, err.Error())
	}
	if err := state.CommitImmediately(); err != nil {
		return nil, sdkerrors.Wrap(types.ErrFailedCommitStore, err.Error())
	}
	res = contractHandler.OnCommit(ctx, res)
	data, err := k.MakeContractCallResponseData(res.GetData(), state.OPs())
	if err != nil { // internal error
		return nil, err
	}
	ctx.EventManager().EmitEvents(res.GetEvents())
	return &sdk.Result{Data: data, Events: ctx.EventManager().ABCIEvents()}, nil
}
