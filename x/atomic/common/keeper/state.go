package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/datachainlab/cross/x/atomic/common/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
)

func (k Keeper) SetContractResult(ctx sdk.Context, txID crosstypes.TxID, txIndex crosstypes.TxIndex, result crosstypes.ContractHandlerResult) {
	bz := k.cdc.MustMarshalBinaryBare(&result)
	k.store(ctx).Set(types.KeyContractResult(txID, txIndex), bz)
}

func (k Keeper) GetContractResult(ctx sdk.Context, txID crosstypes.TxID, txIndex crosstypes.TxIndex) *crosstypes.ContractHandlerResult {
	bz := k.store(ctx).Get(types.KeyContractResult(txID, txIndex))
	if bz == nil {
		return nil
	}
	var result crosstypes.ContractHandlerResult
	k.cdc.MustUnmarshalBinaryBare(bz, &result)
	return &result
}

func (k Keeper) RemoveContractResult(ctx sdk.Context, txID crosstypes.TxID, txIndex crosstypes.TxIndex) {
	k.store(ctx).Delete(types.KeyContractResult(txID, txIndex))
}

func (k Keeper) SetCoordinatorState(ctx sdk.Context, txID crosstypes.TxID, cs types.CoordinatorState) {
	bz := k.cdc.MustMarshalBinaryBare(&cs)
	k.store(ctx).Set(types.KeyCoordinatorState(txID), bz)
}

func (k Keeper) GetCoordinatorState(ctx sdk.Context, txID crosstypes.TxID) (*types.CoordinatorState, bool) {
	bz := k.store(ctx).Get(types.KeyCoordinatorState(txID))
	if bz == nil {
		return nil, false
	}
	var cs types.CoordinatorState
	k.cdc.MustUnmarshalBinaryBare(bz, &cs)
	return &cs, true
}

func (k Keeper) SetContractTransactionState(ctx sdk.Context, txID crosstypes.TxID, txIndex crosstypes.TxIndex, txState types.ContractTransactionState) {
	bz := k.cdc.MustMarshalBinaryBare(&txState)
	k.store(ctx).Set(types.KeyContractTransactionState(txID, txIndex), bz)
}

func (k Keeper) GetContractTransactionState(ctx sdk.Context, txID crosstypes.TxID, txIndex crosstypes.TxIndex) (*types.ContractTransactionState, bool) {
	bz := k.store(ctx).Get(types.KeyContractTransactionState(txID, txIndex))
	if bz == nil {
		return nil, false
	}
	var txState types.ContractTransactionState
	k.cdc.MustUnmarshalBinaryBare(bz, &txState)
	return &txState, true
}

func (k Keeper) store(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), k.keyPrefix)
}
