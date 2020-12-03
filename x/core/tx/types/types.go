package types

import (
	"bytes"
	fmt "fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"

	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	accounttypes "github.com/datachainlab/cross/x/core/account/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
	"github.com/datachainlab/cross/x/packets"
)

type (
	// TxID represents a ID of transaction. This value must be unique in a chain
	TxID = []byte
	// TxIndex represents an index of an array of contract transactions
	TxIndex = uint32
)

func NewTx(id TxID, commitProtocol CommitProtocol, ctxs []ResolvedContractTransaction, timeoutHeight clienttypes.Height, timeoutTimestamp uint64) Tx {
	return Tx{
		Id:                   id,
		CommitProtocol:       commitProtocol,
		ContractTransactions: ctxs,
		TimeoutHeight:        timeoutHeight,
		TimeoutTimestamp:     timeoutTimestamp,
	}
}

func (tx Tx) ValidateBasic() error {
	return nil
}

// TxRunner defines the expected transaction runner
type TxRunner interface {
	// RunTx executes a given transaction
	RunTx(ctx sdk.Context, tx Tx, ps packets.PacketSender) error
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
	GetCrossChainChannel(m codec.Marshaler) xcctypes.XCC
	WithCrossChainChannel(m codec.Marshaler, xcc xcctypes.XCC) Object
	Evaluate([]byte) ([]byte, error)
}

var _ Object = (*ConstantValueObject)(nil)

// MakeConstantValueObject returns ConstantValueObject
func MakeConstantValueObject(xcc xcctypes.XCC, key []byte, value []byte) ConstantValueObject {
	anyXCC, err := xcctypes.PackCrossChainChannel(xcc)
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
func (obj ConstantValueObject) GetCrossChainChannel(m codec.Marshaler) xcctypes.XCC {
	xcc, err := xcctypes.UnpackCrossChainChannel(m, obj.CrossChainChannel)
	if err != nil {
		panic(err)
	}
	return xcc
}

// WithChainID implements Object.WithCrossChainChannel
func (obj ConstantValueObject) WithCrossChainChannel(m codec.Marshaler, xcc xcctypes.XCC) Object {
	anyXCC, err := xcctypes.PackCrossChainChannel(xcc)
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
	Resolve(xcc xcctypes.XCC, key []byte) (Object, error)
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
func (r *SequentialResolver) Resolve(xcc xcctypes.XCC, key []byte) (Object, error) {
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
func (FakeResolver) Resolve(xcc xcctypes.XCC, key []byte) (Object, error) {
	panic(fmt.Errorf("FakeResolver cannot resolve any objects, but received '%v' '%X'", xcc, key))
}

func NewResolvedContractTransaction(anyXCC *codectypes.Any, signers []accounttypes.AccountID, callInfo ContractCallInfo, returnValue *ReturnValue, linkObjects []Object) ResolvedContractTransaction {
	anyObjects, err := PackObjects(linkObjects)
	if err != nil {
		panic(err)
	}
	return ResolvedContractTransaction{
		CrossChainChannel: anyXCC,
		Signers:           signers,
		CallInfo:          callInfo,
		ReturnValue:       returnValue,
		Objects:           anyObjects,
	}
}

func (tx ResolvedContractTransaction) ValidateBasic() error {
	return nil
}

func (tx ResolvedContractTransaction) UnpackObjects(m codec.Marshaler) []Object {
	objects, err := UnpackObjects(m, tx.Objects)
	if err != nil {
		panic(err)
	}
	return objects
}

func (tx ResolvedContractTransaction) GetCrossChainChannel(m codec.Marshaler) (xcctypes.XCC, error) {
	var xcc xcctypes.XCC
	if err := m.UnpackAny(tx.CrossChainChannel, &xcc); err != nil {
		return nil, err
	}
	return xcc, nil
}

type ContractCallInfo []byte

func NewReturnValue(v []byte) *ReturnValue {
	rv := ReturnValue{Value: v}
	return &rv
}

func (rv *ReturnValue) IsNil() bool {
	if rv == nil {
		return true
	}
	return false
}

func (rv *ReturnValue) Equal(other *ReturnValue) bool {
	if rv.IsNil() && other.IsNil() {
		return true
	} else if rv.IsNil() && !other.IsNil() {
		return false
	} else if !rv.IsNil() && other.IsNil() {
		return false
	} else {
		return bytes.Equal(rv.Value, other.Value)
	}
}
