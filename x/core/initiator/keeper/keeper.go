package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	crosshost "github.com/datachainlab/cross/x/core/host"
	"github.com/datachainlab/cross/x/core/initiator/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	"github.com/datachainlab/cross/x/packets"
	xcctypes "github.com/datachainlab/cross/x/xcc/types"
)

type Keeper struct {
	m                codec.Marshaler
	storeKey         sdk.StoreKey
	portKeeper       types.PortKeeper
	channelKeeper    types.ChannelKeeper
	scopedKeeper     capabilitykeeper.ScopedKeeper
	packetMiddleware packets.PacketMiddleware
	xccResolver      xcctypes.XCCResolver
	txProcessor      txtypes.TxProcessor
	packets.PacketSendKeeper
}

// NewKeeper creates a new instance of Cross Keeper
func NewKeeper(
	m codec.Marshaler,
	storeKey sdk.StoreKey,
	channelKeeper types.ChannelKeeper,
	portKeeper types.PortKeeper,
	scopedKeeper capabilitykeeper.ScopedKeeper,
	packetMiddleware packets.PacketMiddleware,
	xccResolver xcctypes.XCCResolver,
	txProcessor txtypes.TxProcessor,
) Keeper {
	psk := packets.NewPacketSendKeeper(
		m,
		channelKeeper,
		portKeeper,
		scopedKeeper,
	)
	return Keeper{
		m:                m,
		storeKey:         storeKey,
		portKeeper:       portKeeper,
		channelKeeper:    channelKeeper,
		scopedKeeper:     scopedKeeper,
		packetMiddleware: packetMiddleware,
		xccResolver:      xccResolver,
		txProcessor:      txProcessor,
		PacketSendKeeper: psk,
	}
}

func (k Keeper) ChannelKeeper() types.ChannelKeeper {
	return k.channelKeeper
}

func (k Keeper) CrossChainChannelResolver() xcctypes.XCCResolver {
	return k.xccResolver
}

// Logger returns a logger instance
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s-%s", host.ModuleName, types.SubModuleName))
}

// BindPort defines a wrapper function for the ort Keeper's function in
// order to expose it to module's InitGenesis function
func (k Keeper) BindPort(ctx sdk.Context, portID string) error {
	cap := k.portKeeper.BindPort(ctx, portID)
	return k.ClaimCapability(ctx, cap, host.PortPath(portID))
}

// ClaimCapability allows the transfer module that can claim a capability that IBC module
// passes to it
func (k Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}

// IsBound checks if the transfer module is already bound to the desired port
func (k Keeper) IsBound(ctx sdk.Context, portID string) bool {
	_, ok := k.scopedKeeper.GetCapability(ctx, host.PortPath(portID))
	return ok
}

func (k Keeper) store(ctx sdk.Context) sdk.KVStore {
	switch storeKey := k.storeKey.(type) {
	case *crosshost.PrefixStoreKey:
		return prefix.NewStore(ctx.KVStore(storeKey.StoreKey), storeKey.Prefix)
	default:
		return ctx.KVStore(k.storeKey)
	}
}
