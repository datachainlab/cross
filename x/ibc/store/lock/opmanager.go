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
	case cross.PreStateConstraint:
		return newReadOPManager(), nil
	case cross.PostStateConstraint:
		return newWriteOPManager(), nil
	default:
		return nil, fmt.Errorf("unknown type '%v'", tp)
	}
}

type commonOPManager struct {
	ops     []OP
	changes map[string]uint64
}

func newCommonOPManager() commonOPManager {
	return commonOPManager{changes: make(map[string]uint64)}
}

type item struct {
	tp  uint8
	idx uint32
}

func (m *commonOPManager) AddWrite(k, v []byte) {
	m.ops = append(m.ops, WriteOP{k, v})
	m.changes[string(k)] = uint64(len(m.ops) - 1)
}

func (m commonOPManager) GetUpdatedValue(key []byte) ([]byte, bool) {
	idx, ok := m.changes[string(key)]
	if !ok {
		return nil, false
	}
	return m.ops[idx].(WriteOP).V, true
}

func (m commonOPManager) LockOPs() []LockOP {
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

var _ OPManager = (*readWriteOPManager)(nil)

type readWriteOPManager struct {
	commonOPManager
}

func newReadWriteOPManager() *readWriteOPManager {
	return &readWriteOPManager{
		commonOPManager: newCommonOPManager(),
	}
}

func (m *readWriteOPManager) AddRead(k, v []byte) {
	m.ops = append(m.ops, ReadOP{k, v})
}

func (m readWriteOPManager) OPs() cross.OPs {
	ops := make(cross.OPs, 0, len(m.ops))
	for _, op := range m.ops {
		ops = append(ops, op)
	}
	return ops
}

type readOPManager struct {
	commonOPManager
}

var _ OPManager = (*readOPManager)(nil)

func newReadOPManager() *readOPManager {
	return &readOPManager{
		commonOPManager: newCommonOPManager(),
	}
}

func (m *readOPManager) AddRead(k, v []byte) {
	m.ops = append(m.ops, ReadOP{k, v})
}

func (m readOPManager) OPs() cross.OPs {
	ops := make(cross.OPs, 0, len(m.ops))
	for _, op := range m.ops {
		if op.Type() == OP_TYPE_READ {
			ops = append(ops, op)
		} else {
			continue
		}
	}
	return ops
}

type writeOPManager struct {
	commonOPManager
}

var _ OPManager = (*writeOPManager)(nil)

func newWriteOPManager() *writeOPManager {
	return &writeOPManager{
		commonOPManager: newCommonOPManager(),
	}
}

func (m *writeOPManager) AddRead(k, v []byte) {}

func (m writeOPManager) OPs() cross.OPs {
	ops := make(cross.OPs, 0, len(m.ops))
	for _, op := range m.ops {
		if op.Type() == OP_TYPE_WRITE {
			ops = append(ops, op)
		} else {
			continue
		}
	}
	return ops
}
