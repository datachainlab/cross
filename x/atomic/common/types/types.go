package types

import (
	"errors"
	"fmt"

	"github.com/datachainlab/cross/x/core/types"
)

// NewCoordinatorState creates a new instance of CoordinatorState
func NewCoordinatorState(commitFlowType CommitFlowType, phase CoordinatorPhase, channels []types.ChannelInfo) CoordinatorState {
	return CoordinatorState{
		Type:     commitFlowType,
		Phase:    phase,
		Channels: channels,
	}
}

// Confirm ...
func (cs *CoordinatorState) Confirm(txIndex types.TxIndex, channel types.ChannelInfo) error {
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

// NewContractTransactionState creates a new instance of ContractTransactionState
func NewContractTransactionState(status ContractTransactionStatus, prepareResult PrepareResult, coordinatorChannel types.ChannelInfo) ContractTransactionState {
	return ContractTransactionState{
		Status:             status,
		PrepareResult:      prepareResult,
		CoordinatorChannel: coordinatorChannel,
	}
}
