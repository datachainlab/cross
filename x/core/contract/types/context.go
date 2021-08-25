package types

import (
	"context"
	"fmt"

	authtypes "github.com/datachainlab/cross/x/core/auth/types"
)

type contractRuntimeInfoContextKey struct{}

// ContractRuntimeFromContext returns the ContractRuntimeInfo from context
func ContractRuntimeFromContext(ctx context.Context) ContractRuntimeInfo {
	return ctx.Value(contractRuntimeInfoContextKey{}).(ContractRuntimeInfo)
}

// ContextWithContractRuntimeInfo returns a context with an updated ContractRuntimeInfo
func ContextWithContractRuntimeInfo(ctx context.Context, runtimeInfo ContractRuntimeInfo) context.Context {
	return context.WithValue(ctx, contractRuntimeInfoContextKey{}, runtimeInfo)
}

// CommitModeFromContext returns the CommitMode from context
func CommitModeFromContext(ctx context.Context) CommitMode {
	switch v := ctx.Value(contractRuntimeInfoContextKey{}).(type) {
	case ContractRuntimeInfo:
		return v.CommitMode
	case nil:
		return UnspecifiedMode
	default:
		panic(fmt.Sprintf("unknown type: %T", v))
	}
}

type contractSignersContextKey struct{}

// ContractSignersFromContext returns the []AccountID from context
func ContractSignersFromContext(ctx context.Context) []authtypes.AccountID {
	return ctx.Value(contractSignersContextKey{}).([]authtypes.AccountID)
}

// ContextWithContractSigners returns a context with an updated accounts
func ContextWithContractSigners(ctx context.Context, accounts []authtypes.Account) context.Context {
	return context.WithValue(ctx, contractSignersContextKey{}, accounts)
}
