package types

import (
	"crypto/sha256"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/gogo/protobuf/proto"
)

// msg types
const (
	TypeInitiateTx = "InitiateTx"
	TypeSignTx     = "SignTx"
	TypeIBCSignTx  = "IBCSignTx"
)

var _ sdk.Msg = (*MsgInitiateTx)(nil)

// NewMsgInitiateTx creates a new MsgInitiateTx instance
func NewMsgInitiateTx(
	sender AccountID, chainID string, nonce uint64,
	commitProtocol uint32, ctxs []ContractTransaction,
	timeoutHeight clienttypes.Height, timeoutTimestamp uint64,
) *MsgInitiateTx {
	return &MsgInitiateTx{
		Sender:               sender,
		ChainId:              chainID,
		Nonce:                nonce,
		CommitProtocol:       commitProtocol,
		ContractTransactions: ctxs,
		TimeoutHeight:        timeoutHeight,
		TimeoutTimestamp:     timeoutTimestamp,
	}
}

// Route implements sdk.Msg
func (MsgInitiateTx) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (MsgInitiateTx) Type() string {
	return TypeInitiateTx
}

// ValidateBasic performs a basic check of the MsgInitiateTx fields.
// NOTE: timeout height or timestamp values can be 0 to disable the timeout.
func (msg MsgInitiateTx) ValidateBasic() error {
	if len(msg.Sender) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender address")
	}
	return nil
}

// GetSignBytes implements sdk.Msg. The function will panic since it is used
// for amino transaction verification which IBC does not support.
func (msg MsgInitiateTx) GetSignBytes() []byte {
	panic("IBC messages do not support amino")
}

// GetSigners implements sdk.Msg
// GetSigners returns the addresses that must sign the transaction.
// Addresses are returned in a deterministic order.
// Duplicate addresses will be omitted.
func (msg MsgInitiateTx) GetSigners() []sdk.AccAddress {
	seen := map[string]bool{}
	signers := []sdk.AccAddress{msg.Sender.AccAddress()}

	for _, s := range msg.Signers {
		addr := s.AccAddress().String()
		if !seen[addr] {
			signers = append(signers, s.AccAddress())
			seen[addr] = true
		}
	}

	return signers
}

func (msg MsgInitiateTx) GetAccounts(selfXCC CrossChainChannel) []Account {
	var accs []Account
	signers := msg.GetSigners()
	for _, id := range signers {
		accs = append(accs, NewAccount(selfXCC, AccountID(id)))
	}
	return accs
}

func (msg MsgInitiateTx) GetRequiredAccounts() []Account {
	var accs []Account
	for _, tx := range msg.ContractTransactions {
		for _, id := range tx.Signers {
			accs = append(accs, Account{CrossChainChannel: tx.CrossChainChannel, Id: id})
		}
	}
	return accs
}

var _ sdk.Msg = (*MsgSignTx)(nil)

// Route implements sdk.Msg
func (MsgSignTx) Route() string {
	return RouterKey
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

var _ sdk.Msg = (*MsgIBCSignTx)(nil)

// Route implements sdk.Msg
func (MsgIBCSignTx) Route() string {
	return RouterKey
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

// MakeTxID generates TxID with a given msg
func MakeTxID(msg *MsgInitiateTx) TxID {
	bz, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	txID := sha256.Sum256(bz)
	return txID[:]
}
