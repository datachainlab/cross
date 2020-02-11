package lock

import (
	"fmt"

	"github.com/bluele/crossccc/x/ibc/crossccc"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	MainStorePrefix uint8 = iota
	TxStorePrefix
	LockStorePrefix
)

var _ crossccc.Store = (*Store)(nil)
var _ crossccc.State = (*Store)(nil)

type Store struct {
	main sdk.KVStore // committed store

	// serialize opeations to bytes and save it on store
	// txID => serialize(OPs)
	txs sdk.KVStore

	lockStore lockStore

	opmgr *OPManager
}

func NewStore(kvs sdk.KVStore) Store {
	main := prefix.NewStore(kvs, []byte{MainStorePrefix})
	txs := prefix.NewStore(kvs, []byte{TxStorePrefix})
	locks := prefix.NewStore(kvs, []byte{LockStorePrefix})

	return Store{main: main, txs: txs, lockStore: newLockStore(locks), opmgr: NewOPManager()}
}

func (s Store) OPs() crossccc.OPs {
	return s.opmgr.COPs()
}

func (s Store) Precommit(id []byte) error {
	if s.txs.Has(id) {
		return fmt.Errorf("id '%x' already exists", id)
	}
	ops := s.opmgr.OPs()
	b, err := cdc.MarshalBinaryLengthPrefixed(ops)
	if err != nil {
		return err
	}
	s.txs.Set(id, b)

	for _, op := range ops {
		s.lockStore.Lock(convertTxOPTypeToLockType(op.Type()), op.Key())
	}

	return nil
}

func (s Store) Commit(id []byte) error {
	b := s.txs.Get(id)
	if b == nil {
		return fmt.Errorf("id '%x' not found", id)
	}
	var ops OPs
	if err := cdc.UnmarshalBinaryLengthPrefixed(b, &ops); err != nil {
		return err
	}
	if err := s.apply(ops); err != nil {
		return err
	}
	if err := s.clean(id, ops); err != nil {
		return err
	}
	return nil
}

func (s Store) CommitImmediately() error {
	if err := s.apply(s.opmgr.OPs()); err != nil {
		return err
	}
	return nil
}

func (s Store) apply(ops OPs) error {
	for _, op := range ops {
		op.ApplyTo(s.main)
	}
	return nil
}

func (s Store) clean(id []byte, ops OPs) error {
	if !s.txs.Has(id) {
		return fmt.Errorf("id '%x' not found", id)
	}
	s.txs.Delete(id)
	for _, op := range ops {
		s.lockStore.Unlock(convertTxOPTypeToLockType(op.Type()), op.Key())
	}
	return nil
}

func (s Store) Discard(id []byte) error {
	b := s.txs.Get(id)
	if b == nil {
		return fmt.Errorf("id '%x' not found", id)
	}
	var ops OPs
	if err := cdc.UnmarshalBinaryLengthPrefixed(b, &ops); err != nil {
		return err
	}
	return s.clean(id, ops)
}

func (s Store) isAvailableKey(require uint8, key []byte) bool {
	if tp, locked := s.lockStore.HasAnyLocked(key); locked {
		return !IsConflictLock(require, tp)
	} else {
		return true
	}
}

// Implement KVStore

func (s Store) Get(key []byte) []byte {
	if !s.isAvailableKey(OP_TYPE_READ, key) {
		panic(fmt.Errorf("currently key '%x' is non-available", key))
	}
	op, ok := s.opmgr.GetLastChange(key)
	s.opmgr.AddOP(Read{key})
	if ok {
		return op.Value()
	} else {
		return s.main.Get(key)
	}
}

func (s Store) Has(key []byte) bool {
	if !s.isAvailableKey(OP_TYPE_READ, key) {
		panic(fmt.Errorf("currently key '%x' is non-available", key))
	}
	_, ok := s.opmgr.GetLastChange(key)
	s.opmgr.AddOP(Read{key})
	if ok {
		return true
	} else {
		return s.main.Has(key)
	}
}

func (s Store) Set(key, value []byte) {
	if !s.isAvailableKey(OP_TYPE_WRITE, key) {
		panic(fmt.Errorf("currently key '%x' is non-available", key))
	}
	s.opmgr.AddOP(Write{key, value})
}

func (s Store) Delete(key []byte) {
	if !s.isAvailableKey(OP_TYPE_WRITE, key) {
		panic(fmt.Errorf("currently key '%x' is non-available", key))
	}
	s.opmgr.AddOP(Write{key, nil})
}

func IsConflictLock(require, current uint8) bool {
	if require == 0 || current == 0 {
		panic("invalid op type")
	}

	switch current {
	case LOCK_TYPE_READ:
		return require != LOCK_TYPE_READ
	case LOCK_TYPE_WRITE:
		return true
	}
	panic("unreachable here")
}

func convertTxOPTypeToLockType(tp uint8) uint8 {
	switch tp {
	case OP_TYPE_READ:
		return LOCK_TYPE_READ
	case OP_TYPE_WRITE:
		return LOCK_TYPE_WRITE
	default:
		panic("unexpected type")
	}
}
