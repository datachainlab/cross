package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = (*MsgInitiate)(nil)

type MsgInitiate struct {
	Sender               sdk.AccAddress
	ChainID              string                // chainID of Coordinator node
	ContractTransactions []ContractTransaction // TODO: sorted by Source?
	TimeoutHeight        int64                 // Timeout for this msg
	Nonce                uint64
}

func NewMsgInitiate(sender sdk.AccAddress, chainID string, transactions []ContractTransaction, timeoutHeight int64, nonce uint64) MsgInitiate {
	return MsgInitiate{Sender: sender, ChainID: chainID, ContractTransactions: transactions, TimeoutHeight: timeoutHeight, Nonce: nonce}
}

// Route implements sdk.Msg
func (MsgInitiate) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (MsgInitiate) Type() string {
	return TypeInitiate
}

// ValidateBasic implements sdk.Msg
func (msg MsgInitiate) ValidateBasic() error {
	if l := len(msg.ContractTransactions); l == 0 {
		return errors.New("this msg includes no transisions")
	} else if l > MaxContractTransactoinNum {
		return fmt.Errorf("The number of ContractTransactions exceeds limit: %v > %v", l, MaxContractTransactoinNum)
	}
	for _, st := range msg.ContractTransactions {
		if err := st.ValidateBasic(); err != nil {
			return err
		}
	}
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgInitiate) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
// GetSigners returns the addresses that must sign the transaction.
// Addresses are returned in a deterministic order.
// Duplicate addresses will be omitted.
func (msg MsgInitiate) GetSigners() []sdk.AccAddress {
	seen := map[string]bool{}
	signers := []sdk.AccAddress{msg.Sender}
	for _, t := range msg.ContractTransactions {
		for _, addr := range t.Signers {
			if !seen[addr.String()] {
				signers = append(signers, addr)
				seen[addr.String()] = true
			}
		}
	}
	return signers
}

type ChannelInfo struct {
	Port    string `json:"port" yaml:"port"`       // the port on which the packet will be sent
	Channel string `json:"channel" yaml:"channel"` // the channel by which the packet will be sent
}

func NewChannelInfo(port, channel string) ChannelInfo {
	return ChannelInfo{Port: port, Channel: channel}
}

func (c ChannelInfo) ValidateBasic() error {
	if len(c.Port) == 0 || len(c.Channel) == 0 {
		return errors.New("port and channel must not be empty")
	}
	return nil
}

type ContractTransaction struct {
	Source ChannelInfo `json:"source" yaml:"source"`

	Signers  []sdk.AccAddress `json:"signers" yaml:"signers"`
	Contract []byte           `json:"contract" yaml:"contract"`
	OPs      []OP             `json:"ops" yaml:"ops"`
}

type ContractTransactions = []ContractTransaction

func NewContractTransaction(src ChannelInfo, signers []sdk.AccAddress, contract []byte, ops []OP) ContractTransaction {
	return ContractTransaction{
		Source:   src,
		Signers:  signers,
		Contract: contract,
		OPs:      ops,
	}
}

func (t ContractTransaction) ValidateBasic() error {
	if err := t.Source.ValidateBasic(); err != nil {
		return err
	}
	if len(t.Signers) == 0 {
		return errors.New("Signers must not be empty")
	}
	return nil
}
