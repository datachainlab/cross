package keeper

import (
	"github.com/bluele/cross/x/ibc/cross/internal/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

/* TODO
should return Store(Store?) instead of StoreKeys
*/
type ContractHandler interface {
	Handle(ctx sdk.Context, contract []byte) (types.State, error)
	GetState(ctx sdk.Context, contract []byte) (types.State, error)
}
