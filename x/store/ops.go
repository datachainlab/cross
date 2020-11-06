package store

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/core/types"
)

// This order equals a priority of operation
const (
	OpTypeRead uint8 = iota + 1
	OpTypeWrite
)

type LockOP interface {
	types.OP

	Key() []byte
	ApplyTo(sdk.KVStore)
}

type OP interface {
	types.OP

	Key() []byte
	Type() uint8
}

var _ OP = (*ReadOP)(nil)

func (r ReadOP) Key() []byte {
	return r.K
}

func (r ReadOP) Value() []byte {
	return r.V
}

func (r ReadOP) Type() uint8 {
	return OpTypeRead
}

func (r ReadOP) Equal(cop types.OP) bool {
	op := cop.(OP)
	if r.Type() != op.Type() {
		return false
	}
	return bytes.Equal(r.K, op.(*ReadOP).Key()) && bytes.Equal(r.V, op.(*ReadOP).Value())
}

var (
	_ OP     = (*WriteOP)(nil)
	_ LockOP = (*WriteOP)(nil)
)

func (w WriteOP) Key() []byte {
	return w.K
}

func (w WriteOP) Value() []byte {
	return w.V
}

func (w WriteOP) Type() uint8 {
	return OpTypeWrite
}

func (w WriteOP) Equal(cop types.OP) bool {
	op := cop.(OP)
	if w.Type() != op.Type() {
		return false
	}
	return bytes.Equal(w.K, op.(*WriteOP).Key()) && bytes.Equal(w.V, op.(*WriteOP).Value())
}

func (w WriteOP) ApplyTo(kvs sdk.KVStore) {
	if w.V == nil {
		kvs.Delete(w.K)
	} else {
		kvs.Set(w.K, w.V)
	}
}

type OPManager interface {
	AddRead(key, value []byte)
	AddWrite(key, value []byte)
	GetUpdatedValue(key []byte) ([]byte, bool)
	OPs() types.OPs
	LockOPs() []LockOP
}

func getOPManager(tp types.StateConstraintType) (OPManager, error) {
	switch tp {
	case types.ExactMatchStateConstraint:
		return newReadWriteOPManager(), nil
	case types.PreStateConstraint:
		return newReadOPManager(), nil
	case types.PostStateConstraint:
		return newWriteOPManager(), nil
	case types.NoStateConstraint:
		return newNoOPManager(), nil
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
	m.ops = append(m.ops, &WriteOP{k, v})
	m.changes[string(k)] = uint64(len(m.ops) - 1)
}

func (m commonOPManager) GetUpdatedValue(key []byte) ([]byte, bool) {
	idx, ok := m.changes[string(key)]
	if !ok {
		return nil, false
	}
	return m.ops[idx].(*WriteOP).V, true
}

func (m commonOPManager) LockOPs() []LockOP {
	items := make(map[string]item)
	for i, op := range m.ops {
		if op.Type() != OpTypeWrite {
			continue
		}
		items[string(op.Key())] = item{op.Type(), uint32(i)}
	}
	ops := make([]LockOP, 0, len(items))
	for i, op := range m.ops {
		if op.Type() != OpTypeWrite {
			continue
		}
		if items[string(op.Key())].idx == uint32(i) {
			ops = append(ops, op.(*WriteOP))
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
	m.ops = append(m.ops, &ReadOP{k, v})
}

func (m readWriteOPManager) OPs() types.OPs {
	ops, err := convertOPItemsToOPs(m.ops)
	if err != nil {
		panic(err)
	}
	return *ops
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
	m.ops = append(m.ops, &ReadOP{k, v})
}

func (m readOPManager) OPs() types.OPs {
	items := make([]OP, 0, len(m.ops))
	for _, op := range m.ops {
		if op.Type() == OpTypeRead {
			items = append(items, op)
		} else {
			continue
		}
	}
	ops, err := convertOPItemsToOPs(items)
	if err != nil {
		panic(err)
	}
	return *ops
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

func (m writeOPManager) OPs() types.OPs {
	items := make([]OP, 0, len(m.ops))
	for _, op := range m.ops {
		if op.Type() == OpTypeWrite {
			items = append(items, op)
		} else {
			continue
		}
	}
	ops, err := convertOPItemsToOPs(items)
	if err != nil {
		panic(err)
	}
	return *ops
}

type noOPManager struct {
	commonOPManager
}

var _ OPManager = (*noOPManager)(nil)

func newNoOPManager() *noOPManager {
	return &noOPManager{
		commonOPManager: newCommonOPManager(),
	}
}

func (m *noOPManager) AddRead(k, v []byte) {}

func (m noOPManager) OPs() types.OPs {
	return types.OPs{}
}

func convertOPsToLockOPs(m codec.Marshaler, ops types.OPs) ([]LockOP, error) {
	var lks []LockOP
	for _, op := range ops.Items {
		var lk LockOP
		if err := m.UnpackAny(&op, &lk); err != nil {
			return nil, err
		}
		lks = append(lks, lk)
	}
	return lks, nil
}

func convertLockOPsToOPs(lks []LockOP) (*types.OPs, error) {
	var ops types.OPs
	for _, lk := range lks {
		var any codectypes.Any
		if err := any.Pack(lk); err != nil {
			return nil, err
		}
		ops.Items = append(ops.Items, any)
	}
	return &ops, nil
}

func convertOPItemsToOPs(items []OP) (*types.OPs, error) {
	var ops types.OPs
	for _, it := range items {
		var any codectypes.Any
		if err := any.Pack(it); err != nil {
			return nil, err
		}
		ops.Items = append(ops.Items, any)
	}
	return &ops, nil
}
