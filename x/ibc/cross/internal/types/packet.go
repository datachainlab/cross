package types

import (
	"encoding/binary"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
)

type PacketDataPrepare struct {
	Sender              sdk.AccAddress
	TxID                TxID
	TransactionID       int
	ContractTransaction ContractTransaction
}

func NewPacketDataPrepare(sender sdk.AccAddress, txID TxID, transactionID int, transaction ContractTransaction) PacketDataPrepare {
	return PacketDataPrepare{Sender: sender, TxID: txID, TransactionID: transactionID, ContractTransaction: transaction}
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
	return "cross/prepare"
}

const (
	PREPARE_STATUS_OK uint8 = iota + 1
	PREPARE_STATUS_FAILED
)

type PacketDataPrepareResult struct {
	Sender        sdk.AccAddress
	TxID          TxID
	TransactionID int
	Status        uint8
}

func NewPacketDataPrepareResult(sender sdk.AccAddress, txID TxID, transactionID int, status uint8) PacketDataPrepareResult {
	return PacketDataPrepareResult{Sender: sender, TxID: txID, TransactionID: transactionID, Status: status}
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
	return "cross/prepareresult"
}

func (p PacketDataPrepareResult) IsOK() bool {
	return p.Status == PREPARE_STATUS_OK
}

type PacketDataCommit struct {
	Sender        sdk.AccAddress
	TxID          TxID
	IsCommittable bool
}

func NewPacketDataCommit(sender sdk.AccAddress, txID TxID, isCommittable bool) PacketDataCommit {
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
	return "cross/commit"
}

var _ channelexported.PacketAcknowledgementI = AckDataCommit{}

type AckDataCommit struct {
	TransactionID int
}

func NewAckDataCommit(transactionID int) AckDataCommit {
	return AckDataCommit{TransactionID: transactionID}
}

// GetBytes implements channelexported.PacketAcknowledgementI
func (ack AckDataCommit) GetBytes() []byte {
	var bz [8]byte
	binary.BigEndian.PutUint64(bz[:], uint64(ack.TransactionID))
	return bz[:]
}
