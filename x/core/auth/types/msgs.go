package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
)

// msg types
const (
	TypeSignTx    = "SignTx"
	TypeIBCSignTx = "IBCSignTx"
	TypeExtSignTx = "ExtSignTx"
)

var _ sdk.Msg = (*MsgSignTx)(nil)

// Route implements sdk.Msg
func (MsgSignTx) Route() string {
	return crosstypes.RouterKey
}

// Type implements sdk.Msg
func (MsgSignTx) Type() string {
	return TypeSignTx
}

// ValidateBasic performs a basic check of the MsgInitiateTx fields.
// NOTE: timeout height or timestamp values can be 0 to disable the timeout.
func (msg MsgSignTx) ValidateBasic() error {
	if len(msg.Signers) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing signers")
	}
	return nil
}

// GetSignBytes implements sdk.Msg. The function will panic since it is used
// for amino transaction verification which IBC does not support.
func (msg MsgSignTx) GetSignBytes() []byte {
	panic("IBC messages do not support amino")
}

// GetSigners implements sdk.Msg
// GetSigners returns the addresses that must sign the transaction.
// Addresses are returned in a deterministic order.
// Duplicate addresses will be omitted.
func (msg MsgSignTx) GetSigners() []sdk.AccAddress {
	seen := map[string]bool{}
	signers := []sdk.AccAddress{}

	for _, s := range msg.Signers {
		addr := s.AccAddress().String()
		if !seen[addr] {
			signers = append(signers, s.AccAddress())
			seen[addr] = true
		}
	}

	return signers
}

var (
	_ sdk.Msg                            = (*MsgIBCSignTx)(nil)
	_ codectypes.UnpackInterfacesMessage = (*MsgIBCSignTx)(nil)
)

// NewMsgIBCSignTx creates a new instance of MsgIBCSignTx
func NewMsgIBCSignTx(
	anyXCC *codectypes.Any, txID crosstypes.TxID, signers []AccountID,
	timeoutHeight clienttypes.Height, timeoutTimestamp uint64,
) *MsgIBCSignTx {
	return &MsgIBCSignTx{
		CrossChainChannel: anyXCC,
		TxID:              txID,
		Signers:           signers,
		TimeoutHeight:     timeoutHeight,
		TimeoutTimestamp:  timeoutTimestamp,
	}
}

// Route implements sdk.Msg
func (MsgIBCSignTx) Route() string {
	return crosstypes.RouterKey
}

// Type implements sdk.Msg
func (MsgIBCSignTx) Type() string {
	return TypeIBCSignTx
}

// ValidateBasic performs a basic check of the MsgInitiateTx fields.
// NOTE: timeout height or timestamp values can be 0 to disable the timeout.
func (msg MsgIBCSignTx) ValidateBasic() error {
	if len(msg.Signers) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing signers")
	}
	return nil
}

// GetSignBytes implements sdk.Msg. The function will panic since it is used
// for amino transaction verification which IBC does not support.
func (msg MsgIBCSignTx) GetSignBytes() []byte {
	panic("IBC messages do not support amino")
}

// GetSigners implements sdk.Msg
// GetSigners returns the addresses that must sign the transaction.
// Addresses are returned in a deterministic order.
// Duplicate addresses will be omitted.
func (msg MsgIBCSignTx) GetSigners() []sdk.AccAddress {
	seen := map[string]bool{}
	signers := []sdk.AccAddress{}

	for _, s := range msg.Signers {
		addr := s.AccAddress().String()
		if !seen[addr] {
			signers = append(signers, s.AccAddress())
			seen[addr] = true
		}
	}

	return signers
}

// UnpackInterfaces implements UnpackInterfacesMessage
func (msg *MsgIBCSignTx) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return unpacker.UnpackAny(msg.CrossChainChannel, new(xcctypes.XCC))
}

// ExtAuthMsg defines an interface that supports an extension signing method
type ExtAuthMsg interface {
	GetSignerAccounts() []Account
}

var (
	_ sdk.Msg                            = (*MsgExtSignTx)(nil)
	_ ExtAuthMsg                         = (*MsgExtSignTx)(nil)
	_ codectypes.UnpackInterfacesMessage = (*MsgExtSignTx)(nil)
)

// ValidateBasic does a simple validation check that
// doesn't require access to any other information.
func (msg MsgExtSignTx) ValidateBasic() error {
	return nil
}

// Signers returns the addrs of signers that must sign.
// CONTRACT: All signatures must be present to be valid.
// CONTRACT: Returns addrs in some deterministic order.
func (msg MsgExtSignTx) GetSigners() []types.AccAddress {
	seen := map[string]bool{}
	signers := []sdk.AccAddress{}

	for _, s := range msg.Signers {
		acc := s.HexString()
		if !seen[acc] {
			signers = append(signers, s.Id.AccAddress())
			seen[acc] = true
		}
	}

	return signers
}

// Route implements sdk.Msg
func (MsgExtSignTx) Route() string {
	return crosstypes.RouterKey
}

// Type implements sdk.Msg
func (MsgExtSignTx) Type() string {
	return TypeExtSignTx
}

// GetSignerAccounts implements ExtAuthMsg
func (msg MsgExtSignTx) GetSignerAccounts() []Account {
	return msg.Signers
}

// UnpackInterfaces implements UnpackInterfacesMessage
func (msg *MsgExtSignTx) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, signer := range msg.Signers {
		if err := signer.UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}
	return nil
}
