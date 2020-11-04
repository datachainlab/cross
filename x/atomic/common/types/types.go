package types

import (
	"errors"
	"fmt"

	crosstypes "github.com/datachainlab/cross/x/core/types"
)

// NewCoordinatorState creates a new instance of CoordinatorState
func NewCoordinatorState(commitFlowType CommitFlowType, phase CoordinatorPhase, channels []crosstypes.ChannelInfo) CoordinatorState {
	return CoordinatorState{
		Type:     commitFlowType,
		Phase:    phase,
		Channels: channels,
	}
}

// Confirm ...
func (cs *CoordinatorState) Confirm(txIndex crosstypes.TxIndex, channel crosstypes.ChannelInfo) error {
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
func NewContractTransactionState(status ContractTransactionStatus, prepareResult PrepareResult, coordinatorChannel crosstypes.ChannelInfo) ContractTransactionState {
	return ContractTransactionState{
		Status:             status,
		PrepareResult:      prepareResult,
		CoordinatorChannel: coordinatorChannel,
	}
}
