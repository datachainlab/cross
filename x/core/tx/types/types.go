package types

import (
	"bytes"
	fmt "fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"

	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	authtypes "github.com/datachainlab/cross/x/core/auth/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
	"github.com/datachainlab/cross/x/packets"
)

func NewTx(id crosstypes.TxID, commitProtocol CommitProtocol, ctxs []ResolvedContractTransaction, timeoutHeight clienttypes.Height, timeoutTimestamp uint64) Tx {
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

// CallResultType is a type of CallResult
type CallResultType = uint8

const (
	// ConstantValueCallResultType indicates a type of constant value
	ConstantValueCallResultType CallResultType = iota + 1
)

// CallResult wraps an actual value
type CallResult interface {
	proto.Message
	Type() CallResultType
	Key() []byte
	Value() []byte
	GetCrossChainChannel(m codec.Codec) xcctypes.XCC
	WithCrossChainChannel(m codec.Codec, xcc xcctypes.XCC) CallResult
}

var _ CallResult = (*ConstantValueCallResult)(nil)

// NewConstantValueCallResult returns ConstantValueObject
func NewConstantValueCallResult(xcc xcctypes.XCC, key []byte, value []byte) ConstantValueCallResult {
	anyXCC, err := xcctypes.PackCrossChainChannel(xcc)
	if err != nil {
		panic(err)
	}
	return ConstantValueCallResult{
		CrossChainChannel: *anyXCC,
		K:                 key,
		V:                 value,
	}
}

// Type implements CallResult.Type
func (ConstantValueCallResult) Type() CallResultType {
	return ConstantValueCallResultType
}

// GetCrossChainChannel implements CallResult.GetCrossChainChannel
func (r ConstantValueCallResult) GetCrossChainChannel(m codec.Codec) xcctypes.XCC {
	xcc, err := xcctypes.UnpackCrossChainChannel(m, r.CrossChainChannel)
	if err != nil {
		panic(err)
	}
	return xcc
}

// WithChainID implements CallResult.WithCrossChainChannel
func (r ConstantValueCallResult) WithCrossChainChannel(m codec.Codec, xcc xcctypes.XCC) CallResult {
	anyXCC, err := xcctypes.PackCrossChainChannel(xcc)
	if err != nil {
		panic(err)
	}
	r.CrossChainChannel = *anyXCC
	return &r
}

// Key implements CallResult.Key
func (r ConstantValueCallResult) Key() []byte {
	return r.K
}

// Evaluate returns a constant value
func (r ConstantValueCallResult) Value() []byte {
	return r.V
}

// CallResolverProvider is a provider of CallResultResolver
type CallResolverProvider func(m codec.Codec, results []CallResult) (CallResolver, error)

// DefaultCallResolverProvider returns a default implements of CallResolverProvider
func DefaultCallResolverProvider() CallResolverProvider {
	return func(m codec.Codec, results []CallResult) (CallResolver, error) {
		return NewSequentialResolver(m, results), nil
	}
}

// CallResolver resolves a given key to CallResult
type CallResolver interface {
	Resolve(xcc xcctypes.XCC, key []byte) (CallResult, error)
}

// SequentialResolver is a resolver that resolves a CallResult in sequential
type SequentialResolver struct {
	m       codec.Codec
	seq     uint8
	results []CallResult
}

var _ CallResolver = (*SequentialResolver)(nil)

// NewSequentialResolver returns SequentialResolver
func NewSequentialResolver(m codec.Codec, results []CallResult) *SequentialResolver {
	return &SequentialResolver{m: m, seq: 0, results: results}
}

// Resolve implements ObjectResolver.Resolve
// If success, resolver increments the internal sequence
func (r *SequentialResolver) Resolve(xcc xcctypes.XCC, key []byte) (CallResult, error) {
	if len(r.results) <= int(r.seq) {
		return nil, fmt.Errorf("result not found: seq=%X", r.seq)
	}
	obj := r.results[r.seq]
	if !bytes.Equal(obj.Key(), key) {
		return nil, fmt.Errorf("keys mismatch: %X != %X", obj.Key(), key)
	}
	if objXCC := obj.GetCrossChainChannel(r.m); !objXCC.Equal(xcc) {
		return nil, fmt.Errorf("cross-chain channel mismatch: %v != %v", objXCC, xcc)
	}
	r.seq++
	return obj, nil
}

// FakeResolver is a resolver that always fails to resolve an result
type FakeResolver struct{}

var _ CallResolver = (*FakeResolver)(nil)

// NewFakeResolver returns FakeResolver
func NewFakeResolver() FakeResolver {
	return FakeResolver{}
}

// Resolve implements CallResolver.Resolve
func (FakeResolver) Resolve(xcc xcctypes.XCC, key []byte) (CallResult, error) {
	panic(fmt.Errorf("FakeResolver cannot resolve any results, but received '%v' '%X'", xcc, key))
}

func NewResolvedContractTransaction(anyXCC *codectypes.Any, signers []authtypes.Account, callInfo ContractCallInfo, returnValue *ReturnValue, results []CallResult) ResolvedContractTransaction {
	anyResults, err := PackCallResults(results)
	if err != nil {
		panic(err)
	}
	return ResolvedContractTransaction{
		CrossChainChannel: anyXCC,
		Signers:           signers,
		CallInfo:          callInfo,
		ReturnValue:       returnValue,
		CallResults:       anyResults,
	}
}

func (tx ResolvedContractTransaction) ValidateBasic() error {
	return nil
}

func (tx ResolvedContractTransaction) UnpackCallResults(m codec.Codec) []CallResult {
	results, err := UnpackCallResults(m, tx.CallResults)
	if err != nil {
		panic(err)
	}
	return results
}

func (tx ResolvedContractTransaction) GetCrossChainChannel(m codec.Codec) (xcctypes.XCC, error) {
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
