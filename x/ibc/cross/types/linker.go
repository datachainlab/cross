package types

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

// Link is a link to an Object that is referenced in the call to the contract
type Link interface {
	Type() LinkType
	SourceIndex() TxIndex
}

// LinkType is a type of link
type LinkType = uint8

const (
	// LinkTypeCallResult is LinkType that indicates a link with an object returned from external contract call
	LinkTypeCallResult LinkType = iota + 1
)

var _ Link = (*CallResultLink)(nil)

// CallResultLink is a link with an object returned from external contract call
type CallResultLink struct {
	ContractTransactionIndex TxIndex
}

// NewCallResultLink returns CallResultLink
func NewCallResultLink(idx TxIndex) CallResultLink {
	return CallResultLink{ContractTransactionIndex: idx}
}

// Type implements Link.Type
func (l CallResultLink) Type() LinkType {
	return LinkTypeCallResult
}

// SourceIndex implements Link.SourceIndex
func (l CallResultLink) SourceIndex() TxIndex {
	return l.ContractTransactionIndex
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

// Linker resolves links that each ContractTransaction has.
type Linker struct {
	objects map[TxIndex]lazyObject
}

// MakeLinker returns Linker
func MakeLinker(txs ContractTransactions) (*Linker, error) {
	lkr := Linker{objects: make(map[TxIndex]lazyObject, len(txs))}
	for i, tx := range txs {
		idx := TxIndex(i)
		lkr.objects[idx] = makeLazyObject(func() returnObject {
			if tx.ReturnValue.IsNil() {
				return returnObject{err: errors.New("On cross-chain call, each contractTransaction must be specified a return-value")}
			}
			return returnObject{obj: MakeConstantValueObject(MakeObjectKey(tx.CallInfo, tx.Signers), *tx.ReturnValue)}
		})
	}
	return &lkr, nil
}

// Resolve resolves given links and returns resolved Object
func (lkr Linker) Resolve(lks []Link) ([]Object, error) {
	var objects []Object
	for _, lk := range lks {
		lzObj, ok := lkr.objects[lk.SourceIndex()]
		if !ok {
			return nil, fmt.Errorf("idx '%v' not found", lk.SourceIndex())
		}
		ret := lzObj()
		if ret.err != nil {
			return nil, ret.err
		}
		objects = append(objects, ret.obj)
	}
	return objects, nil
}

// MakeObjectKey returns a key that can be used to identify a contract call
func MakeObjectKey(callInfo ContractCallInfo, signers []sdk.AccAddress) []byte {
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
	Type() ObjectType
	Key() []byte
	Evaluate([]byte) ([]byte, error)
}

var _ Object = (*ConstantValueObject)(nil)

// ConstantValueObject is an Object wraps a constant value
type ConstantValueObject struct {
	K []byte
	V []byte
}

// MakeConstantValueObject returns ConstantValueObject
func MakeConstantValueObject(key []byte, value []byte) ConstantValueObject {
	return ConstantValueObject{
		K: key,
		V: value,
	}
}

// Type implements Object.Type
func (l ConstantValueObject) Type() ObjectType {
	return ObjectTypeConstantValue
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
type ObjectResolverProvider func(libs []Object) (ObjectResolver, error)

// DefaultResolverProvider returns a default implements of ObjectResolver
func DefaultResolverProvider() ObjectResolverProvider {
	return func(libs []Object) (ObjectResolver, error) {
		return NewSequentialResolver(libs), nil
	}
}

// ObjectResolver resolves a given key to Object
type ObjectResolver interface {
	Resolve(key []byte) (Object, error)
}

// SequentialResolver is a resolver that resolves an object in sequential
type SequentialResolver struct {
	seq     uint8
	objects []Object
}

var _ ObjectResolver = (*SequentialResolver)(nil)

// NewSequentialResolver returns SequentialResolver
func NewSequentialResolver(objects []Object) *SequentialResolver {
	return &SequentialResolver{seq: 0, objects: objects}
}

// Resolve implements ObjectResolver.Resolve
// If success, resolver increments the internal sequence
func (r *SequentialResolver) Resolve(key []byte) (Object, error) {
	if len(r.objects) <= int(r.seq) {
		return nil, fmt.Errorf("object not found: seq=%X", r.seq)
	}
	obj := r.objects[r.seq]
	if !bytes.Equal(obj.Key(), key) {
		return nil, fmt.Errorf("keys mismatch: %X != %X", obj.Key(), key)
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
func (FakeResolver) Resolve(key []byte) (Object, error) {
	panic(fmt.Errorf("FakeResolver cannot resolve any objects, but received '%X'", key))
}
