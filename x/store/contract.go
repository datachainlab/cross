package store

import (
	"context"

	"github.com/datachainlab/cross/x/core/types"
)

func DefaultContractHandleDecorator() []types.ContractHandleDecorator {
	return []types.ContractHandleDecorator{
		SetUpContractHandleDecorator{},
	}
}

type SetUpContractHandleDecorator struct{}

var _ types.ContractHandleDecorator = (*SetUpContractHandleDecorator)(nil)

func (cd SetUpContractHandleDecorator) Handle(ctx context.Context, callInfo types.ContractCallInfo) (newCtx context.Context, err error) {
	opmgr, err := GetOPManager(types.ContractRuntimeFromContext(ctx).StateConstraintType)
	if err != nil {
		panic(err)
	}
	return contextWithOPManager(ctx, opmgr), nil
}

type opManagerContextKey struct{}

func opManagerFromContext(ctx context.Context) OPManager {
	return ctx.Value(opManagerContextKey{}).(OPManager)
}

func contextWithOPManager(ctx context.Context, opmgr OPManager) context.Context {
	return context.WithValue(ctx, opManagerContextKey{}, opmgr)
}
