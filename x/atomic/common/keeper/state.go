package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	commontypes "github.com/datachainlab/cross/x/atomic/common/types"
	"github.com/datachainlab/cross/x/core/types"
)

func (k Keeper) SetContractResult(ctx sdk.Context, txID types.TxID, txIndex types.TxIndex, result types.ContractHandlerResult) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(&result)
	store.Set(types.KeyContractResult(txID, txIndex), bz)
}

func (k Keeper) GetContractResult(ctx sdk.Context, txID types.TxID, txIndex types.TxIndex) *types.ContractHandlerResult {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyContractResult(txID, txIndex))
	if bz == nil {
		return nil
	}
	var result types.ContractHandlerResult
	k.cdc.MustUnmarshalBinaryBare(bz, &result)
	return &result
}

func (k Keeper) RemoveContractResult(ctx sdk.Context, txID types.TxID, txIndex types.TxIndex) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyContractResult(txID, txIndex))
}

func (k Keeper) SetCoordinatorState(ctx sdk.Context, txID types.TxID, cs commontypes.CoordinatorState) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(&cs)
	store.Set(types.KeyCoordinatorState(txID), bz)
}

func (k Keeper) GetCoordinatorState(ctx sdk.Context, txID types.TxID) (*commontypes.CoordinatorState, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyCoordinatorState(txID))
	if bz == nil {
		return nil, false
	}
	var cs commontypes.CoordinatorState
	k.cdc.MustUnmarshalBinaryBare(bz, &cs)
	return &cs, true
}

func (k Keeper) SetContractTransactionState(ctx sdk.Context, txID types.TxID, txIndex types.TxIndex, txState commontypes.ContractTransactionState) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(&txState)
	store.Set(types.KeyContractTransactionState(txID, txIndex), bz)
}

func (k Keeper) GetContractTransactionState(ctx sdk.Context, txID types.TxID, txIndex types.TxIndex) (*commontypes.ContractTransactionState, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyContractTransactionState(txID, txIndex))
	if bz == nil {
		return nil, false
	}
	var txState commontypes.ContractTransactionState
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &txState)
	return &txState, true
}
