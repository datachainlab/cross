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
)

const (
	CO_DECISION_NONE uint8 = iota
	CO_DECISION_COMMIT
	CO_DECISION_ABORT
)

const (
	COMMIT_PROTOCOL_SIMPLE uint8 = iota // Default
	COMMIT_PROTOCOL_TPC                 // Two-phase commit
)

const (
	PREPARE_RESULT_OK uint8 = iota + 1
	PREPARE_RESULT_FAILED
)

const (
	MaxContractTransactoinNum = math.MaxUint8
)

type (
	TxID    = HexByteArray32
	TxIndex = uint8
)

type ContractRuntimeInfo struct {
	StateConstraintType    StateConstraintType
	ExternalObjectResolver ObjectResolver
}

type ContractHandler interface {
	GetState(ctx sdk.Context, callInfo ContractCallInfo, rtInfo ContractRuntimeInfo) (State, error)
	Handle(ctx sdk.Context, callInfo ContractCallInfo, rtInfo ContractRuntimeInfo) (State, ContractHandlerResult, error)
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
	Transactions []string      // {TxIndex => ConnectionID}
	Channels     []ChannelInfo // {TxIndex => Channel}

	Status                uint8
	Decision              uint8
	ConfirmedTransactions []TxIndex // [TxIndex]
	Acks                  []TxIndex // [TxIndex]
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
	ContractCallInfo        []byte `json:"contract_call_info" yaml:"contract_call_info"`
}

func NewTxInfo(status, prepareResult uint8, coordinatorConnectionID string, contractCallInfo []byte) TxInfo {
	return TxInfo{Status: status, PrepareResult: prepareResult, CoordinatorConnectionID: coordinatorConnectionID, ContractCallInfo: contractCallInfo}
}

type ContractCallResult struct {
	ChainID         string           `json:"chain_id" yaml:"chain_id"`
	Height          int64            `json:"height" yaml:"height"`
	Signers         []sdk.AccAddress `json:"signers" yaml:"signers"`
	CallInfo        ContractCallInfo `json:"call_info" yaml:"call_info"`
	StateConstraint StateConstraint  `json:"state_constraint" yaml:"state_constraint"`
}

func (r ContractCallResult) String() string {
	// TODO make this more readable
	return fmt.Sprintf("%#v", r)
}
