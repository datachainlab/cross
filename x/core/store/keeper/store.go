package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	contracttypes "github.com/datachainlab/cross/x/core/contract/types"
	"github.com/datachainlab/cross/x/core/store/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
	"github.com/gogo/protobuf/proto"
)

var _ types.KVStoreI = (*kvStore)(nil)

type kvStore struct {
	storeKey sdk.StoreKey
	prefix   []byte
}

func newKVStore(storeKey sdk.StoreKey) types.KVStoreI {
	return &kvStore{storeKey: storeKey}
}

func (s kvStore) Prefix(prefix []byte) types.KVStoreI {
	p := make([]byte, len(s.prefix)+len(prefix))
	copy(p[0:len(s.prefix)], s.prefix)
	copy(p[len(s.prefix):], prefix)
	s.prefix = p
	return s
}

func (s kvStore) KVStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(s.store(ctx), s.prefix)
}

func (s kvStore) Set(ctx sdk.Context, key, value []byte) {
	s.KVStore(ctx).Set(key, value)
}

func (s kvStore) Get(ctx sdk.Context, key []byte) []byte {
	return s.KVStore(ctx).Get(key)
}

func (s kvStore) Has(ctx sdk.Context, key []byte) bool {
	return s.KVStore(ctx).Has(key)
}

func (s kvStore) Delete(ctx sdk.Context, key []byte) {
	s.KVStore(ctx).Delete(key)
}

func (s kvStore) store(ctx sdk.Context) sdk.KVStore {
	switch storeKey := s.storeKey.(type) {
	case *crosstypes.PrefixStoreKey:
		return prefix.NewStore(ctx.KVStore(storeKey.StoreKey), storeKey.Prefix)
	default:
		return ctx.KVStore(s.storeKey)
	}
}

type CommitKVStore struct {
	storeKey   sdk.StoreKey
	m          codec.Codec
	stateStore types.KVStoreI
	lockStore  LockStore
	txStore    types.KVStoreI
	prefix     []byte
}

var _ types.CommitKVStoreI = (*CommitKVStore)(nil)

func NewStore(m codec.Codec, storeKey sdk.StoreKey) CommitKVStore {
	return CommitKVStore{
		storeKey:   storeKey,
		m:          m,
		stateStore: newKVStore(storeKey).Prefix([]byte{0}),
		lockStore:  newLockStore(newKVStore(storeKey).Prefix([]byte{1})),
		txStore:    newKVStore(storeKey).Prefix([]byte{2}),
	}
}

func (s CommitKVStore) Prefix(prefix []byte) types.KVStoreI {
	s.stateStore = s.stateStore.Prefix(prefix)
	s.lockStore = s.lockStore.Prefix(prefix)
	// for LockManager
	newprefix := make([]byte, len(s.prefix)+len(prefix))
	copy(newprefix[:len(s.prefix)], s.prefix)
	copy(newprefix[len(s.prefix):], prefix)
	s.prefix = newprefix
	return s
}

func (s CommitKVStore) KVStore(ctx sdk.Context) sdk.KVStore {
	panic("not implemented error")
}

func (s CommitKVStore) Set(ctx sdk.Context, key, value []byte) {
	if s.lockStore.IsLocked(ctx, key) {
		panic(fmt.Errorf("currently key '%x' is non-available", key))
	}
	switch contracttypes.CommitModeFromContext(ctx.Context()) {
	case contracttypes.UnspecifiedMode, contracttypes.BasicMode:
		s.stateStore.Set(ctx, key, value)
		return
	case contracttypes.AtomicMode:
		types.LockManagerFromContext(ctx.Context()).AddWrite(s.buildKey(key), value)
		return
	default:
		panic(fmt.Sprintf("unknown mode '%v'", contracttypes.CommitModeFromContext(ctx.Context())))
	}
}

func (s CommitKVStore) Get(ctx sdk.Context, key []byte) []byte {
	if s.lockStore.IsLocked(ctx, key) {
		panic(fmt.Errorf("currently key '%x' is non-available", key))
	}
	switch contracttypes.CommitModeFromContext(ctx.Context()) {
	case contracttypes.UnspecifiedMode, contracttypes.BasicMode:
		return s.stateStore.Get(ctx, key)
	case contracttypes.AtomicMode:
		lkmgr := types.LockManagerFromContext(ctx.Context())
		v, ok := lkmgr.GetUpdatedValue(s.buildKey(key))
		if !ok {
			v = s.stateStore.Get(ctx, key)
		}
		return v
	default:
		panic(fmt.Sprintf("unknown mode '%v'", contracttypes.CommitModeFromContext(ctx.Context())))
	}
}

