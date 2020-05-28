package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

type Link interface {
	Type() LinkType
	SourceIndex() TxIndex
}

type LinkType = uint8

const (
	LinkTypeCallResult LinkType = iota + 1
)

var _ Link = (*CallResultLink)(nil)

type CallResultLink struct {
	ContractTransactionIndex TxIndex
}

func NewCallResultLink(idx TxIndex) CallResultLink {
	return CallResultLink{ContractTransactionIndex: idx}
}

func (l CallResultLink) Type() LinkType {
	return LinkTypeCallResult
}

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

type Linker struct {
	objects map[TxIndex]lazyObject
}

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

func (lkr Linker) Lookup(lks []Link) ([]Object, error) {
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

func MakeObjectKey(callInfo ContractCallInfo, signers []sdk.AccAddress) []byte {
	h := tmhash.New()
	h.Write(callInfo)
	for _, signer := range signers {
		h.Write(signer)
	}
	return h.Sum(nil)
}

type ObjectType = uint8

const (
	ObjectTypeConstantValue ObjectType = iota + 1
)

type Object interface {
	Type() ObjectType
	Key() []byte
	Evaluate([]byte) ([]byte, error)
}

var _ Object = (*ConstantValueObject)(nil)

type ConstantValueObject struct {
	K []byte
	V []byte
}

func MakeConstantValueObject(key []byte, value []byte) ConstantValueObject {
	return ConstantValueObject{
		K: key,
		V: value,
	}
}

func (l ConstantValueObject) Type() ObjectType {
	return ObjectTypeConstantValue
}

func (l ConstantValueObject) Key() []byte {
	return l.K
}

func (l ConstantValueObject) Evaluate(bz []byte) ([]byte, error) {
	return l.V, nil
}

type Resolver interface {
	Resolve(bz []byte) (Object, error)
}

type MapResolver struct {
	libs map[string]Object
}

func MakeResolver(libs []Object) (*MapResolver, error) {
	r := &MapResolver{libs: make(map[string]Object)}

	for _, lib := range libs {
		key := hex.EncodeToString(lib.Key())
		if _, ok := r.libs[key]; ok {
			return nil, fmt.Errorf("duplicated key '%X'", lib.Key())
		}
		r.libs[key] = lib
	}

	return r, nil
}

func (r MapResolver) Resolve(bz []byte) (Object, error) {
	key := hex.EncodeToString(bz)
	lib, ok := r.libs[key]
	if !ok {
		return nil, fmt.Errorf("key not found: %X", bz)
	}
	return lib, nil
}

type FakeResolver struct{}

var _ Resolver = (*FakeResolver)(nil)

func NewFakeResolver() FakeResolver {
	return FakeResolver{}
}

func (FakeResolver) Resolve(bz []byte) (Object, error) {
	panic(fmt.Errorf("FakeResolver cannot resolve any objects, but received '%X'", bz))
}
