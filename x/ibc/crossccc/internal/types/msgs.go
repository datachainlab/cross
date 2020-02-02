package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

var _ sdk.Msg = (*MsgInitiate)(nil)

type MsgInitiate struct {
	Sender           sdk.AccAddress
	StateTransitions []StateTransition // TODO: sorted by Source?
	Nonce            uint64
}

func NewMsgInitiate(sender sdk.AccAddress, transitions []StateTransition, Nonce uint64) MsgInitiate {
	return MsgInitiate{Sender: sender, StateTransitions: transitions, Nonce: Nonce}
}

// Route implements sdk.Msg
func (MsgInitiate) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (MsgInitiate) Type() string {
	return "crossccc_initiate"
}

// ValidateBasic implements sdk.Msg
func (msg MsgInitiate) ValidateBasic() error {
	if len(msg.StateTransitions) == 0 {
		return errors.New("this msg includes no transisions")
	}
	for _, st := range msg.StateTransitions {
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
func (msg MsgInitiate) GetSigners() []sdk.AccAddress {
	addrs := []sdk.AccAddress{msg.Sender}
	for _, t := range msg.StateTransitions {
		addrs = append(addrs, t.Signer)
	}
	return addrs
}

func (msg MsgInitiate) GetTxID() []byte {
	return tmhash.Sum(msg.GetSignBytes())
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

type StateTransition struct {
	Source ChannelInfo `json:"source" yaml:"source"`

	Signer   sdk.AccAddress `json:"signer" yaml:"signer"`
	Contract []byte         `json:"contract" yaml:"contract"`
	OPs      []OP           `json:"ops" yaml:"ops"`
}

func NewStateTransition(src ChannelInfo, signer sdk.AccAddress, contract []byte, ops []OP) StateTransition {
	return StateTransition{
		Source:   src,
		Signer:   signer,
		Contract: contract,
		OPs:      ops,
	}
}

func (t StateTransition) ValidateBasic() error {
	if err := t.Source.ValidateBasic(); err != nil {
		return err
	}
	return nil
}

var _ sdk.Msg = (*MsgConfirm)(nil)

type MsgConfirm struct {
	TxID           []byte
	PreparePackets []PreparePacket
	Signer         sdk.AccAddress
}

type MultiplePackets interface {
	Packets() []channel.MsgPacket
}

var _ MultiplePackets = (*MsgConfirm)(nil)

type PreparePacket struct {
	Packet channel.MsgPacket
	Status uint8       `json:"status" yaml:"status"`
	Source ChannelInfo `json:"source" yaml:"source"`
}

func NewPreparePacket(msgPacket channel.MsgPacket, status uint8, src ChannelInfo) PreparePacket {
	return PreparePacket{Packet: msgPacket, Status: status, Source: src}
}

type Statuses []uint8

func NewMsgConfirm(txID []byte, prepares []PreparePacket, signer sdk.AccAddress) MsgConfirm {
	return MsgConfirm{TxID: txID, PreparePackets: prepares, Signer: signer}
}

func (m MsgConfirm) Packets() []channel.MsgPacket {
	packets := make([]channel.MsgPacket, len(m.PreparePackets))
	for i, p := range m.PreparePackets {
		packets[i] = p.Packet
	}
	return packets
}

func (m MsgConfirm) IsCommittable() bool {
	for _, p := range m.PreparePackets {
		if p.Status == PREPARE_STATUS_FAILED {
			return false
		}
	}
	return true
}

func (MsgConfirm) Route() string {
	return RouterKey
}

func (MsgConfirm) Type() string {
	return "crossccc_confirm"
}

func (msg MsgConfirm) ValidateBasic() error {
	if len(msg.TxID) == 0 {
		return errors.New("TxID is required")
	} else if len(msg.Signer) == 0 {
		return errors.New("Signer is required")
	} else if len(msg.PreparePackets) == 0 {
		return errors.New("PrepareInfoList must not be empty")
	} else {
		return nil
	}
}

func (msg MsgConfirm) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgConfirm) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
