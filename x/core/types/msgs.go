package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
)

// msg types
const (
	TypeInitiate = "initiate"
)

var _ sdk.Msg = (*MsgInitiate)(nil)

// NewMsgInitiate creates a new MsgInitiate instance
func NewMsgInitiate(
	sender sdk.AccAddress, receiver string,
	timeoutHeight clienttypes.Height, timeoutTimestamp uint64,
) *MsgInitiate {
	return &MsgInitiate{
		Sender:           sender.String(),
		TimeoutHeight:    timeoutHeight,
		TimeoutTimestamp: timeoutTimestamp,
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
	if msg.Sender == "" {
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
	signers := []sdk.AccAddress{mustAccAddressFromBech32(msg.Sender)}
	for _, t := range msg.ContractTransactions {
		for _, addr := range t.Signers {
			if !seen[addr] {
				signers = append(signers, mustAccAddressFromBech32(addr))
				seen[addr] = true
			}
		}
	}
	return signers
}

func mustAccAddressFromBech32(s string) sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(s)
	if err != nil {
		panic(err)
	}
	return addr
}
