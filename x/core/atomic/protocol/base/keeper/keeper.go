package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	"github.com/datachainlab/cross/x/core/atomic/protocol/base/types"
	"github.com/datachainlab/cross/x/packets"
)

type Keeper struct {
	cdc      codec.Marshaler
	storeKey sdk.StoreKey

	channelKeeper types.ChannelKeeper
	packets.PacketSendKeeper
}

func NewKeeper(
	cdc codec.Marshaler,
	storeKey sdk.StoreKey,
	channelKeeper types.ChannelKeeper,
	portKeeper types.PortKeeper,
	scopedKeeper capabilitykeeper.ScopedKeeper,
) Keeper {
	psk := packets.NewPacketSendKeeper(cdc, channelKeeper, portKeeper, scopedKeeper)
	return Keeper{
		cdc:              cdc,
		storeKey:         storeKey,
		PacketSendKeeper: psk,
		channelKeeper:    channelKeeper,
	}
}

func (k Keeper) ChannelKeeper() types.ChannelKeeper {
	return k.channelKeeper
}
