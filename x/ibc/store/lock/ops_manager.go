package lock

import (
	"fmt"

	"github.com/datachainlab/cross/x/ibc/cross"
)

type OPManager interface {
	AddRead(key, value []byte)
	AddWrite(key, value []byte)
	GetUpdatedValue(key []byte) ([]byte, bool)
	OPs() cross.OPs
	LockOPs() []LockOP
}

func GetOPManager(tp cross.StateConstraintType) (OPManager, error) {
	switch tp {
	case cross.ExactMatchStateConstraint:
		return newReadWriteOPManager(), nil
	default:
		return nil, fmt.Errorf("unknown type '%v'", tp)
	}
}

var _ OPManager = (*readWriteOPManager)(nil)

type readWriteOPManager struct {
	ops     []OP
	changes map[string]uint64
}

func newReadWriteOPManager() *readWriteOPManager {
	return &readWriteOPManager{changes: make(map[string]uint64)}
}

type item struct {
	tp  uint8
	idx uint32
}

func (m *readWriteOPManager) AddRead(k, v []byte) {
	m.ops = append(m.ops, ReadValueOP{k, v})
}

func (m *readWriteOPManager) AddWrite(k, v []byte) {
	m.ops = append(m.ops, WriteOP{k, v})
	m.changes[string(k)] = uint64(len(m.ops) - 1)
}

func (m readWriteOPManager) GetUpdatedValue(key []byte) ([]byte, bool) {
	idx, ok := m.changes[string(key)]
	if !ok {
		return nil, false
	}
	return m.ops[idx].(WriteOP).V, true
}

func (m readWriteOPManager) LockOPs() []LockOP {
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

func (m readWriteOPManager) OPs() cross.OPs {
	ops := make(cross.OPs, 0, len(m.ops))
	for _, op := range m.ops {
		ops = append(ops, op)
	}
	return ops
}
