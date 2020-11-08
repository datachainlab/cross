package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/datachainlab/cross/x/atomic/common/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
)

// TODO use channelInfo to create a key
// SetContractCallResult sets the store to a result
func (k Keeper) SetContractCallResult(ctx sdk.Context, txID crosstypes.TxID, txIndex crosstypes.TxIndex, result crosstypes.ContractCallResult) {
	bz := k.cdc.MustMarshalBinaryBare(&result)
	k.store(ctx).Set(types.KeyContractCallResult(txID, txIndex), bz)
}

// GetContractCallResult returns the result of a given txID and txIndex
func (k Keeper) GetContractCallResult(ctx sdk.Context, txID crosstypes.TxID, txIndex crosstypes.TxIndex) *crosstypes.ContractCallResult {
	bz := k.store(ctx).Get(types.KeyContractCallResult(txID, txIndex))
	if bz == nil {
		return nil
	}
	var result crosstypes.ContractCallResult
	k.cdc.MustUnmarshalBinaryBare(bz, &result)
	return &result
}

// RemoveContractCallResult removes the result from store
func (k Keeper) RemoveContractCallResult(ctx sdk.Context, txID crosstypes.TxID, txIndex crosstypes.TxIndex) {
	k.store(ctx).Delete(types.KeyContractCallResult(txID, txIndex))
}

// SetCoordinatorState sets the store to a CoordinatorState
func (k Keeper) SetCoordinatorState(ctx sdk.Context, txID crosstypes.TxID, cs types.CoordinatorState) {
	bz := k.cdc.MustMarshalBinaryBare(&cs)
	k.store(ctx).Set(types.KeyCoordinatorState(txID), bz)
}

// GetCoordinatorState returns the CoordinatorState of a given txID
func (k Keeper) GetCoordinatorState(ctx sdk.Context, txID crosstypes.TxID) (*types.CoordinatorState, bool) {
	bz := k.store(ctx).Get(types.KeyCoordinatorState(txID))
	if bz == nil {
		return nil, false
	}
	var cs types.CoordinatorState
	k.cdc.MustUnmarshalBinaryBare(bz, &cs)
	return &cs, true
}

// TODO use channelInfo to create a key
// SetContractTransactionState sets the store to a ContractTransactionState
func (k Keeper) SetContractTransactionState(ctx sdk.Context, txID crosstypes.TxID, txIndex crosstypes.TxIndex, txState types.ContractTransactionState) {
	bz := k.cdc.MustMarshalBinaryBare(&txState)
	k.store(ctx).Set(types.KeyContractTransactionState(txID, txIndex), bz)
}

// GetContractTransactionState returns the GetContractTransactionState of a given txID and txIndex
func (k Keeper) GetContractTransactionState(ctx sdk.Context, txID crosstypes.TxID, txIndex crosstypes.TxIndex) (*types.ContractTransactionState, bool) {
	bz := k.store(ctx).Get(types.KeyContractTransactionState(txID, txIndex))
	if bz == nil {
		return nil, false
	}
	var txState types.ContractTransactionState
	k.cdc.MustUnmarshalBinaryBare(bz, &txState)
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
	return prefix.NewStore(ctx.KVStore(k.storeKey), k.keyPrefix)
}
