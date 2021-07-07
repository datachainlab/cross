package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"

	authtypes "github.com/datachainlab/cross/x/core/auth/types"
	"github.com/datachainlab/cross/x/core/initiator/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
	"github.com/datachainlab/cross/x/packets"
)

type Keeper struct {
	m                codec.Codec
	storeKey         sdk.StoreKey
	portKeeper       types.PortKeeper
	channelKeeper    types.ChannelKeeper
	scopedKeeper     capabilitykeeper.ScopedKeeper
	packetMiddleware packets.PacketMiddleware
	xccResolver      xcctypes.XCCResolver
	authenticator    authtypes.TxAuthenticator
	txRunner         txtypes.TxRunner
	packets.PacketSendKeeper
}

// NewKeeper creates a new instance of Cross Keeper
func NewKeeper(
	m codec.Codec,
	storeKey sdk.StoreKey,
	channelKeeper types.ChannelKeeper,
	authenticator authtypes.TxAuthenticator,
	packetSendKeeper packets.PacketSendKeeper,
	packetMiddleware packets.PacketMiddleware,
	xccResolver xcctypes.XCCResolver,
	txRunner txtypes.TxRunner,
) Keeper {
	return Keeper{
		m:                m,
		storeKey:         storeKey,
		channelKeeper:    channelKeeper,
		authenticator:    authenticator,
		packetMiddleware: packetMiddleware,
		xccResolver:      xccResolver,
		txRunner:         txRunner,
		PacketSendKeeper: packetSendKeeper,
	}
}

func (k Keeper) ChannelKeeper() types.ChannelKeeper {
	return k.channelKeeper
}

// Logger returns a logger instance
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s-%s", crosstypes.ModuleName, types.SubModuleName))
}

func (k Keeper) store(ctx sdk.Context) sdk.KVStore {
	switch storeKey := k.storeKey.(type) {
	case *crosstypes.PrefixStoreKey:
		return prefix.NewStore(ctx.KVStore(storeKey.StoreKey), storeKey.Prefix)
	default:
		return ctx.KVStore(k.storeKey)
	}
}
