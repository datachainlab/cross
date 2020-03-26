package types

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
)

const (
	PREPARE_STATUS_OK uint8 = iota + 1
	PREPARE_STATUS_FAILED
)

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

func (p PacketDataPrepare) Type() string {
	return TypePrepare
}

type PacketDataPrepareResult struct {
	Sender  sdk.AccAddress
	TxID    TxID
	TxIndex TxIndex
	Status  uint8
}

func NewPacketDataPrepareResult(sender sdk.AccAddress, txID TxID, txIndex TxIndex, status uint8) PacketDataPrepareResult {
	return PacketDataPrepareResult{Sender: sender, TxID: txID, TxIndex: txIndex, Status: status}
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
	return TypePrepareResult
}

func (p PacketDataPrepareResult) IsOK() bool {
	return p.Status == PREPARE_STATUS_OK
}

type PacketDataCommit struct {
	Sender        sdk.AccAddress
	TxID          TxID
	TxIndex       TxIndex
	IsCommittable bool
}

func NewPacketDataCommit(sender sdk.AccAddress, txID TxID, txIndex TxIndex, isCommittable bool) PacketDataCommit {
	return PacketDataCommit{Sender: sender, TxID: txID, TxIndex: txIndex, IsCommittable: isCommittable}
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
	return TypeCommit
}

var _ channelexported.PacketAcknowledgementI = AckDataCommit{}

type AckDataCommit struct {
	TxIndex TxIndex
}

func NewAckDataCommit(txIndex TxIndex) AckDataCommit {
	return AckDataCommit{TxIndex: txIndex}
}

// GetBytes implements channelexported.PacketAcknowledgementI
func (ack AckDataCommit) GetBytes() []byte {
	return []byte{ack.TxIndex}
}
