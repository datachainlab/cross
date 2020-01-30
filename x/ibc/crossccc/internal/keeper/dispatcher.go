package keeper

import (
	"github.com/bluele/crossccc/x/ibc/crossccc/internal/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

/*
should return Store(Store?) instead of StoreKeys
*/
type ContractHandler interface {
	Handle(ctx sdk.Context, contract []byte) (types.State, error)
	GetState(ctx sdk.Context, contract []byte) (types.State, error)
}