func (s CommitKVStore) Has(ctx sdk.Context, key []byte) bool {
	if s.lockStore.IsLocked(ctx, key) {
		panic(fmt.Errorf("currently key '%x' is non-available", key))
	}
	switch contracttypes.CommitModeFromContext(ctx.Context()) {
	case contracttypes.UnspecifiedMode, contracttypes.BasicMode:
		return s.stateStore.Has(ctx, key)
	case contracttypes.AtomicMode:
		lkmgr := types.LockManagerFromContext(ctx.Context())
		found := false
		v, ok := lkmgr.GetUpdatedValue(s.buildKey(key))
		if !ok {
			v = s.stateStore.Get(ctx, key)
			if v != nil {
				found = true
			}
		} else {
			found = true
		}
		return found
	default:
		panic(fmt.Sprintf("unknown mode '%v'", contracttypes.CommitModeFromContext(ctx.Context())))
	}
}

func (s CommitKVStore) Delete(ctx sdk.Context, key []byte) {
	if s.lockStore.IsLocked(ctx, key) {
		panic(fmt.Errorf("currently key '%x' is non-available", key))
	}
	switch contracttypes.CommitModeFromContext(ctx.Context()) {
	case contracttypes.UnspecifiedMode, contracttypes.BasicMode:
		s.stateStore.Delete(ctx, key)
		return
	case contracttypes.AtomicMode:
		types.LockManagerFromContext(ctx.Context()).AddWrite(s.buildKey(key), nil)
	default:
		panic(fmt.Sprintf("unknown mode '%v'", contracttypes.CommitModeFromContext(ctx.Context())))
	}
}

func (s CommitKVStore) Precommit(ctx sdk.Context, id []byte) error {
	if s.txStore.Has(ctx, id) {
		return fmt.Errorf("id '%x' already exists", id)
	}
	lks := types.LockManagerFromContext(ctx.Context()).LockOPs()
	bz, err := proto.Marshal(&lks)
	if err != nil {
		return err
	}
	s.txStore.Set(ctx, id, bz)
	for _, lk := range lks.Ops {
		s.lockStore.Lock(ctx, lk.Key())
	}
	return nil
}

func (s CommitKVStore) Abort(ctx sdk.Context, id []byte) error {
	bz := s.txStore.Get(ctx, id)
	if bz == nil {
		// NOTE: unknown id may be indicates the aborted transaction
		return nil
	}
	var lks types.LockOPs
	if err := proto.Unmarshal(bz, &lks); err != nil {
		return err
	}
	s.clean(ctx, id, lks.Ops)
	return nil
}

func (s CommitKVStore) Commit(ctx sdk.Context, id []byte) error {
	bz := s.txStore.Get(ctx, id)
	if bz == nil {
		return fmt.Errorf("id '%x' not found", id)
	}
	var lks types.LockOPs
	if err := proto.Unmarshal(bz, &lks); err != nil {
		return err
	}
	s.apply(ctx, lks.Ops)
	s.clean(ctx, id, lks.Ops)
	return nil
}

func (s CommitKVStore) CommitImmediately(ctx sdk.Context) {
	lks := types.LockManagerFromContext(ctx.Context()).LockOPs()
	s.apply(ctx, lks.Ops)
}

func (s CommitKVStore) apply(ctx sdk.Context, ops []types.LockOP) {
	for _, op := range ops {
		op.ApplyTo(s.stateStore.KVStore(ctx))
	}
}

func (s CommitKVStore) clean(ctx sdk.Context, id []byte, ops []types.LockOP) {
	if !s.txStore.Has(ctx, id) {
		panic(fmt.Errorf("id '%x' not found", id))
	}
	s.txStore.Delete(ctx, id)
	for _, op := range ops {
		s.lockStore.Unlock(ctx, op.Key())
	}
}

func (s CommitKVStore) buildKey(key []byte) []byte {
	newkey := make([]byte, len(s.prefix)+len(key))
	copy(newkey[:len(s.prefix)], s.prefix)
	copy(newkey[len(s.prefix):], key)
	return newkey
}
