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
	lkmgr := NewLockManager()
	return ContextWithLockManager(ctx, lkmgr), nil
}

type lkManagerContextKey struct{}

func LockManagerFromContext(ctx context.Context) LockManager {
	return ctx.Value(lkManagerContextKey{}).(LockManager)
}

func ContextWithLockManager(ctx context.Context, lkmgr LockManager) context.Context {
	return context.WithValue(ctx, lkManagerContextKey{}, lkmgr)
}
