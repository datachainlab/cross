package naive

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/ibc/cross/keeper/common"
	"github.com/datachainlab/cross/x/ibc/cross/types"
)

const TypeName = "naive"

type Keeper struct {
	cdc      *codec.Codec // The wire codec for binary encoding/decoding.
	storeKey sdk.StoreKey // Unexposed key to access store from sdk.Context

	resolverProvider types.ObjectResolverProvider
	common.Keeper
}

func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, resolverProvider types.ObjectResolverProvider, ck common.Keeper) Keeper {
	return Keeper{
		cdc:              cdc,
		storeKey:         storeKey,
		resolverProvider: resolverProvider,
		Keeper:           ck,
	}
}
