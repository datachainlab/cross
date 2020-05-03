package types

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	PREPARE_STATUS_OK uint8 = iota + 1
	PREPARE_STATUS_FAILED
)

type PacketData interface {
	ValidateBasic() error
	GetBytes() []byte
	GetTimeoutHeight() uint64
	GetTimeoutTimestamp() uint64
	Type() string
}

type PacketAcknowledgement interface {
	ValidateBasic() error
	GetBytes() []byte
	Type() string
}

var _ PacketData = (*PacketDataPrepare)(nil)

type PacketDataPrepare struct {
	Sender              sdk.AccAddress
	TxID                TxID
	TxIndex             TxIndex
	ContractTransaction ContractTransaction
}

func NewPacketDataPrepare(sender sdk.AccAddress, txID TxID, txIndex TxIndex, transaction ContractTransaction) PacketDataPrepare {
	return PacketDataPrepare{Sender: sender, TxID: txID, TxIndex: txIndex, ContractTransaction: transaction}
}

func (p PacketDataPrepare) ValidateBasic() error {
	if err := p.ContractTransaction.ValidateBasic(); err != nil {
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

func (p PacketDataPrepare) GetTimeoutTimestamp() uint64 {
	return 0
}

func (p PacketDataPrepare) Type() string {
	return TypePrepare
}

var _ PacketAcknowledgement = (*PacketPrepareAcknowledgement)(nil)

type PacketPrepareAcknowledgement struct {
	Status uint8
}

func NewPacketPrepareAcknowledgement(status uint8) PacketPrepareAcknowledgement {
	return PacketPrepareAcknowledgement{Status: status}
}

func (p PacketPrepareAcknowledgement) ValidateBasic() error {
	return nil
}

func (p PacketPrepareAcknowledgement) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(p))
}

func (p PacketPrepareAcknowledgement) Type() string {
	return TypePrepareResult
}

func (p PacketPrepareAcknowledgement) IsOK() bool {
	return p.Status == PREPARE_STATUS_OK
}

var _ PacketData = (*PacketDataCommit)(nil)

type PacketDataCommit struct {
	TxID          TxID
	TxIndex       TxIndex
	IsCommittable bool
}

func NewPacketDataCommit(txID TxID, txIndex TxIndex, isCommittable bool) PacketDataCommit {
	return PacketDataCommit{TxID: txID, TxIndex: txIndex, IsCommittable: isCommittable}
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

func (p PacketDataCommit) GetTimeoutTimestamp() uint64 {
	return 0
}

func (p PacketDataCommit) Type() string {
	return TypeCommit
}

var _ PacketAcknowledgement = (*PacketCommitAcknowledgement)(nil)

type PacketCommitAcknowledgement struct{}

func NewPacketCommitAcknowledgement() PacketCommitAcknowledgement {
	return PacketCommitAcknowledgement{}
}

func (p PacketCommitAcknowledgement) ValidateBasic() error {
	return nil
}

func (p PacketCommitAcknowledgement) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(p))
}

func (p PacketCommitAcknowledgement) Type() string {
	return TypeAckCommit
}
