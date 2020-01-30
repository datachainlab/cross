package lock

import (
	"fmt"

	"github.com/bluele/crossccc/x/ibc/crossccc"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Store struct {
	main sdk.KVStore // committed store

	/*
		txID => serialize(OPs)
	*/
	txs sdk.KVStore // serialize and save ops, store is separated by txID

	/*
		{key} => txID
	*/
	locks sdk.KVStore

	opmgr *OPManager
}

func NewStore(kvs sdk.KVStore) *Store {
	main := prefix.NewStore(kvs, []byte{0})
	txs := prefix.NewStore(kvs, []byte{1})
	locks := prefix.NewStore(kvs, []byte{2})
	return &Store{main: main, txs: txs, locks: locks, opmgr: NewOPManager()}
}

func (s *Store) OPs() crossccc.OPs {
	return s.opmgr.COPs()
}

func (s *Store) Precommit(id []byte) error {
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
		s.locks.Set(op.Key(), id)
	}

	return nil
}

func (s *Store) Commit(id []byte) error {
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

func (s *Store) CommitImmediately() error {
	if err := s.apply(s.opmgr.OPs()); err != nil {
		return err
	}
	return nil
}

func (s *Store) apply(ops OPs) error {
	for _, op := range ops {
		op.ApplyTo(s.main)
	}
	return nil
}

func (s *Store) clean(id []byte, ops OPs) error {
	if !s.txs.Has(id) {
		return fmt.Errorf("id '%x' not found", id)
	}
	s.txs.Delete(id)
	for _, op := range ops {
		s.locks.Delete(op.Key())
	}
	return nil
}

func (s *Store) Discard(id []byte) error {
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

func (s *Store) isAvailableKey(require uint8, key []byte) bool {
	v := s.locks.Get(key)
	if len(v) > 0 { // any locks exist
		return false
	}
	return true
}

// Implement KVStore

func (s *Store) Get(key []byte) []byte {
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

func (s *Store) Has(key []byte) bool {
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

func (s *Store) Set(key, value []byte) {
	if !s.isAvailableKey(OP_TYPE_WRITE, key) {
		panic(fmt.Errorf("currently key '%x' is non-available", key))
	}
	s.opmgr.AddOP(Write{key, value})
}

func (s *Store) Delete(key []byte) {
	if !s.isAvailableKey(OP_TYPE_WRITE, key) {
		panic(fmt.Errorf("currently key '%x' is non-available", key))
	}
	s.opmgr.AddOP(Write{key, nil})
}
