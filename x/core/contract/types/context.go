package types

import (
	"context"
	"fmt"
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
