package types

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

// Linker resolves links that each ContractTransaction has.
type Linker struct {
	cdc           codec.Marshaler
	chainResolver CrossChainChannelResolver
	objects       map[TxIndex]lazyObject
}

// MakeLinker returns Linker
func MakeLinker(cdc codec.Marshaler, chainResolver CrossChainChannelResolver, txs []ContractTransaction) (*Linker, error) {
	lkr := Linker{cdc: cdc, objects: make(map[TxIndex]lazyObject, len(txs))}
	for i, tx := range txs {
		idx := TxIndex(i)
		lkr.objects[idx] = makeLazyObject(func() returnObject {
			if tx.ReturnValue.IsNil() {
				return returnObject{err: errors.New("On cross-chain call, each contractTransaction must be specified a return-value")}
			}
			xcc, err := tx.GetCrossChainChannel(lkr.cdc)
			if err != nil {
				return returnObject{err: err}
			}
			obj := MakeConstantValueObject(xcc, MakeObjectKey(tx.CallInfo, tx.Signers), tx.ReturnValue.Value)
			return returnObject{obj: &obj}
		})
	}
	return &lkr, nil
}

// Resolve resolves given links and returns resolved Object
func (lkr Linker) Resolve(ctx sdk.Context, callerc CrossChainChannel, lks []Link) ([]Object, error) {
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
		calleeID := ret.obj.GetCrossChainChannel(lkr.cdc)
		xcc, err := lkr.chainResolver.ConvertCrossChainChannel(ctx, calleeID, callerc)
		if err != nil {
			return nil, err
		}
		obj := ret.obj.WithCrossChainChannel(lkr.cdc, xcc)
		objects = append(objects, obj)
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
func MakeObjectKey(callInfo ContractCallInfo, signers []AccountID) []byte {
	h := tmhash.New()
	h.Write(callInfo)
	for _, signer := range signers {
		h.Write(signer)
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
	GetCrossChainChannel(m codec.Marshaler) CrossChainChannel
	WithCrossChainChannel(m codec.Marshaler, xcc CrossChainChannel) Object
	Evaluate([]byte) ([]byte, error)
}

var _ Object = (*ConstantValueObject)(nil)

// MakeConstantValueObject returns ConstantValueObject
func MakeConstantValueObject(xcc CrossChainChannel, key []byte, value []byte) ConstantValueObject {
	anyXCC, err := PackCrossChainChannel(xcc)
	if err != nil {
		panic(err)
	}
	return ConstantValueObject{
		CrossChainChannel: *anyXCC,
		K:                 key,
		V:                 value,
	}
}

// Type implements Object.Type
func (ConstantValueObject) Type() ObjectType {
	return ObjectTypeConstantValue
}

// GetCrossChainChannel implements Object.GetCrossChainChannel
func (obj ConstantValueObject) GetCrossChainChannel(m codec.Marshaler) CrossChainChannel {
	xcc, err := UnpackCrossChainChannel(m, obj.CrossChainChannel)
	if err != nil {
		panic(err)
	}
	return xcc
}

// WithChainID implements Object.WithCrossChainChannel
func (obj ConstantValueObject) WithCrossChainChannel(m codec.Marshaler, xcc CrossChainChannel) Object {
	anyXCC, err := PackCrossChainChannel(xcc)
	if err != nil {
		panic(err)
	}
	obj.CrossChainChannel = *anyXCC
	return &obj
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
	Resolve(xcc CrossChainChannel, key []byte) (Object, error)
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
func (r *SequentialResolver) Resolve(xcc CrossChainChannel, key []byte) (Object, error) {
	if len(r.objects) <= int(r.seq) {
		return nil, fmt.Errorf("object not found: seq=%X", r.seq)
	}
	obj := r.objects[r.seq]
	if !bytes.Equal(obj.Key(), key) {
		return nil, fmt.Errorf("keys mismatch: %X != %X", obj.Key(), key)
	}
	if objXCC := obj.GetCrossChainChannel(r.m); !objXCC.Equal(xcc) {
		return nil, fmt.Errorf("cross-chain channel mismatch: %v != %v", objXCC, xcc)
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
func (FakeResolver) Resolve(xcc CrossChainChannel, key []byte) (Object, error) {
	panic(fmt.Errorf("FakeResolver cannot resolve any objects, but received '%v' '%X'", xcc, key))
}
