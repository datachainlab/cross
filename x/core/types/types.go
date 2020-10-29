package types

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type TxIndex = uint32

type AccountAddress []byte

func (ac AccountAddress) AccAddress() sdk.AccAddress {
	return sdk.AccAddress(ac)
}

func (tx ContractTransaction) GetChainID(m codec.Marshaler) (ChainID, error) {
	var chainID ChainID
	if err := m.UnpackAny(&tx.ChainId, &chainID); err != nil {
		return nil, err
	}
	return chainID, nil
}

// ChainID represents an ID of chain that contains a contract function to be called
type ChainID interface {
	Type() string
	Equal(ChainID) bool
	String() string
}

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

type ContractCallInfo []byte
