package types

import (
	"context"

	contracttypes "github.com/datachainlab/cross/x/contract/types"
	txtypes "github.com/datachainlab/cross/x/tx/types"
)

func DefaultContractHandleDecorators() contracttypes.ContractHandleDecorator {
	return contracttypes.ContractHandleDecorators{
		SetUpContractHandleDecorator{},
	}
}

type SetUpContractHandleDecorator struct{}

var _ contracttypes.ContractHandleDecorator = (*SetUpContractHandleDecorator)(nil)

func (cd SetUpContractHandleDecorator) Handle(ctx context.Context, callInfo txtypes.ContractCallInfo) (newCtx context.Context, err error) {
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
