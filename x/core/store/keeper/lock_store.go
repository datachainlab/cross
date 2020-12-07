package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/core/store/types"
)

type LockStore interface {
	Lock(ctx sdk.Context, key []byte)
	Unlock(ctx sdk.Context, key []byte)
	IsLocked(ctx sdk.Context, key []byte) bool

	Prefix(prefix []byte) LockStore
}

type lockStore struct {
	store types.KVStoreI
}

var _ LockStore = (*lockStore)(nil)

func newLockStore(store types.KVStoreI) lockStore {
	return lockStore{store: store}
}

func (s lockStore) Lock(ctx sdk.Context, key []byte) {
	lock := s.store.Get(ctx, key)
	if lock != nil {
		panic("fatal error")
	}
	s.store.Set(ctx, key, []byte{1})
}

func (s lockStore) Unlock(ctx sdk.Context, key []byte) {
	s.store.Delete(ctx, key)
}

func (s lockStore) IsLocked(ctx sdk.Context, key []byte) bool {
	return s.store.Get(ctx, key) != nil
}

func (s lockStore) Prefix(prefix []byte) LockStore {
	s.store = s.store.Prefix(prefix)
	return s
}
