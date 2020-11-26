package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type ContractManager interface {
	PrepareCommit(
		ctx sdk.Context,
		txID TxID,
		txIndex TxIndex,
		tx ResolvedContractTransaction,
	) error
	CommitImmediately(
		ctx sdk.Context,
		txID TxID,
		txIndex TxIndex,
		tx ResolvedContractTransaction,
	) (*ContractCallResult, error)
	Commit(
		ctx sdk.Context,
		txID TxID,
		txIndex TxIndex,
	) (*ContractCallResult, error)
	Abort(
		ctx sdk.Context,
		txID TxID,
		txIndex TxIndex,
	) error
}

// GetEvents converts Events to sdk.Events
func (res ContractCallResult) GetEvents() sdk.Events {
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
