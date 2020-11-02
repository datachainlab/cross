package types

import "context"

type contractRuntimeInfoContextKey struct{}

func ContractRuntimeFromContext(ctx context.Context) ContractRuntimeInfo {
	return ctx.Value(contractRuntimeInfoContextKey{}).(ContractRuntimeInfo)
}

func ContextWithContractRuntimeInfo(ctx context.Context, runtimeInfo ContractRuntimeInfo) context.Context {
	return context.WithValue(ctx, contractRuntimeInfoContextKey{}, runtimeInfo)
}

func CommitModeFromContext(ctx context.Context) CommitMode {
	return ContractRuntimeFromContext(ctx).CommitMode
}
