package types

import (
	"context"
)

type ContractModule interface {
	OnContractCall(ctx context.Context, callInfo ContractCallInfo) (*ContractCallResult, *OPs, error)
}

type ContractHandler func(ctx context.Context, callInfo ContractCallInfo) (*ContractCallResult, *OPs, error)

type ContractHandleDecorator interface {
	Handle(ctx context.Context, callInfo ContractCallInfo) (newCtx context.Context, err error)
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

func SetupContractContext(ctx context.Context, runtimeInfo ContractRuntimeInfo) context.Context {
	ctx = ContextWithContractRuntimeInfo(ctx, runtimeInfo)
	return ctx
}
