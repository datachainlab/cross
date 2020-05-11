package lock

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/ibc/cross"
)

const (
	MainStorePrefix uint8 = iota
	TxStorePrefix
	LockStorePrefix
)

var _ cross.Store = (*Store)(nil)
var _ cross.State = (*Store)(nil)

type Store struct {
	main sdk.KVStore // committed store

	// serialize opeations to bytes and save it on store
	// txID => serialize(OPs)
	txs sdk.KVStore

	lockStore lockStore

	opmgr OPManager
}

func NewStore(kvs sdk.KVStore, tp cross.StateConditionType) Store {
	main := prefix.NewStore(kvs, []byte{MainStorePrefix})
	txs := prefix.NewStore(kvs, []byte{TxStorePrefix})
	locks := prefix.NewStore(kvs, []byte{LockStorePrefix})
	opmgr, err := GetOPManager(tp)
	if err != nil {
		panic(err)
	}
	return Store{main: main, txs: txs, lockStore: newLockStore(locks), opmgr: opmgr}
}

func (s Store) OPs() cross.OPs {
	return s.opmgr.OPs()
}

func (s Store) Precommit(id []byte) error {
	if s.txs.Has(id) {
		return fmt.Errorf("id '%x' already exists", id)
	}
	locks := s.opmgr.LockOPs()
	bz, err := cdc.MarshalBinaryLengthPrefixed(locks)
	if err != nil {
		return err
	}
	s.txs.Set(id, bz)

	for _, op := range locks {
		s.lockStore.Lock(op.Key())
	}

	return nil
}

func (s Store) Commit(id []byte) error {
	bz := s.txs.Get(id)
	if bz == nil {
		return fmt.Errorf("id '%x' not found", id)
	}
	var ops []LockOP
	if err := cdc.UnmarshalBinaryLengthPrefixed(bz, &ops); err != nil {
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
	if err := s.apply(s.opmgr.LockOPs()); err != nil {
		return err
	}
	return nil
}

func (s Store) apply(ops []LockOP) error {
	for _, op := range ops {
		op.ApplyTo(s.main)
	}
	return nil
}

func (s Store) clean(id []byte, ops []LockOP) error {
	if !s.txs.Has(id) {
		return fmt.Errorf("id '%x' not found", id)
	}
	s.txs.Delete(id)
	for _, op := range ops {
		s.lockStore.Unlock(op.Key())
	}
	return nil
}

func (s Store) Discard(id []byte) error {
	b := s.txs.Get(id)
	if b == nil {
		return fmt.Errorf("id '%x' not found", id)
	}
	var ops []LockOP
	if err := cdc.UnmarshalBinaryLengthPrefixed(b, &ops); err != nil {
		return err
	}
	return s.clean(id, ops)
}

// Implement KVStore

func (s Store) Get(key []byte) []byte {
	if s.lockStore.IsLocked(key) {
		panic(fmt.Errorf("currently key '%x' is non-available", key))
	}
	v, ok := s.opmgr.GetUpdatedValue(key)
	s.opmgr.AddRead(key, v)
	if ok {
		return v
	} else {
		return s.main.Get(key)
	}
}

func (s Store) Has(key []byte) bool {
	if s.lockStore.IsLocked(key) {
		panic(fmt.Errorf("currently key '%x' is non-available", key))
	}
	v, ok := s.opmgr.GetUpdatedValue(key)
	s.opmgr.AddRead(key, v)
	if ok {
		return true
	} else {
		return s.main.Has(key)
	}
}

func (s Store) Set(key, value []byte) {
	if s.lockStore.IsLocked(key) {
		panic(fmt.Errorf("currently key '%x' is non-available", key))
	}
	s.opmgr.AddWrite(key, value)
}

func (s Store) Delete(key []byte) {
	if s.lockStore.IsLocked(key) {
		panic(fmt.Errorf("currently key '%x' is non-available", key))
	}
	s.opmgr.AddWrite(key, nil)
}
