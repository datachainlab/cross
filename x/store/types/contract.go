package types

import (
	"context"

	crosstypes "github.com/datachainlab/cross/x/core/types"
)

func DefaultContractHandleDecorators() crosstypes.ContractHandleDecorator {
	return crosstypes.ContractHandleDecorators{
		SetUpContractHandleDecorator{},
	}
}

type SetUpContractHandleDecorator struct{}

var _ crosstypes.ContractHandleDecorator = (*SetUpContractHandleDecorator)(nil)

func (cd SetUpContractHandleDecorator) Handle(ctx context.Context, callInfo crosstypes.ContractCallInfo) (newCtx context.Context, err error) {
	opmgr, err := GetOPManager(crosstypes.ContractRuntimeFromContext(ctx).StateConstraintType)
	if err != nil {
		return nil, err
	}
	return ContextWithOPManager(ctx, opmgr), nil
}

type opManagerContextKey struct{}

func OPManagerFromContext(ctx context.Context) OPManager {
	return ctx.Value(opManagerContextKey{}).(OPManager)
}

func ContextWithOPManager(ctx context.Context, opmgr OPManager) context.Context {
	return context.WithValue(ctx, opManagerContextKey{}, opmgr)
}
