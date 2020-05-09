package types

import (
	"errors"
	"fmt"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	CO_STATUS_NONE uint8 = iota
	CO_STATUS_INIT
	CO_STATUS_DECIDED // abort or commit

	CO_DECISION_NONE uint8 = iota
	CO_DECISION_COMMIT
	CO_DECISION_ABORT
)

const (
	MaxContractTransactoinNum = math.MaxUint8
)

type (
	TxID    = HexByteArray32
	TxIndex = uint8
)

type ContractHandler interface {
	GetState(ctx sdk.Context, contract []byte) (State, error)
	Handle(ctx sdk.Context, contract []byte) (State, ContractHandlerResult, error)
	OnCommit(ctx sdk.Context, result ContractHandlerResult) ContractHandlerResult
}

type ContractHandlerResult interface {
	GetData() []byte
	GetEvents() sdk.Events
}

type ContractHandlerAbortResult struct{}

func (ContractHandlerAbortResult) GetData() []byte {
	return nil
}

func (ContractHandlerAbortResult) GetEvents() sdk.Events {
	return nil
}

type CoordinatorInfo struct {
	Transactions []string      // {TransactionID => ConnectionID}
	Channels     []ChannelInfo // {TransactionID => Channel}

	Status                uint8
	Decision              uint8
	ConfirmedTransactions []TxIndex // [TransactionID]
	Acks                  []TxIndex // [TransactionID]
}

func NewCoordinatorInfo(status uint8, tss []string, channels []ChannelInfo) CoordinatorInfo {
	if len(tss) != len(channels) {
		panic("fatal error")
	}
	return CoordinatorInfo{Status: status, Transactions: tss, Channels: channels, Decision: CO_DECISION_NONE}
}

func (ci *CoordinatorInfo) Confirm(txIndex TxIndex, connectionID string) error {
	for _, id := range ci.ConfirmedTransactions {
		if txIndex == id {
			return errors.New("this transaction is already confirmed")
		}
	}

	if int(txIndex) >= len(ci.Transactions) {
		return fmt.Errorf("txIndex '%v' not found", txIndex)
	} else if cid := ci.Transactions[txIndex]; cid != connectionID {
		return fmt.Errorf("expected connectionID is '%v', but got '%v'", cid, connectionID)
	}

	ci.ConfirmedTransactions = append(ci.ConfirmedTransactions, txIndex)
	return nil
}

func (ci *CoordinatorInfo) IsCompleted() bool {
	return len(ci.Transactions) == len(ci.ConfirmedTransactions)
}

func (ci *CoordinatorInfo) AddAck(txIndex TxIndex) bool {
	for _, id := range ci.Acks {
		if txIndex == id {
			return false
		}
	}
	ci.Acks = append(ci.Acks, txIndex)
	return true
}

func (ci *CoordinatorInfo) IsReceivedALLAcks() bool {
	return len(ci.Transactions) == len(ci.Acks)
}

const (
	TX_STATUS_PREPARE uint8 = iota + 1
	TX_STATUS_COMMIT
	TX_STATUS_ABORT
)

type TxInfo struct {
	Status                  uint8  `json:"status" yaml:"status"`
	PrepareResult           uint8  `json:"prepare_result" yaml:"prepare_result"`
	CoordinatorConnectionID string `json:"coordinator_connection_id" yaml:"coordinator_connection_id"`
	Contract                []byte `json:"contract" yaml:"contract"`
}

func NewTxInfo(status, prepareResult uint8, coordinatorConnectionID string, contract []byte) TxInfo {
	return TxInfo{Status: status, PrepareResult: prepareResult, CoordinatorConnectionID: coordinatorConnectionID, Contract: contract}
}

type ContractCallResult struct {
	ChainID  string           `json:"chain_id" yaml:"chain_id"`
	Height   int64            `json:"height" yaml:"height"`
	Signers  []sdk.AccAddress `json:"signers" yaml:"signers"`
	Contract []byte           `json:"contract" yaml:"contract"`
	OPs      []OP             `json:"ops" yaml:"ops"`
}

func (r ContractCallResult) String() string {
	// TODO make this more readable
	return fmt.Sprintf("%#v", r)
}
