package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	"math"
)

type PacketDataInitiate struct {
	Sender          sdk.AccAddress
	TxID            []byte
	StateTransition StateTransition
}

func NewPacketDataInitiate(sender sdk.AccAddress, txID []byte, transition StateTransition) PacketDataInitiate {
	return PacketDataInitiate{Sender: sender, TxID: txID, StateTransition: transition}
}

func (p PacketDataInitiate) Hash() []byte {
	b := ModuleCdc.MustMarshalBinaryBare(p)
	return tmhash.Sum(b)
}

func (p PacketDataInitiate) ValidateBasic() error {
	if err := p.StateTransition.ValidateBasic(); err != nil {
		return err
	}
	return nil
}

func (p PacketDataInitiate) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(p))
}

func (p PacketDataInitiate) GetTimeoutHeight() uint64 {
	return math.MaxUint64
}

func (p PacketDataInitiate) Type() string {
	return "crossccc/initiate"
}

const (
	PREPARE_STATUS_OK uint8 = iota + 1
	PREPARE_STATUS_FAILED
)

type PacketDataPrepare struct {
	Sender sdk.AccAddress
	TxID   []byte
	Status uint8
}

func NewPacketDataPrepare(sender sdk.AccAddress, txID []byte, status uint8) PacketDataPrepare {
	return PacketDataPrepare{Sender: sender, TxID: txID, Status: status}
}

func (p PacketDataPrepare) ValidateBasic() error {
	return nil
}

func (p PacketDataPrepare) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(p))
}

func (p PacketDataPrepare) GetTimeoutHeight() uint64 {
	return math.MaxUint64
}

func (p PacketDataPrepare) Type() string {
	return "crossccc/prepare"
}

func (p PacketDataPrepare) IsOK() bool {
	return p.Status == PREPARE_STATUS_OK
}

type PacketDataCommit struct {
	Sender        sdk.AccAddress
	TxID          []byte
	IsCommittable bool
}

func NewPacketDataCommit(sender sdk.AccAddress, txID []byte, isCommittable bool) PacketDataCommit {
	return PacketDataCommit{Sender: sender, TxID: txID, IsCommittable: isCommittable}
}

func (p PacketDataCommit) ValidateBasic() error {
	return nil
}

func (p PacketDataCommit) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(p))
}

func (p PacketDataCommit) GetTimeoutHeight() uint64 {
	return math.MaxUint64
}

func (p PacketDataCommit) Type() string {
	return "crossccc/commit"
}
