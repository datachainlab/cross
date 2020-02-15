package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
	storeKey sdk.StoreKey
}

func NewKeeper(storeKey sdk.StoreKey) Keeper {
	return Keeper{storeKey: storeKey}
}

func (k Keeper) GetContractStateStore(ctx sdk.Context, id []byte) sdk.KVStore {
	st := ctx.KVStore(k.storeKey)
	return prefix.NewStore(st, id)
}
