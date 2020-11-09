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

var _ sdk.Msg = (*MsgInitiate)(nil)

// NewMsgInitiate creates a new MsgInitiate instance
func NewMsgInitiate(
	sender AccountAddress, chainID string, nonce uint64,
	commitProtocol uint32, ctxs []ContractTransaction,
	timeoutHeight clienttypes.Height, timeoutTimestamp uint64,
) *MsgInitiate {
	return &MsgInitiate{
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
func (MsgInitiate) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (MsgInitiate) Type() string {
	return TypeInitiate
}

// ValidateBasic performs a basic check of the MsgInitiate fields.
// NOTE: timeout height or timestamp values can be 0 to disable the timeout.
func (msg MsgInitiate) ValidateBasic() error {
	if len(msg.Sender) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender address")
	}
	return nil
}

// GetSignBytes implements sdk.Msg. The function will panic since it is used
// for amino transaction verification which IBC does not support.
func (msg MsgInitiate) GetSignBytes() []byte {
	panic("IBC messages do not support amino")
}

// GetSigners implements sdk.Msg
// GetSigners returns the addresses that must sign the transaction.
// Addresses are returned in a deterministic order.
// Duplicate addresses will be omitted.
func (msg MsgInitiate) GetSigners() []sdk.AccAddress {
	seen := map[string]bool{}
	signers := []sdk.AccAddress{msg.Sender.AccAddress()}
	for _, t := range msg.ContractTransactions {
		for _, s := range t.Signers {
			addr := s.AccAddress().String()
			if !seen[addr] {
				signers = append(signers, s.AccAddress())
				seen[addr] = true
			}
		}
	}
	return signers
}

// MakeTxID generates TxID with a given msg
func MakeTxID(msg *MsgInitiate) TxID {
	bz, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}
