package lock

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type lockStore struct {
	locks sdk.KVStore
}

func newLockStore(kvs sdk.KVStore) lockStore {
	return lockStore{locks: kvs}
}

func (ls lockStore) Lock(key []byte) {
	lock := ls.locks.Get(key)
	if lock != nil {
		panic("fatal error")
	}
	ls.locks.Set(key, []byte{1})
}

func (ls lockStore) Unlock(key []byte) {
	ls.locks.Delete(key)
}

func (ls lockStore) IsLocked(key []byte) bool {
	return ls.locks.Get(key) != nil
}
