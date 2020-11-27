package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	contracttypes "github.com/datachainlab/cross/x/core/contract/types"
)

// KVStoreI defines the expected key-value store
type KVStoreI interface {
	Prefix(prefix []byte) KVStoreI
	KVStore(ctx sdk.Context) sdk.KVStore

	Set(ctx sdk.Context, key, value []byte)
	Get(ctx sdk.Context, key []byte) []byte
	Has(ctx sdk.Context, key []byte) bool
	Delete(ctx sdk.Context, key []byte)
}

// CommitKVStoreI defines the expected key-value commit store
type CommitKVStoreI interface {
	KVStoreI
	contracttypes.CommitStoreI
}
