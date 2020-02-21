package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/ibc/cross/internal/types"
)

/* TODO
should return Store(Store?) instead of StoreKeys
*/
type ContractHandler interface {
	Handle(ctx sdk.Context, contract []byte) (types.State, error)
	GetState(ctx sdk.Context, contract []byte) (types.State, error)
}
