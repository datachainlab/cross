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

const OP_NUM uint8 = 2

type OPs []OP

type OP interface {
	cross.OP

	Key() []byte
	Type() uint8
	ApplyTo(sdk.KVStore)
}

type ReadOP interface {
	OP
}

type WriteOP interface {
	OP
	Value() []byte
}

var _ ReadOP = (*Read)(nil)

type Read struct {
	K []byte
}

func (r Read) Key() []byte {
	return r.K
}

func (r Read) Equal(cop cross.OP) bool {
	op := cop.(OP)
	if r.Type() != op.Type() {
		return false
	}
	return bytes.Equal(r.K, op.(ReadOP).Key())
}

func (r Read) Type() uint8 {
	return OP_TYPE_READ
}

func (r Read) ApplyTo(sdk.KVStore) {}

func (r Read) String() string {
	return fmt.Sprintf("Read{%X}", r.K)
}

var _ WriteOP = (*Write)(nil)

type Write struct {
	K []byte
	V []byte
}

func (w Write) Key() []byte {
	return w.K
}

func (w Write) Value() []byte {
	return w.V
}

func (w Write) Type() uint8 {
	return OP_TYPE_WRITE
}

func (w Write) Equal(cop cross.OP) bool {
	op := cop.(OP)
	if w.Type() != op.Type() {
		return false
	}
	return bytes.Equal(w.K, op.(WriteOP).Key()) && bytes.Equal(w.V, op.(WriteOP).Value())
}

func (w Write) ApplyTo(kvs sdk.KVStore) {
	if w.V == nil {
		kvs.Delete(w.K)
	} else {
		kvs.Set(w.K, w.V)
	}
}

func (w Write) String() string {
	return fmt.Sprintf("Write{%X %X}", w.K, w.V)
}

type OPManager struct {
	ops     OPs
	changes map[string]uint64
}

func NewOPManager() *OPManager {
	return &OPManager{changes: make(map[string]uint64)}
}

func (m *OPManager) AddOP(op OP) {
	m.ops = append(m.ops, op)
	if op.Type() == OP_TYPE_WRITE {
		m.changes[string(op.Key())] = uint64(len(m.ops) - 1)
	}
}

func (m *OPManager) GetLastChange(key []byte) (WriteOP, bool) {
	idx, ok := m.changes[string(key)]
	if !ok {
		return nil, false
	}
	return m.ops[idx].(WriteOP), true
}

func (m *OPManager) OPs() OPs {
	return OptimizeOPs(m.ops)
}

func (m *OPManager) COPs() cross.OPs {
	ops := m.OPs()
	cops := make(cross.OPs, len(ops))
	for i, op := range ops {
		cops[i] = op
	}
	return cops
}

type item struct {
	tp  uint8
	idx uint32
}

func OptimizeOPs(ops OPs) OPs {
	m := make(map[string]item)

	for i, op := range ops {
		v, ok := m[string(op.Key())]
		if !ok {
			m[string(op.Key())] = item{op.Type(), uint32(i)}
		} else {
			if op.Type() >= v.tp {
				m[string(op.Key())] = item{op.Type(), uint32(i)}
			}
		}
	}

	optimized := make([]OP, 0, len(m))
	for i, op := range ops {
		if m[string(op.Key())].idx == uint32(i) {
			optimized = append(optimized, op)
		}
	}
	return optimized
}
