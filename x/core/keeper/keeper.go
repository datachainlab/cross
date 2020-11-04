package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	commonkeeper "github.com/datachainlab/cross/x/atomic/common/keeper"
	simplekeeper "github.com/datachainlab/cross/x/atomic/simple/keeper"
	tpckeeper "github.com/datachainlab/cross/x/atomic/tpc/keeper"
	"github.com/datachainlab/cross/x/core/types"
	"github.com/datachainlab/cross/x/packets"
)

type Keeper struct {
	m                codec.Marshaler
	storeKey         sdk.StoreKey
	portKeeper       types.PortKeeper
	scopedKeeper     capabilitykeeper.ScopedKeeper
	packetMiddleware packets.PacketMiddleware

	simpleKeeper simplekeeper.Keeper
	tpcKeeper    tpckeeper.Keeper
	commonkeeper.Keeper
}

// NewKeeper creates a new instance of Cross Keeper
func NewKeeper(
	m codec.Marshaler,
	storeKey sdk.StoreKey,
	channelKeeper types.ChannelKeeper,
	portKeeper types.PortKeeper,
	scopedKeeper capabilitykeeper.ScopedKeeper,
	packetMiddleware packets.PacketMiddleware,
	contractHandler types.ContractHandler,
	commitStore types.CommitStore,
) Keeper {
	ck := commonkeeper.NewKeeper(m, storeKey, types.KeyAtomicKeeperPrefixBytes(), channelKeeper, portKeeper, scopedKeeper, contractHandler, commitStore)
	return Keeper{
		m:                m,
		storeKey:         storeKey,
		portKeeper:       portKeeper,
		scopedKeeper:     scopedKeeper,
		packetMiddleware: packetMiddleware,

		simpleKeeper: simplekeeper.NewKeeper(m, ck),
		tpcKeeper:    tpckeeper.NewKeeper(),
		Keeper:       ck,
	}
}

// SimpleKeeper returns the simple commit keeper
func (k Keeper) SimpleKeeper() simplekeeper.Keeper {
	return k.simpleKeeper
}

// TPCKeeper returns the two-phase commit keeper
func (k Keeper) TPCKeeper() tpckeeper.Keeper {
	return k.tpcKeeper
}

// Logger returns a logger instance
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s-%s", host.ModuleName, types.ModuleName))
}

// GetPort returns portID
func (k Keeper) GetPort(ctx sdk.Context) string {
	return types.PortID
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
