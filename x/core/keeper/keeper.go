package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"github.com/datachainlab/cross/x/core/keeper/common"
	"github.com/datachainlab/cross/x/core/keeper/simple"
	"github.com/datachainlab/cross/x/core/keeper/tpc"
	"github.com/datachainlab/cross/x/core/types"
)

type Keeper struct {
	m            codec.Marshaler
	storeKey     sdk.StoreKey
	portKeeper   types.PortKeeper
	scopedKeeper capabilitykeeper.ScopedKeeper

	simpleKeeper simple.Keeper
	tpcKeeper    tpc.Keeper
	common.Keeper
}

// NewKeeper creates a new instance of Cross Keeper
func NewKeeper(
	m codec.Marshaler,
	storeKey sdk.StoreKey,
	channelKeeper types.ChannelKeeper,
	portKeeper types.PortKeeper,
	scopedKeeper capabilitykeeper.ScopedKeeper,
) Keeper {
	// TODO set fields to values
	return Keeper{}
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
