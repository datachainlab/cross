package types

import (
	"crypto/sha256"

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

var _ sdk.Msg = (*MsgInitiateTx)(nil)

// NewMsgInitiateTx creates a new MsgInitiateTx instance
func NewMsgInitiateTx(
	sender authtypes.AccountID, chainID string, nonce uint64,
	commitProtocol txtypes.CommitProtocol, ctxs []ContractTransaction,
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
	return crosstypes.RouterKey
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

func (msg MsgInitiateTx) GetAccounts(selfXCC xcctypes.XCC) []authtypes.Account {
	var accs []authtypes.Account
	signers := msg.GetSigners()
	for _, id := range signers {
		accs = append(accs, authtypes.NewAccount(selfXCC, authtypes.AccountID(id)))
	}
	return accs
}

func (msg MsgInitiateTx) GetRequiredAccounts() []authtypes.Account {
	var accs []authtypes.Account
	for _, tx := range msg.ContractTransactions {
		for _, id := range tx.Signers {
			accs = append(accs, authtypes.Account{CrossChainChannel: tx.CrossChainChannel, Id: id})
		}
	}
	return accs
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
