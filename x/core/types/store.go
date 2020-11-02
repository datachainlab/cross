package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CommitMode = uint8

const (
	BasicMode CommitMode = iota + 1
	AtomicMode
)

type Store interface {
	Prefix(prefix []byte) Store
	KVStore(ctx sdk.Context) sdk.KVStore

	Set(ctx sdk.Context, key, value []byte)
	Get(ctx sdk.Context, key []byte) []byte
	Has(ctx sdk.Context, key []byte) bool
	Delete(ctx sdk.Context, key []byte)
}

type CommitStore interface {
	Store

	Precommit(ctx sdk.Context, id []byte) error
	Abort(ctx sdk.Context, id []byte) error
	Commit(ctx sdk.Context, id []byte) error
	CommitImmediately(ctx sdk.Context)
}
