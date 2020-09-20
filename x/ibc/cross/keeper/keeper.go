package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	porttypes "github.com/cosmos/cosmos-sdk/x/ibc/05-port/types"
	"github.com/datachainlab/cross/x/ibc/cross/keeper/common"
	"github.com/datachainlab/cross/x/ibc/cross/keeper/simple"
	"github.com/datachainlab/cross/x/ibc/cross/keeper/tpc"
	"github.com/datachainlab/cross/x/ibc/cross/types"
)

// Keeper maintains the link to storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	cdc      *codec.Codec // The wire codec for binary encoding/decoding.
	storeKey sdk.StoreKey // Unexposed key to access store from sdk.Context

	simpleKeeper simple.Keeper
	tpcKeeper    tpc.Keeper
	portKeeper   types.PortKeeper
	scopedKeeper capability.ScopedKeeper

	common.Keeper
}

// NewKeeper creates new instances of the cross Keeper
func NewKeeper(
	cdc *codec.Codec,
	storeKey sdk.StoreKey,
	channelKeeper types.ChannelKeeper,
	portKeeper types.PortKeeper,
	scopedKeeper capability.ScopedKeeper,
	resolverProvider types.ObjectResolverProvider,
	channelResolver types.ChannelResolver,
) Keeper {
	ck := common.NewKeeper(cdc, storeKey, channelKeeper, portKeeper, scopedKeeper, resolverProvider, channelResolver)
	return Keeper{
		cdc:          cdc,
		storeKey:     storeKey,
		simpleKeeper: simple.NewKeeper(cdc, storeKey, ck),
		tpcKeeper:    tpc.NewKeeper(cdc, storeKey, ck),
		portKeeper:   portKeeper,
		scopedKeeper: scopedKeeper,
		Keeper:       ck,
	}
}

func (k Keeper) TPCKeeper() tpc.Keeper {
	return k.tpcKeeper
}

func (k Keeper) SimpleKeeper() simple.Keeper {
	return k.simpleKeeper
}

// BindPort defines a wrapper function for the ort TPCKeeper's function in
// order to expose it to module's InitGenesis function
func (k Keeper) BindPort(ctx sdk.Context, portID string) (*capability.Capability, error) {
	cap := k.portKeeper.BindPort(ctx, portID)
	if err := k.ClaimCapability(ctx, cap, porttypes.PortPath(portID)); err != nil {
		return nil, err
	}
	return cap, nil
}

// ClaimCapability allows the transfer module that can claim a capability that IBC module
// passes to it
func (k Keeper) ClaimCapability(ctx sdk.Context, cap *capability.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}
