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

type OP interface {
	cross.OP

	Key() []byte
	Type() uint8
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
