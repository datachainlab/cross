package types

import (
	"encoding/hex"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
	"github.com/datachainlab/cross/x/utils"
	"github.com/gogo/protobuf/proto"
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

var _ codectypes.UnpackInterfacesMessage = (*Account)(nil)

// NewAccount creates a new instance of Account
func NewAccount(id AccountID, authType AuthType) Account {
	return Account{Id: id, AuthType: authType}
}

func (acc *Account) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if acc.AuthType.Option == nil {
		return nil
	}
	switch acc.AuthType.Mode {
	case AuthMode_AUTH_MODE_CHANNEL:
		return unpacker.UnpackAny(acc.AuthType.Option, new(xcctypes.XCC))
	case AuthMode_AUTH_MODE_EXTENSION:
		return unpacker.UnpackAny(acc.AuthType.Option, new(AuthExtensionVerifier))
	default:
		return nil
	}
}

func NewAccountFromHexString(s string) (*Account, error) {
	bz, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	var acc Account
	if err := proto.Unmarshal(bz, &acc); err != nil {
		return nil, err
	}
	return &acc, nil
}

func (acc Account) HexString() string {
	bz, err := proto.Marshal(&acc)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(bz)
}

func NewLocalAccount(id AccountID) Account {
	return Account{Id: id, AuthType: NewAuthTypeLocal()}
}

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

func NewAuthTypeExtenstion(extension AuthExtensionVerifier) AuthType {
	any, err := utils.PackAny(extension)
	if err != nil {
		panic(err)
	}
	return AuthType{
		Mode:   AuthMode_AUTH_MODE_EXTENSION,
		Option: any,
	}
}
