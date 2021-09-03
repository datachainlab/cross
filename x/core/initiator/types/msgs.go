package types

import (
	"crypto/sha256"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	authtypes "github.com/datachainlab/cross/x/core/auth/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
	"github.com/gogo/protobuf/proto"
)

// msg types
const (
	TypeInitiateTx = "InitiateTx"
)

var (
	_ sdk.Msg                            = (*MsgInitiateTx)(nil)
	_ authtypes.ExtAuthMsg               = (*MsgInitiateTx)(nil)
	_ codectypes.UnpackInterfacesMessage = (*MsgInitiateTx)(nil)
)

// NewMsgInitiateTx creates a new MsgInitiateTx instance
func NewMsgInitiateTx(
	signers []authtypes.Account, chainID string, nonce uint64,
	commitProtocol txtypes.CommitProtocol, ctxs []ContractTransaction,
	timeoutHeight clienttypes.Height, timeoutTimestamp uint64,
) *MsgInitiateTx {
	return &MsgInitiateTx{
		ChainId:              chainID,
		Nonce:                nonce,
		CommitProtocol:       commitProtocol,
		ContractTransactions: ctxs,
		Signers:              signers,
		TimeoutHeight:        timeoutHeight,
		TimeoutTimestamp:     timeoutTimestamp,
	}
}

// Route implements sdk.Msg
func (MsgInitiateTx) Route() string {
	return crosstypes.RouterKey
}

// Type implements sdk.Msg
func (MsgInitiateTx) Type() string {
	return TypeInitiateTx
}

// ValidateBasic performs a basic check of the MsgInitiateTx fields.
// NOTE: timeout height or timestamp values can be 0 to disable the timeout.
func (msg MsgInitiateTx) ValidateBasic() error {
	if len(msg.Signers) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing signer address")
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

// GetSignerAccounts implements authtypes.ExtAuthMsg
func (msg MsgInitiateTx) GetSignerAccounts() []authtypes.Account {
	return msg.Signers
}

func (msg MsgInitiateTx) GetRequiredAccounts() []authtypes.Account {
	var accs []authtypes.Account
	for _, tx := range msg.ContractTransactions {
		accs = append(accs, tx.Signers...)
	}
	return accs
}

// UnpackInterfaces implements UnpackInterfacesMessage
func (msg *MsgInitiateTx) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, tx := range msg.ContractTransactions {
		if err := tx.UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}
	for _, signer := range msg.Signers {
		if err := signer.UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}
	return nil
}

var _ codectypes.UnpackInterfacesMessage = (*ContractTransaction)(nil)

func (tx *ContractTransaction) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if err := unpacker.UnpackAny(tx.CrossChainChannel, new(xcctypes.XCC)); err != nil {
		return err
	}
	for _, signer := range tx.Signers {
		if err := signer.UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}
	return nil
}

// MakeTxID generates TxID with a given msg
func MakeTxID(msg *MsgInitiateTx) crosstypes.TxID {
	bz, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	txID := sha256.Sum256(bz)
	return txID[:]
}
