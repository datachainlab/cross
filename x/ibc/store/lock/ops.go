package lock

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/ibc/cross"
)

const (
	// This order equals a priority of operation
	OP_TYPE_READ uint8 = iota + 1
	OP_TYPE_WRITE
)

type LockOP interface {
	cross.OP

	Key() []byte
	ApplyTo(sdk.KVStore)
}

type OPManager interface {
	AddRead(key, value []byte)
	AddWrite(key, value []byte)
	GetUpdatedValue(key []byte) ([]byte, bool)
	OPs() cross.OPs
	LockOPs() []LockOP
}

var _ OP = (*ReadOP)(nil)

type ReadOP struct {
	K []byte
}

func (r ReadOP) Key() []byte {
	return r.K
}

func (r ReadOP) Type() uint8 {
	return OP_TYPE_READ
}

func (r ReadOP) Equal(cop cross.OP) bool {
	op := cop.(OP)
	if r.Type() != op.Type() {
		return false
	}
	return bytes.Equal(r.K, op.(ReadOP).Key())
}

func (r ReadOP) String() string {
	return fmt.Sprintf("Read{%X}", r.K)
}

type ReadValueOP struct {
	K []byte
	V []byte
}

func (r ReadValueOP) Key() []byte {
	return r.K
}

func (r ReadValueOP) Value() []byte {
	return r.V
}

func (r ReadValueOP) Type() uint8 {
	return OP_TYPE_READ
}

func (r ReadValueOP) Equal(cop cross.OP) bool {
	op := cop.(OP)
	if r.Type() != op.Type() {
		return false
	}
	return bytes.Equal(r.K, op.(ReadValueOP).Key()) && bytes.Equal(r.V, op.(ReadValueOP).Value())
}

func (r ReadValueOP) String() string {
	return fmt.Sprintf("ReadValue{%X %X}", r.K, r.V)
}

var (
	_ OP     = (*WriteOP)(nil)
	_ LockOP = (*WriteOP)(nil)
)

type WriteOP struct {
	K []byte
	V []byte
}

func (w WriteOP) Key() []byte {
	return w.K
}

func (w WriteOP) Value() []byte {
	return w.V
}

func (w WriteOP) Type() uint8 {
	return OP_TYPE_WRITE
}

func (w WriteOP) Equal(cop cross.OP) bool {
	op := cop.(OP)
	if w.Type() != op.Type() {
		return false
	}
	return bytes.Equal(w.K, op.(WriteOP).Key()) && bytes.Equal(w.V, op.(WriteOP).Value())
}

func (w WriteOP) ApplyTo(kvs sdk.KVStore) {
	if w.V == nil {
		kvs.Delete(w.K)
	} else {
		kvs.Set(w.K, w.V)
	}
}

func (w WriteOP) String() string {
	return fmt.Sprintf("Write{%X %X}", w.K, w.V)
}

type OP interface {
	cross.OP
	Key() []byte
	Type() uint8
}

func GetOPManager(tp cross.StateConstraintType) (OPManager, error) {
	switch tp {
	case cross.ExactMatchStateConstraint:
		return newExactOPManager(), nil
	default:
		return nil, fmt.Errorf("unknown type '%v'", tp)
	}
}

var _ OPManager = (*exactOPManager)(nil)

type exactOPManager struct {
	ops     []OP
	changes map[string]uint64
}

func newExactOPManager() *exactOPManager {
	return &exactOPManager{changes: make(map[string]uint64)}
}

type item struct {
	tp  uint8
	idx uint32
}

func (m *exactOPManager) AddRead(k, v []byte) {
	m.ops = append(m.ops, ReadValueOP{k, v})
}

func (m *exactOPManager) AddWrite(k, v []byte) {
	m.ops = append(m.ops, WriteOP{k, v})
	m.changes[string(k)] = uint64(len(m.ops) - 1)
}

func (m exactOPManager) GetUpdatedValue(key []byte) ([]byte, bool) {
	idx, ok := m.changes[string(key)]
	if !ok {
		return nil, false
	}
	return m.ops[idx].(WriteOP).V, true
}

func (m exactOPManager) LockOPs() []LockOP {
	items := make(map[string]item)
	for i, op := range m.ops {
		if op.Type() != OP_TYPE_WRITE {
			continue
		}
		items[string(op.Key())] = item{op.Type(), uint32(i)}
	}
	ops := make([]LockOP, 0, len(items))
	for i, op := range m.ops {
		if op.Type() != OP_TYPE_WRITE {
			continue
		}
		if items[string(op.Key())].idx == uint32(i) {
			ops = append(ops, op.(WriteOP))
		}
	}
	return ops
}

func (m exactOPManager) OPs() cross.OPs {
	ops := make(cross.OPs, 0, len(m.ops))
	for _, op := range m.ops {
		ops = append(ops, op)
	}
	return ops
}
