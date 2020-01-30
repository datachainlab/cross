package lock

import (
	"bytes"

	"github.com/bluele/crossccc/x/ibc/crossccc"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	OP_TYPE_READ uint8 = iota + 1
	OP_TYPE_WRITE
)

const OP_NUM uint8 = 2

type OPs []OP

type OP interface {
	crossccc.OP

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

type Read struct {
	K []byte
}

func (r Read) Key() []byte {
	return r.K
}

func (r Read) Equal(cop crossccc.OP) bool {
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

func (w Write) Equal(cop crossccc.OP) bool {
	op := cop.(OP)
	if w.Type() != op.Type() {
		return false
	}
	return bytes.Equal(w.K, op.(WriteOP).Key()) && bytes.Equal(w.V, op.(WriteOP).Value())
}

func (w Write) ApplyTo(kvs sdk.KVStore) {
	kvs.Set(w.K, w.V)
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
	return m.ops
}

func (m *OPManager) COPs() crossccc.OPs {
	ops := make(crossccc.OPs, len(m.ops))
	for i, op := range m.ops {
		ops[i] = op
	}
	return ops
}
