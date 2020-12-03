package types

import (
	"errors"
	"fmt"

	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
)

// NewCoordinatorState creates a new instance of CoordinatorState
func NewCoordinatorState(cp txtypes.CommitProtocol, phase CoordinatorPhase, channels []xcctypes.ChannelInfo) CoordinatorState {
	if len(channels) == 0 {
		panic("channels must not be empty")
	}
	return CoordinatorState{
		Type:     cp,
		Phase:    phase,
		Channels: channels,
	}
}

// Confirm append a given txIndex to confirmedTxs if it doesn't exist
func (cs *CoordinatorState) Confirm(txIndex txtypes.TxIndex, channel xcctypes.ChannelInfo) error {
	for _, id := range cs.ConfirmedTxs {
		if txIndex == id {
			return errors.New("this tx is already confirmed")
		}
	}

	if int(txIndex) >= len(cs.Channels) {
		return fmt.Errorf("txIndex '%v' not found", txIndex)
	} else if c := cs.Channels[txIndex]; c != channel {
		return fmt.Errorf("expected channel is '%v', but got '%v'", c, channel)
	}

	cs.ConfirmedTxs = append(cs.ConfirmedTxs, txIndex)
	return nil
}

// IsConfirmedALLPrepares returns a boolean value whether all txs are confirmed
func (cs CoordinatorState) IsConfirmedALLPrepares() bool {
	return len(cs.Channels) == len(cs.ConfirmedTxs)
}

// AddAck adds txIndex to Acks if it doesn't exist
func (cs *CoordinatorState) AddAck(txIndex txtypes.TxIndex) bool {
	for _, id := range cs.Acks {
		if txIndex == id {
			return false
		}
	}
	cs.Acks = append(cs.Acks, txIndex)
	return true
}

// IsConfirmedALLCommits returns a boolean whether all acks are received
func (cs *CoordinatorState) IsConfirmedALLCommits() bool {
	return len(cs.Channels) == len(cs.Acks)
}

// NewContractTransactionState creates a new instance of ContractTransactionState
func NewContractTransactionState(status ContractTransactionStatus, prepareResult PrepareResult, coordinatorChannel xcctypes.ChannelInfo) ContractTransactionState {
	return ContractTransactionState{
		Status:             status,
		PrepareResult:      prepareResult,
		CoordinatorChannel: coordinatorChannel,
	}
}
