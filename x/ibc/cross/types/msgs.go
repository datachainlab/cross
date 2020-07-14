package types

import (
	"bytes"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeInitiate = "cross_initiate"

var _ sdk.Msg = (*MsgInitiate)(nil)

// MsgInitiate initiates a Cross-chain transaction
type MsgInitiate struct {
	Sender               sdk.AccAddress
	ChainID              string // chainID of Coordinator node
	ContractTransactions []ContractTransaction
	TimeoutHeight        int64 // Timeout for this msg
	Nonce                uint64
	CommitProtocol       uint8 // Commit type
}

// NewMsgInitiate returns MsgInitiate
func NewMsgInitiate(sender sdk.AccAddress, chainID string, transactions []ContractTransaction, timeoutHeight int64, nonce uint64, cp uint8) MsgInitiate {
	return MsgInitiate{Sender: sender, ChainID: chainID, ContractTransactions: transactions, TimeoutHeight: timeoutHeight, Nonce: nonce, CommitProtocol: cp}
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
	} else {
		switch msg.CommitProtocol {
		case COMMIT_PROTOCOL_NAIVE:
			if l != 2 {
				return fmt.Errorf("For Commit Protocol 'naive', the number of ContractTransactions must be 2")
			}
			if src := msg.ContractTransactions[0].Source; src.Channel != "" || src.Port != "" {
				return fmt.Errorf("ContractTransactions[0] must be an empty source")
			}
		case COMMIT_PROTOCOL_TPC:
		default:
			return fmt.Errorf("unknown Commit Protocol '%v'", msg.CommitProtocol)
		}
	}

	for id, st := range msg.ContractTransactions {
		if err := st.ValidateBasic(); err != nil {
			return err
		}
		for _, link := range st.Links {
			if src := link.SourceIndex(); src == TxIndex(id) {
				return fmt.Errorf("cyclic reference is unsupported: index=%v", src)
			} else if src >= TxIndex(len(msg.ContractTransactions)) {
				return fmt.Errorf("not found index: index=%v", src)
			} else {
			} // OK
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

type ContractCallInfo []byte

type ReturnValue []byte

func NewReturnValue(v []byte) *ReturnValue {
	rv := ReturnValue(v)
	return &rv
}

func (rv *ReturnValue) IsNil() bool {
	if rv == nil {
		return true
	}
	return false
}

func (rv *ReturnValue) Equal(bz []byte) bool {
	if rv.IsNil() {
		return false
	}
	return bytes.Equal(*rv, bz)
}

type ContractTransaction struct {
	Source          ChannelInfo      `json:"source" yaml:"source"`
	Signers         []sdk.AccAddress `json:"signers" yaml:"signers"`
	CallInfo        ContractCallInfo `json:"call_info" yaml:"call_info"`
	StateConstraint StateConstraint  `json:"state_constraint" yaml:"state_constraint"`
	ReturnValue     *ReturnValue     `json:"return_value" yaml:"return_value"`
	Links           []Link           `json:"links" yaml:"links"`
}

func NewContractTransaction(src ChannelInfo, signers []sdk.AccAddress, callInfo ContractCallInfo, sc StateConstraint, rv *ReturnValue, links []Link) ContractTransaction {
	return ContractTransaction{
		Source:          src,
		Signers:         signers,
		CallInfo:        callInfo,
		StateConstraint: sc,
		ReturnValue:     rv,
		Links:           links,
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

type ContractTransactionInfo struct {
	Transaction ContractTransaction `json:"transaction" yaml:"transaction"`
	LinkObjects []Object            `json:"link_objects" yaml:"link_objects"`
}

func NewContractTransactionInfo(tx ContractTransaction, linkObjects []Object) ContractTransactionInfo {
	return ContractTransactionInfo{
		Transaction: tx,
		LinkObjects: linkObjects,
	}
}

func (ti ContractTransactionInfo) ValidateBasic() error {
	return ti.ValidateBasic()
}

type StateConstraintType = uint8

const (
	NoStateConstraint         StateConstraintType = iota // NoStateConstraint indicates that no constraints on the state before and after the precommit is performed
	ExactMatchStateConstraint                            // ExactMatchStateConstraint indicates the constraint on state state before and after the precommit is performed
	PreStateConstraint                                   // PreStateConstraint indicates the constraint on state before the precommit is performed
	PostStateConstraint                                  // PostStateConstraint indicates the constraint on state after the precommit is performed
)

type StateConstraint struct {
	Type StateConstraintType `json:"type" yaml:"type"`
	OPs  OPs                 `json:"ops" yaml:"ops"`
}

func NewStateConstraint(tp StateConstraintType, ops OPs) StateConstraint {
	return StateConstraint{Type: tp, OPs: ops}
}

type ContractTransactions = []ContractTransaction
