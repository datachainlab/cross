package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/core/atomic/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
)

// SetCoordinatorState sets the store to a CoordinatorState
func (k Keeper) SetCoordinatorState(ctx sdk.Context, txID crosstypes.TxID, cs types.CoordinatorState) {
	bz := k.cdc.MustMarshal(&cs)
	k.store(ctx).Set(types.KeyCoordinatorState(txID), bz)
}

// GetCoordinatorState returns the CoordinatorState of a given txID
func (k Keeper) GetCoordinatorState(ctx sdk.Context, txID crosstypes.TxID) (*types.CoordinatorState, bool) {
	bz := k.store(ctx).Get(types.KeyCoordinatorState(txID))
	if bz == nil {
		return nil, false
	}
	var cs types.CoordinatorState
	k.cdc.MustUnmarshal(bz, &cs)
	return &cs, true
}

// TODO use channelInfo to create a key
// SetContractTransactionState sets the store to a ContractTransactionState
func (k Keeper) SetContractTransactionState(ctx sdk.Context, txID crosstypes.TxID, txIndex crosstypes.TxIndex, txState types.ContractTransactionState) {
	bz := k.cdc.MustMarshal(&txState)
	k.store(ctx).Set(types.KeyContractTransactionState(txID, txIndex), bz)
}

// GetContractTransactionState returns the GetContractTransactionState of a given txID and txIndex
func (k Keeper) GetContractTransactionState(ctx sdk.Context, txID crosstypes.TxID, txIndex crosstypes.TxIndex) (*types.ContractTransactionState, bool) {
	bz := k.store(ctx).Get(types.KeyContractTransactionState(txID, txIndex))
	if bz == nil {
		return nil, false
	}
	var txState types.ContractTransactionState
	k.cdc.MustUnmarshal(bz, &txState)
	return &txState, true
}

// EnsureContractTransactionStatus ensures that the status of the tx equals a given status
func (k Keeper) EnsureContractTransactionStatus(ctx sdk.Context, txID crosstypes.TxID, txIndex crosstypes.TxIndex, status types.ContractTransactionStatus) (*types.ContractTransactionState, error) {
	txState, found := k.GetContractTransactionState(ctx, txID, txIndex)
	if !found {
		return nil, fmt.Errorf("(txID, txIndex) = ('%x', '%v') not found", txID, txIndex)
	}
	if txState.Status == status {
		return txState, nil
	} else {
		return nil, fmt.Errorf("expected status is %v, but got %v", status, txState.Status)
	}
}

// UpdateContractTransactionStatus updates the status of the tx
func (k Keeper) UpdateContractTransactionStatus(ctx sdk.Context, txID crosstypes.TxID, txIndex crosstypes.TxIndex, status types.ContractTransactionStatus) error {
	txState, found := k.GetContractTransactionState(ctx, txID, txIndex)
	if !found {
		return fmt.Errorf("(txID, txIndex) = ('%x', '%v') not found", txID, txIndex)
	}
	txState.Status = status
	k.SetContractTransactionState(ctx, txID, txIndex, *txState)
	return nil
}

func (k Keeper) store(ctx sdk.Context) sdk.KVStore {
	switch storeKey := k.storeKey.(type) {
	case *crosstypes.PrefixStoreKey:
		return prefix.NewStore(ctx.KVStore(storeKey.StoreKey), storeKey.Prefix)
	default:
		return ctx.KVStore(k.storeKey)
	}
}
