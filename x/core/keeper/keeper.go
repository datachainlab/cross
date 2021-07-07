package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	host "github.com/cosmos/ibc-go/modules/core/24-host"

	authkeeper "github.com/datachainlab/cross/x/core/auth/keeper"
	authtypes "github.com/datachainlab/cross/x/core/auth/types"
	initiatorkeeper "github.com/datachainlab/cross/x/core/initiator/keeper"
	"github.com/datachainlab/cross/x/core/router"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	"github.com/datachainlab/cross/x/core/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
	"github.com/datachainlab/cross/x/packets"
)

type Keeper struct {
	m             codec.Codec
	portKeeper    types.PortKeeper
	channelKeeper types.ChannelKeeper
	scopedKeeper  capabilitykeeper.ScopedKeeper

	router          router.Router
	initiatorKeeper initiatorkeeper.Keeper
	authKeeper      authkeeper.Keeper
}

func NewKeeper(
	cdc codec.Codec, initiatorStoreKey, authStoreKey sdk.StoreKey,
	channelKeeper types.ChannelKeeper, portKeeper types.PortKeeper, scopedKeeper capabilitykeeper.ScopedKeeper,
	packetMiddleware packets.PacketMiddleware, xccResolver xcctypes.XCCResolver, txRunner txtypes.TxRunner, router router.Router,
) Keeper {
	packetSendKeeper := packets.NewPacketSendKeeper(
		cdc,
		channelKeeper,
		portKeeper,
		scopedKeeper,
	)
	authKeeper := authkeeper.NewKeeper(
		cdc, authStoreKey, channelKeeper, packetSendKeeper, packetMiddleware,
		xccResolver,
	)
	initiatorKeeper := initiatorkeeper.NewKeeper(
		cdc, initiatorStoreKey, channelKeeper,
		authKeeper,
		packetSendKeeper,
		packetMiddleware,
		xccResolver,
		txRunner,
	)
	authKeeper.SetTxManager(initiatorKeeper)
	router.AddRoute(authtypes.PacketType, authKeeper)

	return Keeper{
		m:             cdc,
		portKeeper:    portKeeper,
		channelKeeper: channelKeeper,
		scopedKeeper:  scopedKeeper,
		router:        router,

		initiatorKeeper: initiatorKeeper,
		authKeeper:      authKeeper,
	}
}

func (k Keeper) InitiatorKeeper() initiatorkeeper.Keeper {
	return k.initiatorKeeper
}

func (k Keeper) AuthKeeper() authkeeper.Keeper {
	return k.authKeeper
}

// IsBound checks if the transfer module is already bound to the desired port
func (k Keeper) IsBound(ctx sdk.Context, portID string) bool {
	_, ok := k.scopedKeeper.GetCapability(ctx, host.PortPath(portID))
	return ok
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

// AuthenticateCapability wraps the scopedKeeper's AuthenticateCapability function
func (k Keeper) AuthenticateCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) bool {
	return k.scopedKeeper.AuthenticateCapability(ctx, cap, name)
}
