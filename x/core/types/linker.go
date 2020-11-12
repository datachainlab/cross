package types

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

// Linker resolves links that each ContractTransaction has.
type Linker struct {
	objects map[TxIndex]lazyObject
}

// MakeLinker returns Linker
func MakeLinker(m codec.Marshaler, txs []ContractTransaction) (*Linker, error) {
	lkr := Linker{objects: make(map[TxIndex]lazyObject, len(txs))}
	for i, tx := range txs {
		idx := TxIndex(i)
		lkr.objects[idx] = makeLazyObject(func() returnObject {
			if tx.ReturnValue.IsNil() {
				return returnObject{err: errors.New("On cross-chain call, each contractTransaction must be specified a return-value")}
			}
			chainID, err := tx.GetChainID(m)
			if err != nil {
				return returnObject{err: err}
			}
			obj := MakeConstantValueObject(chainID, MakeObjectKey(tx.CallInfo, tx.Signers), tx.ReturnValue.Value)
			return returnObject{obj: &obj}
		})
	}
	return &lkr, nil
}

// Resolve resolves given links and returns resolved Object
func (lkr Linker) Resolve(lks []Link) ([]Object, error) {
	var objects []Object
	for _, lk := range lks {
		idx := lk.GetSrcIndex()
		lzObj, ok := lkr.objects[idx]
		if !ok {
			return nil, fmt.Errorf("idx '%v' not found", idx)
		}
		ret := lzObj()
		if ret.err != nil {
			return nil, ret.err
		}
		objects = append(objects, ret.obj)
	}
	return objects, nil
}

type returnObject struct {
	obj Object
	err error
}

type lazyObject func() returnObject

func makeLazyObject(f func() returnObject) lazyObject {
	var v returnObject
	var once sync.Once
	return func() returnObject {
		once.Do(func() {
			v = f()
			f = nil // so that f can now be GC'ed
		})
		return v
	}
}

// MakeObjectKey returns a key that can be used to identify a contract call
func MakeObjectKey(callInfo ContractCallInfo, signers []Account) []byte {
	h := tmhash.New()
	h.Write(callInfo)
	for _, signer := range signers {
		bz, err := proto.Marshal(&signer)
		if err != nil {
			panic(err)
		}
		h.Write(bz)
	}
	return h.Sum(nil)
}

// ObjectType is a type of Object
type ObjectType = uint8

const (
	// ObjectTypeConstantValue is ObjectType indicates a constant value
	ObjectTypeConstantValue ObjectType = iota + 1
)

// Object wraps an actual value
type Object interface {
	proto.Message
	Type() ObjectType
	Key() []byte
	GetChainID(m codec.Marshaler) ChainID
	Evaluate([]byte) ([]byte, error)
}

var _ Object = (*ConstantValueObject)(nil)

// MakeConstantValueObject returns ConstantValueObject
func MakeConstantValueObject(chainID ChainID, key []byte, value []byte) ConstantValueObject {
	anyChainID, err := PackChainID(chainID)
	if err != nil {
		panic(err)
	}
	return ConstantValueObject{
		ChainId: *anyChainID,
		K:       key,
		V:       value,
	}
}

// Type implements Object.Type
func (l ConstantValueObject) Type() ObjectType {
	return ObjectTypeConstantValue
}

// GetChainID implements Object.GetChainID
func (l ConstantValueObject) GetChainID(m codec.Marshaler) ChainID {
	chainID, err := UnpackChainID(m, l.ChainId)
	if err != nil {
		panic(err)
	}
	return chainID
}

// Key implements Object.Key
func (l ConstantValueObject) Key() []byte {
	return l.K
}

// Evaluate returns a constant value
func (l ConstantValueObject) Evaluate(bz []byte) ([]byte, error) {
	return l.V, nil
}

// ObjectResolverProvider is a provider of ObjectResolver
type ObjectResolverProvider func(m codec.Marshaler, libs []Object) (ObjectResolver, error)

// DefaultResolverProvider returns a default implements of ObjectResolver
func DefaultResolverProvider() ObjectResolverProvider {
	return func(m codec.Marshaler, libs []Object) (ObjectResolver, error) {
		return NewSequentialResolver(m, libs), nil
	}
}

// ObjectResolver resolves a given key to Object
type ObjectResolver interface {
	Resolve(id ChainID, key []byte) (Object, error)
}

// SequentialResolver is a resolver that resolves an object in sequential
type SequentialResolver struct {
	m       codec.Marshaler
	seq     uint8
	objects []Object
}

var _ ObjectResolver = (*SequentialResolver)(nil)

// NewSequentialResolver returns SequentialResolver
func NewSequentialResolver(m codec.Marshaler, objects []Object) *SequentialResolver {
	return &SequentialResolver{m: m, seq: 0, objects: objects}
}

// Resolve implements ObjectResolver.Resolve
// If success, resolver increments the internal sequence
func (r *SequentialResolver) Resolve(id ChainID, key []byte) (Object, error) {
	if len(r.objects) <= int(r.seq) {
		return nil, fmt.Errorf("object not found: seq=%X", r.seq)
	}
	obj := r.objects[r.seq]
	if !bytes.Equal(obj.Key(), key) {
		return nil, fmt.Errorf("keys mismatch: %X != %X", obj.Key(), key)
	}
	if cid := obj.GetChainID(r.m); !cid.Equal(id) {
		return nil, fmt.Errorf("chainID mismatch: %v != %v", cid, id)
	}
	r.seq++
	return obj, nil
}

// FakeResolver is a resolver that always fails to resolve an object
type FakeResolver struct{}

var _ ObjectResolver = (*FakeResolver)(nil)

// NewFakeResolver returns FakeResolver
func NewFakeResolver() FakeResolver {
	return FakeResolver{}
}

// Resolve implements ObjectResolver.Resolve
func (FakeResolver) Resolve(id ChainID, key []byte) (Object, error) {
	panic(fmt.Errorf("FakeResolver cannot resolve any objects, but received '%v' '%X'", id, key))
}
