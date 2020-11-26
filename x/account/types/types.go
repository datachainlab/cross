package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
)

type AccountID []byte

func AccountIDFromAccAddress(acc sdk.AccAddress) AccountID {
	return AccountID(acc)
}

func (id AccountID) AccAddress() sdk.AccAddress {
	return sdk.AccAddress(id)
}

// Account definition

// NewAccount creates a new instance of Account
func NewAccount(xcc xcctypes.XCC, id AccountID) Account {
	var anyCrossChainChannel *codectypes.Any
	if xcc != nil {
		var err error
		anyCrossChainChannel, err = xcctypes.PackCrossChainChannel(xcc)
		if err != nil {
			panic(err)
		}
	}
	return Account{CrossChainChannel: anyCrossChainChannel, Id: id}
}

// GetCrossChainChannel returns CrossChainChannel
func (acc Account) GetCrossChainChannel(m codec.Marshaler) xcctypes.XCC {
	xcc, err := xcctypes.UnpackCrossChainChannel(m, *acc.CrossChainChannel)
	if err != nil {
		panic(err)
	}
	return xcc
}
