package types

import (
	"context"
	"fmt"
)

type contractRuntimeInfoContextKey struct{}

func ContractRuntimeFromContext(ctx context.Context) ContractRuntimeInfo {
	return ctx.Value(contractRuntimeInfoContextKey{}).(ContractRuntimeInfo)
}

func ContextWithContractRuntimeInfo(ctx context.Context, runtimeInfo ContractRuntimeInfo) context.Context {
	return context.WithValue(ctx, contractRuntimeInfoContextKey{}, runtimeInfo)
}

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
