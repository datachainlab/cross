package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
)

// AccountID represents ID of account
// e.g. AccAddress in cosmos-SDK
type AccountID []byte

// AccountIDFromAccAddress converts given AccAddress to AccountID
func AccountIDFromAccAddress(acc sdk.AccAddress) AccountID {
	return AccountID(acc)
}

// AccAddress returns AccAddress
func (id AccountID) AccAddress() sdk.AccAddress {
	return sdk.AccAddress(id)
}

// Account definition

// NewAccount creates a new instance of Account
func NewAccount(id AccountID, authType AuthType) Account {
	return Account{Id: id, AuthType: authType}
}

// // GetCrossChainChannel returns CrossChainChannel
// func (acc Account) GetCrossChainChannel(m codec.Codec) xcctypes.XCC {
// 	xcc, err := xcctypes.UnpackCrossChainChannel(m, *acc.CrossChainChannel)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return xcc
// }

func NewAuthTypeLocal() AuthType {
	return AuthType{
		Mode: AuthMode_AUTH_MODE_LOCAL,
	}
}

func NewAuthTypeChannel(xcc xcctypes.XCC) AuthType {
	anyCrossChainChannel, err := xcctypes.PackCrossChainChannel(xcc)
	if err != nil {
		panic(err)
	}
	return AuthType{
		Mode:   AuthMode_AUTH_MODE_CHANNEL,
		Option: anyCrossChainChannel,
	}
}

func NewAuthTypeChannelWithAny(anyXCC *codectypes.Any) AuthType {
	return AuthType{
		Mode:   AuthMode_AUTH_MODE_CHANNEL,
		Option: anyXCC,
	}
}

func NewAuthTypeExtenstion(extension *codectypes.Any) AuthType {
	return AuthType{
		Mode:   AuthMode_AUTH_MODE_EXTENSION,
		Option: extension,
	}
}
