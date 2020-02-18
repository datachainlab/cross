package types

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

type PacketDataPrepare struct {
	Sender          sdk.AccAddress
	TxID            []byte
	TransitionID    int
	StateTransition StateTransition
}

func NewPacketDataPrepare(sender sdk.AccAddress, txID []byte, transitionID int, transition StateTransition) PacketDataPrepare {
	return PacketDataPrepare{Sender: sender, TxID: txID, TransitionID: transitionID, StateTransition: transition}
}

func (p PacketDataPrepare) Hash() []byte {
	b := ModuleCdc.MustMarshalBinaryBare(p)
	return tmhash.Sum(b)
}

func (p PacketDataPrepare) ValidateBasic() error {
	if err := p.StateTransition.ValidateBasic(); err != nil {
		return err
	}
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

const (
	PREPARE_STATUS_OK uint8 = iota + 1
	PREPARE_STATUS_FAILED
)

type PacketDataPrepareResult struct {
	Sender       sdk.AccAddress
	TxID         []byte
	TransitionID int
	Status       uint8
}

func NewPacketDataPrepareResult(sender sdk.AccAddress, txID []byte, transitionID int, status uint8) PacketDataPrepareResult {
	return PacketDataPrepareResult{Sender: sender, TxID: txID, TransitionID: transitionID, Status: status}
}

func (p PacketDataPrepareResult) ValidateBasic() error {
	return nil
}

func (p PacketDataPrepareResult) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(p))
}

func (p PacketDataPrepareResult) GetTimeoutHeight() uint64 {
	return math.MaxUint64
}

func (p PacketDataPrepareResult) Type() string {
	return "crossccc/prepareresult"
}

func (p PacketDataPrepareResult) IsOK() bool {
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
