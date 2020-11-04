package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/datachainlab/cross/x/atomic/common/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
)

func (k Keeper) SetContractResult(ctx sdk.Context, txID crosstypes.TxID, txIndex crosstypes.TxIndex, result crosstypes.ContractHandlerResult) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(&result)
	store.Set(crosstypes.KeyContractResult(txID, txIndex), bz)
}

func (k Keeper) GetContractResult(ctx sdk.Context, txID crosstypes.TxID, txIndex crosstypes.TxIndex) *crosstypes.ContractHandlerResult {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(crosstypes.KeyContractResult(txID, txIndex))
	if bz == nil {
		return nil
	}
	var result crosstypes.ContractHandlerResult
	k.cdc.MustUnmarshalBinaryBare(bz, &result)
	return &result
}

func (k Keeper) RemoveContractResult(ctx sdk.Context, txID crosstypes.TxID, txIndex crosstypes.TxIndex) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(crosstypes.KeyContractResult(txID, txIndex))
}

func (k Keeper) SetCoordinatorState(ctx sdk.Context, txID crosstypes.TxID, cs types.CoordinatorState) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(&cs)
	store.Set(crosstypes.KeyCoordinatorState(txID), bz)
}

func (k Keeper) GetCoordinatorState(ctx sdk.Context, txID crosstypes.TxID) (*types.CoordinatorState, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(crosstypes.KeyCoordinatorState(txID))
	if bz == nil {
		return nil, false
	}
	var cs types.CoordinatorState
	k.cdc.MustUnmarshalBinaryBare(bz, &cs)
	return &cs, true
}

func (k Keeper) SetContractTransactionState(ctx sdk.Context, txID crosstypes.TxID, txIndex crosstypes.TxIndex, txState types.ContractTransactionState) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(&txState)
	store.Set(crosstypes.KeyContractTransactionState(txID, txIndex), bz)
}

func (k Keeper) GetContractTransactionState(ctx sdk.Context, txID crosstypes.TxID, txIndex crosstypes.TxIndex) (*types.ContractTransactionState, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(crosstypes.KeyContractTransactionState(txID, txIndex))
	if bz == nil {
		return nil, false
	}
	var txState types.ContractTransactionState
	k.cdc.MustUnmarshalBinaryBare(bz, &txState)
	return &txState, true
}
