package store

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/core/types"
	"github.com/gogo/protobuf/proto"
)

var _ types.Store = (*kvStore)(nil)

type kvStore struct {
	storeKey sdk.StoreKey
	prefix   []byte
}

func newKVStore(storeKey sdk.StoreKey) types.Store {
	return &kvStore{storeKey: storeKey}
}

func (s kvStore) Prefix(prefix []byte) types.Store {
	p := make([]byte, len(s.prefix)+len(prefix))
	copy(p[0:len(s.prefix)], s.prefix)
	copy(p[len(s.prefix):], prefix)
	s.prefix = p
	return s
}

func (s kvStore) KVStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(s.storeKey), s.prefix)
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

type Store struct {
	storeKey   sdk.StoreKey
	m          codec.Marshaler
	stateStore types.Store
	lockStore  LockStore
	txStore    types.Store
}

var _ types.Store = (*Store)(nil)
var _ types.CommitStore = (*Store)(nil)

func NewStore(m codec.Marshaler, storeKey sdk.StoreKey) Store {
	return Store{
		storeKey:   storeKey,
		m:          m,
		stateStore: newKVStore(storeKey).Prefix([]byte{0}),
		lockStore:  newLockStore(newKVStore(storeKey).Prefix([]byte{1})),
		txStore:    newKVStore(storeKey).Prefix([]byte{2}),
	}
}

func (s Store) Prefix(prefix []byte) types.Store {
	s.stateStore = s.stateStore.Prefix(prefix)
	s.lockStore = s.lockStore.Prefix(prefix)
	return s
}

func (s Store) KVStore(ctx sdk.Context) sdk.KVStore {
	panic("not implemented error")
}

func (s Store) Set(ctx sdk.Context, key, value []byte) {
	if s.lockStore.IsLocked(ctx, key) {
		panic(fmt.Errorf("currently key '%x' is non-available", key))
	}
	switch types.CommitModeFromContext(ctx.Context()) {
	case types.UnspecifiedMode, types.BasicMode:
		s.stateStore.Set(ctx, key, value)
		return
	case types.AtomicMode:
		opManagerFromContext(ctx.Context()).AddWrite(key, value)
		return
	default:
		panic(fmt.Sprintf("unknown mode '%v'", types.CommitModeFromContext(ctx.Context())))
	}
}

func (s Store) Get(ctx sdk.Context, key []byte) []byte {
	if s.lockStore.IsLocked(ctx, key) {
		panic(fmt.Errorf("currently key '%x' is non-available", key))
	}
	switch types.CommitModeFromContext(ctx.Context()) {
	case types.UnspecifiedMode, types.BasicMode:
		return s.stateStore.Get(ctx, key)
	case types.AtomicMode:
		opmgr := opManagerFromContext(ctx.Context())
		v, ok := opmgr.GetUpdatedValue(key)
		opmgr.AddRead(key, v)
		if ok {
			return v
		} else {
			return s.stateStore.Get(ctx, key)
		}
	default:
		panic(fmt.Sprintf("unknown mode '%v'", types.CommitModeFromContext(ctx.Context())))
	}
}

func (s Store) Has(ctx sdk.Context, key []byte) bool {
	if s.lockStore.IsLocked(ctx, key) {
		panic(fmt.Errorf("currently key '%x' is non-available", key))
	}
	switch types.CommitModeFromContext(ctx.Context()) {
	case types.UnspecifiedMode, types.BasicMode:
		return s.stateStore.Has(ctx, key)
	case types.AtomicMode:
		opmgr := opManagerFromContext(ctx.Context())
		v, ok := opmgr.GetUpdatedValue(key)
		opmgr.AddRead(key, v)
		if ok {
			return true
		} else {
			return s.stateStore.Has(ctx, key)
		}
	default:
		panic(fmt.Sprintf("unknown mode '%v'", types.CommitModeFromContext(ctx.Context())))
	}
}

func (s Store) Delete(ctx sdk.Context, key []byte) {
	if s.lockStore.IsLocked(ctx, key) {
		panic(fmt.Errorf("currently key '%x' is non-available", key))
	}
	switch types.CommitModeFromContext(ctx.Context()) {
	case types.UnspecifiedMode, types.BasicMode:
		s.stateStore.Delete(ctx, key)
		return
	case types.AtomicMode:
		opManagerFromContext(ctx.Context()).AddWrite(key, nil)
	default:
		panic(fmt.Sprintf("unknown mode '%v'", types.CommitModeFromContext(ctx.Context())))
	}
}

func (s Store) Precommit(ctx sdk.Context, id []byte) error {
	if s.txStore.Has(ctx, id) {
		return fmt.Errorf("id '%x' already exists", id)
	}
	lks := opManagerFromContext(ctx.Context()).LockOPs()
	ops, err := convertLockOPsToOPs(lks)
	if err != nil {
		return err
	}
	bz, err := proto.Marshal(ops)
	if err != nil {
		return err
	}
	s.txStore.Set(ctx, id, bz)
	for _, lk := range lks {
		s.lockStore.Lock(ctx, lk.Key())
	}
	return nil
}

func (s Store) Abort(ctx sdk.Context, id []byte) error {
	bz := s.txStore.Get(ctx, id)
	if bz == nil {
		return fmt.Errorf("id '%x' not found", id)
	}
	var ops types.OPs
	if err := proto.Unmarshal(bz, &ops); err != nil {
		return err
	}
	lks, err := convertOPsToLockOPs(s.m, ops)
	if err != nil {
		return err
	}
	s.clean(ctx, id, lks)
	return nil
}

func (s Store) Commit(ctx sdk.Context, id []byte) error {
	bz := s.txStore.Get(ctx, id)
	if bz == nil {
		return fmt.Errorf("id '%x' not found", id)
	}
	var ops types.OPs
	if err := proto.Unmarshal(bz, &ops); err != nil {
		return err
	}
	lks, err := convertOPsToLockOPs(s.m, ops)
	if err != nil {
		return err
	}
	s.apply(ctx, lks)
	s.clean(ctx, id, lks)
	return nil
}

func (s Store) CommitImmediately(ctx sdk.Context) {
	lks := opManagerFromContext(ctx.Context()).LockOPs()
	s.apply(ctx, lks)
}

func (s Store) apply(ctx sdk.Context, ops []LockOP) {
	for _, op := range ops {
		op.ApplyTo(s.stateStore.KVStore(ctx))
	}
}

func (s Store) clean(ctx sdk.Context, id []byte, ops []LockOP) {
	if !s.txStore.Has(ctx, id) {
		panic(fmt.Errorf("id '%x' not found", id))
	}
	s.txStore.Delete(ctx, id)
	for _, op := range ops {
		s.lockStore.Unlock(ctx, op.Key())
	}
}
