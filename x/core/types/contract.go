package types

import (
	"context"
)

type ContractHandler func(ctx context.Context, callInfo ContractCallInfo) (*OPs, error)

type ContractHandleDecorator interface {
	Handle(ctx context.Context, callInfo ContractCallInfo) (newCtx context.Context, err error)
}

func NewContractHandler(h ContractHandler, decs ...ContractHandleDecorator) ContractHandler {
	if h == nil {
		panic("ContractHandler cannot be nil")
	}
	return func(ctx context.Context, callInfo ContractCallInfo) (*OPs, error) {
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

func SetupContractContext(ctx context.Context, runtimeInfo ContractRuntimeInfo) context.Context {
	ctx = ContextWithContractRuntimeInfo(ctx, runtimeInfo)
	return ctx
}
