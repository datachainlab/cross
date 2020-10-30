package common

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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
