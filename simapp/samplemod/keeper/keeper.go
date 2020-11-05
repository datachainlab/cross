package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
)

type Keeper struct {
	storeKey sdk.StoreKey
	xstore   crosstypes.Store
}

func NewKeeper(storeKey sdk.StoreKey, xstore crosstypes.Store) Keeper {
	return Keeper{
		storeKey: storeKey,
		xstore:   xstore,
	}
}

func (k Keeper) HandleContractCall(ctx context.Context, callInfo crosstypes.ContractCallInfo) (*crosstypes.ContractCallResult, *crosstypes.OPs, error) {
	panic("not implemented error")
}
