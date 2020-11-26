package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

type Keeper struct {
	cdc codec.Marshaler

	// atomickeeper.Keeper
}

func NewKeeper() Keeper {
	return Keeper{}
}
