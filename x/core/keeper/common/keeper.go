package common

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	"github.com/datachainlab/cross/x/core/types"
)

type Keeper struct {
	cdc      codec.Marshaler
	storeKey sdk.StoreKey

	channelKeeper types.ChannelKeeper
	portKeeper    types.PortKeeper
	scopedKeeper  capabilitykeeper.ScopedKeeper

	// resolverProvider types.ObjectResolverProvider
	// channelResolver  types.ChannelResolver
}

func NewKeeper(
	cdc codec.Marshaler,
	storeKey sdk.StoreKey,
	channelKeeper types.ChannelKeeper,
	portKeeper types.PortKeeper,
	scopedKeeper capabilitykeeper.ScopedKeeper,
) Keeper {
	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		channelKeeper: channelKeeper,
		portKeeper:    portKeeper,
		scopedKeeper:  scopedKeeper,
	}
}

func (k Keeper) ChannelKeeper() types.ChannelKeeper {
	return k.channelKeeper
}
