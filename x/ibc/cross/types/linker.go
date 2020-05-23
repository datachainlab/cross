package types

import (
	"encoding/hex"
	"fmt"

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

func (l CallResultLink) Type() LinkType {
	return LinkTypeCallResult
}

func (l CallResultLink) SourceIndex() TxIndex {
	return l.ContractTransactionIndex
}

type Linker struct {
	objects map[TxIndex]Object
}

func MakeLinker(txs ContractTransactions) Linker {
	lkr := Linker{objects: make(map[TxIndex]Object, len(txs))}
	for i, tx := range txs {
		idx := TxIndex(i)
		lkr.objects[idx] = MakeConstantValueObject(MakeObjectKey(tx.CallInfo, tx.Signers), tx.ReturnValue)
	}
	return lkr
}

func (lkr Linker) Lookup(lks []Link) ([]Object, error) {
	var objects []Object
	for _, lk := range lks {
		obj, ok := lkr.objects[lk.SourceIndex()]
		if !ok {
			return nil, fmt.Errorf("idx '%v' not found", lk.SourceIndex())
		}
		objects = append(objects, obj)
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

func (l ConstantValueObject) Evaluate(_ []byte) ([]byte, error) {
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
