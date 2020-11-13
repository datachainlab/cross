package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/gogo/protobuf/proto"
)

// msg types
const (
	TypeInitiate = "initiate"
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
	return TypeInitiate
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

func (msg MsgInitiateTx) GetAccounts() []Account {
	var accs []Account
	signers := msg.GetSigners()
	for _, id := range signers {
		accs = append(accs, NewLocalAccount(AccountID(id)))
	}
	return accs
}

func (msg MsgInitiateTx) GetRequiredAccounts() []Account {
	var accs []Account
	for _, tx := range msg.ContractTransactions {
		for _, id := range tx.Signers {
			accs = append(accs, Account{ChainId: tx.ChainId, Id: id})
		}
	}
	return accs
}

// MakeTxID generates TxID with a given msg
func MakeTxID(msg *MsgInitiateTx) TxID {
	bz, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}
