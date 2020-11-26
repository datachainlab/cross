package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/datachainlab/cross/x/atomic/types"
	contracttypes "github.com/datachainlab/cross/x/contract/types"
	"github.com/datachainlab/cross/x/core/host"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
)

// TODO use channelInfo to create a key
// setContractCallResult sets the store to a result
func (k contractManager) setContractCallResult(ctx sdk.Context, txID txtypes.TxID, txIndex txtypes.TxIndex, result contracttypes.ContractCallResult) {
	bz := k.cdc.MustMarshalBinaryBare(&result)
	k.store(ctx).Set(types.KeyContractCallResult(txID, txIndex), bz)
}

// getContractCallResult returns the result of a given txID and txIndex
func (k contractManager) getContractCallResult(ctx sdk.Context, txID txtypes.TxID, txIndex txtypes.TxIndex) *contracttypes.ContractCallResult {
	bz := k.store(ctx).Get(types.KeyContractCallResult(txID, txIndex))
	if bz == nil {
		return nil
	}
	var result contracttypes.ContractCallResult
	k.cdc.MustUnmarshalBinaryBare(bz, &result)
	return &result
}

// removeContractCallResult removes the result from store
func (k contractManager) removeContractCallResult(ctx sdk.Context, txID txtypes.TxID, txIndex txtypes.TxIndex) {
	k.store(ctx).Delete(types.KeyContractCallResult(txID, txIndex))
}

func (k contractManager) store(ctx sdk.Context) sdk.KVStore {
	switch storeKey := k.storeKey.(type) {
	case *host.PrefixStoreKey:
		return prefix.NewStore(ctx.KVStore(storeKey.StoreKey), storeKey.Prefix)
	default:
		return ctx.KVStore(k.storeKey)
	}
}
