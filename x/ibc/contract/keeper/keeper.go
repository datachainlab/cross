package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/ibc/contract/types"
	"github.com/datachainlab/cross/x/ibc/cross"
)

type Keeper struct {
	cdc      *codec.Codec
	storeKey sdk.StoreKey
}

func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey) Keeper {
	return Keeper{cdc: cdc, storeKey: storeKey}
}

func (k Keeper) GetContractStateStore(ctx sdk.Context, id []byte) sdk.KVStore {
	st := ctx.KVStore(k.storeKey)
	return prefix.NewStore(st, id)
}

func (k Keeper) MakeContractCallResponseData(rv []byte, ops cross.OPs) ([]byte, error) {
	return k.cdc.MarshalJSON(types.ContractCallResponse{ReturnValue: rv, OPs: ops})
}
