package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (lo LockOP) Key() []byte {
	return lo.K
}

func (lo LockOP) Value() []byte {
	return lo.V
}

func (lo LockOP) ApplyTo(kvs sdk.KVStore) {
	if lo.V == nil {
		kvs.Delete(lo.K)
	} else {
		kvs.Set(lo.K, lo.V)
	}
}

type LockManager interface {
	AddWrite(key, value []byte) error
	GetUpdatedValue(key []byte) ([]byte, bool)
	LockOPs() LockOPs
}

// NewLockManager returns a LockManager instance
func NewLockManager() LockManager {
	return &lockManager{changes: make(map[string]uint64)}
}

type lockManager struct {
	ops     []LockOP
	changes map[string]uint64
}

func (m *lockManager) AddWrite(k, v []byte) error {
	if len(k) == 0 {
		return errors.New("key cannot be empty")
	}
	if len(v) == 0 {
		return errors.New("value cannot be nil")
	}

	m.ops = append(m.ops, LockOP{k, v})
	m.changes[string(k)] = uint64(len(m.ops) - 1)
	return nil
}

func (m lockManager) GetUpdatedValue(key []byte) ([]byte, bool) {
	idx, ok := m.changes[string(key)]
	if !ok {
		return nil, false
	}
	return m.ops[idx].V, true
}

func (m lockManager) LockOPs() LockOPs {
	items := make(map[string]int)
	for i, op := range m.ops {
		items[string(op.Key())] = i
	}
	ops := make([]LockOP, 0, len(items))
	for i, op := range m.ops {
		if items[string(op.Key())] == i {
			ops = append(ops, op)
		}
	}
	return LockOPs{Ops: ops}
}
