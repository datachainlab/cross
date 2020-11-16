package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ContractModule interface {
	OnContractCall(ctx context.Context, callInfo ContractCallInfo) (*ContractCallResult, *OPs, error)
}

type ContractHandler func(ctx context.Context, callInfo ContractCallInfo) (*ContractCallResult, *OPs, error)

type ContractHandleDecorator interface {
	Handle(ctx context.Context, callInfo ContractCallInfo) (newCtx context.Context, err error)
}

type ContractHandleDecorators []ContractHandleDecorator

var _ ContractHandleDecorator = (*ContractHandleDecorators)(nil)

func (decs ContractHandleDecorators) Handle(ctx context.Context, callInfo ContractCallInfo) (newCtx context.Context, err error) {
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
	return func(ctx context.Context, callInfo ContractCallInfo) (*ContractCallResult, *OPs, error) {
		var err error
		for _, dec := range decs {
			ctx, err = dec.Handle(ctx, callInfo)
			if err != nil {
				return nil, nil, err
			}
		}
		return h(ctx, callInfo)
	}
}

func SetupContractContext(ctx sdk.Context, signers []AccountID, runtimeInfo ContractRuntimeInfo) sdk.Context {
	goCtx := ctx.Context()
	goCtx = ContextWithContractRuntimeInfo(goCtx, runtimeInfo)
	goCtx = ContextWithContractSigners(goCtx, signers)
	return ctx.WithContext(goCtx)
}
