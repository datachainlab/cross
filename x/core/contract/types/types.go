package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	accounttypes "github.com/datachainlab/cross/x/core/account/types"
	initiatortypes "github.com/datachainlab/cross/x/core/initiator/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
)

type ContractModule interface {
	OnContractCall(ctx context.Context, callInfo txtypes.ContractCallInfo) (*txtypes.ContractCallResult, error)
}

type ContractHandler func(ctx context.Context, callInfo txtypes.ContractCallInfo) (*txtypes.ContractCallResult, error)

type ContractHandleDecorator interface {
	Handle(ctx context.Context, callInfo txtypes.ContractCallInfo) (newCtx context.Context, err error)
}

type ContractHandleDecorators []ContractHandleDecorator

var _ ContractHandleDecorator = (*ContractHandleDecorators)(nil)

func (decs ContractHandleDecorators) Handle(ctx context.Context, callInfo txtypes.ContractCallInfo) (newCtx context.Context, err error) {
	for _, dec := range decs {
		ctx, err = dec.Handle(ctx, callInfo)
		if err != nil {
			return nil, err
		}
	}
	return ctx, nil
}

func NewContractHandler(h ContractHandler, decs ...ContractHandleDecorator) ContractHandler {
	if h == nil {
		panic("ContractHandler cannot be nil")
	}
	return func(ctx context.Context, callInfo txtypes.ContractCallInfo) (*txtypes.ContractCallResult, error) {
		var err error
		for _, dec := range decs {
			ctx, err = dec.Handle(ctx, callInfo)
			if err != nil {
				return nil, err
			}
		}
		return h(ctx, callInfo)
	}
}

func SetupContractContext(ctx sdk.Context, signers []accounttypes.AccountID, runtimeInfo ContractRuntimeInfo) sdk.Context {
	goCtx := ctx.Context()
	goCtx = ContextWithContractRuntimeInfo(goCtx, runtimeInfo)
	goCtx = ContextWithContractSigners(goCtx, signers)
	return ctx.WithContext(goCtx)
}

type ExternalContractCaller interface {
	Call(ctx sdk.Context, xcc xcctypes.XCC, callInfo txtypes.ContractCallInfo, signers []accounttypes.AccountID) []byte
}

type externalContractCaller struct{}

var _ ExternalContractCaller = (*externalContractCaller)(nil)

func (cc externalContractCaller) Call(ctx sdk.Context, xcc xcctypes.XCC, callInfo txtypes.ContractCallInfo, signers []accounttypes.AccountID) []byte {
	r := ContractRuntimeFromContext(ctx.Context()).ExternalObjectResolver
	key := initiatortypes.MakeObjectKey(callInfo, signers)
	obj, err := r.Resolve(xcc, key)
	if err != nil {
		panic(err)
	}
	v, err := obj.Evaluate(key)
	if err != nil {
		panic(err)
	}
	return v
}

func NewExternalContractCaller() ExternalContractCaller {
	return externalContractCaller{}
}

type ContractRuntimeInfo struct {
	CommitMode             CommitMode
	ExternalObjectResolver txtypes.ObjectResolver
}
