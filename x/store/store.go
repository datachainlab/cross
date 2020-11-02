package store

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/core/types"
)

var _ types.Store = (*Store)(nil)

type Store struct {
	storeKey sdk.StoreKey
	prefix   []byte
}

func NewStore(storeKey sdk.StoreKey) types.Store {
	return &Store{storeKey: storeKey}
}

func (s Store) Prefix(prefix []byte) types.Store {
	p := make([]byte, len(s.prefix)+len(prefix))
	copy(p[0:len(s.prefix)], s.prefix)
	copy(p[len(s.prefix):], prefix)
	s.prefix = p
	return s
}

func (s Store) KVStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(s.storeKey), s.prefix)
}

func (s Store) Set(ctx sdk.Context, key, value []byte) {
	s.KVStore(ctx).Set(key, value)
}

func (s Store) Get(ctx sdk.Context, key []byte) []byte {
	return s.KVStore(ctx).Get(key)
}

func (s Store) Has(ctx sdk.Context, key []byte) bool {
	return s.KVStore(ctx).Has(key)
}

func (s Store) Delete(ctx sdk.Context, key []byte) {
	s.KVStore(ctx).Delete(key)
}

type CommitStore struct {
	storeKey   sdk.StoreKey
	stateStore types.Store
	lockStore  LockStore
}

var _ types.CommitStore = (*CommitStore)(nil)
var _ types.Store = (*CommitStore)(nil)

func NewCommitStore(storeKey sdk.StoreKey) CommitStore {
	return CommitStore{
		storeKey:   storeKey,
		stateStore: NewStore(storeKey).Prefix([]byte{0}),
		lockStore:  newLockStore(NewStore(storeKey).Prefix([]byte{1})),
	}
}

func (s CommitStore) Prefix(prefix []byte) types.Store {
	s.stateStore = s.stateStore.Prefix(prefix)
	s.lockStore = s.lockStore.Prefix(prefix)
	return s
}

func (s CommitStore) KVStore(ctx sdk.Context) sdk.KVStore {
	panic("not implemented error")
}

func (s CommitStore) Set(ctx sdk.Context, key, value []byte) {
	if s.lockStore.IsLocked(ctx, key) {
		panic(fmt.Errorf("currently key '%x' is non-available", key))
	}
	switch types.ModeFromContext(ctx.Context()) {
	case types.BasicMode:
		s.stateStore.Set(ctx, key, value)
		return
	case types.AtomicMode:
		OPManagerFromContext(ctx.Context()).AddWrite(key, value)
		return
	default:
		panic(fmt.Sprintf("unknown mode '%v'", types.ModeFromContext(ctx.Context())))
	}
}

func (s CommitStore) Get(ctx sdk.Context, key []byte) []byte {
	if s.lockStore.IsLocked(ctx, key) {
		panic(fmt.Errorf("currently key '%x' is non-available", key))
	}
	switch types.ModeFromContext(ctx.Context()) {
	case types.BasicMode:
		return s.stateStore.Get(ctx, key)
	case types.AtomicMode:
		opmgr := OPManagerFromContext(ctx.Context())
		v, ok := opmgr.GetUpdatedValue(key)
		opmgr.AddRead(key, v)
		if ok {
			return v
		} else {
			return s.stateStore.Get(ctx, key)
		}
	default:
		panic(fmt.Sprintf("unknown mode '%v'", types.ModeFromContext(ctx.Context())))
	}
}

func (s CommitStore) Has(ctx sdk.Context, key []byte) bool {
	if s.lockStore.IsLocked(ctx, key) {
		panic(fmt.Errorf("currently key '%x' is non-available", key))
	}
	switch types.ModeFromContext(ctx.Context()) {
	case types.BasicMode:
		return s.stateStore.Has(ctx, key)
	case types.AtomicMode:
		opmgr := OPManagerFromContext(ctx.Context())
		v, ok := opmgr.GetUpdatedValue(key)
		opmgr.AddRead(key, v)
		if ok {
			return true
		} else {
			return s.stateStore.Has(ctx, key)
		}
	default:
		panic(fmt.Sprintf("unknown mode '%v'", types.ModeFromContext(ctx.Context())))
	}
}

func (s CommitStore) Delete(ctx sdk.Context, key []byte) {
	if s.lockStore.IsLocked(ctx, key) {
		panic(fmt.Errorf("currently key '%x' is non-available", key))
	}
	switch types.ModeFromContext(ctx.Context()) {
	case types.BasicMode:
		s.stateStore.Delete(ctx, key)
		return
	case types.AtomicMode:
		OPManagerFromContext(ctx.Context()).AddWrite(key, nil)
	default:
		panic(fmt.Sprintf("unknown mode '%v'", types.ModeFromContext(ctx.Context())))
	}
}

func (s CommitStore) Precommit(ctx sdk.Context, id []byte) error {
	return nil
}

func (s CommitStore) Abort(ctx sdk.Context, id []byte) error {
	return nil
}

func (s CommitStore) Commit(ctx sdk.Context, id []byte) error {
	return nil
}

func (s CommitStore) CommitImmediately(ctx sdk.Context) error {
	return nil
}
