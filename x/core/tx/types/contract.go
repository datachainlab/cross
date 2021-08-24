package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
)

type ContractManager interface {
	PrepareCommit(
		ctx sdk.Context,
		txID crosstypes.TxID,
		txIndex crosstypes.TxIndex,
		tx ResolvedContractTransaction,
	) (*ContractCallResult, error)
	CommitImmediately(
		ctx sdk.Context,
		txID crosstypes.TxID,
		txIndex crosstypes.TxIndex,
		tx ResolvedContractTransaction,
	) (*ContractCallResult, error)
	Commit(
		ctx sdk.Context,
		txID crosstypes.TxID,
		txIndex crosstypes.TxIndex,
	) (*ContractCallResult, error)
	Abort(
		ctx sdk.Context,
		txID crosstypes.TxID,
		txIndex crosstypes.TxIndex,
	) error
}

// GetData returns Data
func (res *ContractCallResult) GetData() []byte {
	if res == nil {
		return nil
	} else {
		return res.Data
	}
}

// GetEvents converts Events to sdk.Events
func (res *ContractCallResult) GetEvents() sdk.Events {
	if res == nil {
		return nil
	}
	events := make(sdk.Events, 0, len(res.Events))
	for _, ev := range res.Events {
		attrs := make([]sdk.Attribute, 0, len(ev.Attributes))
		for _, attr := range ev.Attributes {
			attrs = append(attrs, sdk.NewAttribute(string(attr.Key), string(attr.Value)))
		}
		events = append(events, sdk.NewEvent(ev.Type, attrs...))
	}
	return events
}
