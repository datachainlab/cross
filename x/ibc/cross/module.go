package cross

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/capability"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	port "github.com/cosmos/cosmos-sdk/x/ibc/05-port"
	porttypes "github.com/cosmos/cosmos-sdk/x/ibc/05-port/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
	"github.com/datachainlab/cross/x/ibc/cross/client/cli"
	"github.com/datachainlab/cross/x/ibc/cross/client/rest"
	"github.com/datachainlab/cross/x/ibc/cross/types"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

// type check to ensure the interface is properly implemented
var (
	_ module.AppModule      = AppModule{}
	_ port.IBCModule        = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic is an app module Basics object
type AppModuleBasic struct{}

// Name returns module name
func (AppModuleBasic) Name() string {
	return ModuleName
}

// RegisterCodec returns RegisterCodec
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	RegisterCodec(cdc)
}

// DefaultGenesis returns default genesis state
func (AppModuleBasic) DefaultGenesis(m codec.JSONMarshaler) json.RawMessage {
	return m.MustMarshalJSON(DefaultGenesisState())
}

// ValidateGenesis checks the Genesis
func (AppModuleBasic) ValidateGenesis(m codec.JSONMarshaler, bz json.RawMessage) error {
	var data GenesisState
	err := m.UnmarshalJSON(bz, &data)
	if err != nil {
		return err
	}
	// Once json successfully marshalled, passes along to genesis.go
	return ValidateGenesis(data)
}

// RegisterRESTRoutes returns rest routes
func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router) {
	rest.RegisterRoutes(ctx, rtr, QuerierRoute)
}

// GetQueryCmd returns the root query command of this module
func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetQueryCmd(cdc)
}

// GetTxCmd returns the root tx command of this module
func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetTxCmd(cdc)
}

// AppModule struct
type AppModule struct {
	AppModuleBasic
	keeper          Keeper
	packetReceiver  PacketReceiver
	contractHandler ContractHandler
}

// NewAppModule creates a new AppModule Object
func NewAppModule(k Keeper, contractHandler ContractHandler) AppModule {
	return AppModule{
		AppModuleBasic:  AppModuleBasic{},
		keeper:          k,
		packetReceiver:  NewPacketReceiver(k, contractHandler),
		contractHandler: contractHandler,
	}
}

// Name returns module name
func (AppModule) Name() string {
	return ModuleName
}

// RegisterInvariants is empty
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

// Route returns RouterKey
func (am AppModule) Route() string {
	return RouterKey
}

// NewHandler returns new Handler
func (am AppModule) NewHandler() sdk.Handler {
	return NewHandler(am.keeper)
}

// QuerierRoute returns module name
func (am AppModule) QuerierRoute() string {
	return ModuleName
}

// NewQuerierHandler returns new Querier
func (am AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(am.keeper)
}

// BeginBlock is a callback function
func (am AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlock is a callback function
func (am AppModule) EndBlock(sdk.Context, abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// InitGenesis inits genesis
func (am AppModule) InitGenesis(ctx sdk.Context, m codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState GenesisState
	m.MustUnmarshalJSON(data, &genesisState)
	return InitGenesis(ctx, am.keeper, genesisState)
}

// ExportGenesis exports genesis
func (am AppModule) ExportGenesis(ctx sdk.Context, m codec.JSONMarshaler) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return m.MustMarshalJSON(gs)
}

// Implement IBCModule callbacks
func (am AppModule) OnChanOpenInit(
	ctx sdk.Context,
	order channelexported.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *capability.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) error {
	// TODO: Enforce ordering, currently relayers use ORDERED channels

	if counterparty.PortID != types.PortID {
		return sdkerrors.Wrapf(porttypes.ErrInvalidPort, "counterparty has invalid portid. expected: %s, got %s", types.PortID, counterparty.PortID)
	}

	if version != types.Version {
		return sdkerrors.Wrapf(porttypes.ErrInvalidPort, "invalid version: %s, expected %s", version, types.Version)
	}

	// Claim channel capability passed back by IBC module
	if err := am.keeper.ClaimCapability(ctx, chanCap, ibctypes.ChannelCapabilityPath(portID, channelID)); err != nil {
		return sdkerrors.Wrap(channel.ErrChannelCapabilityNotFound, err.Error()+"by cross chanOpenInit")
	}

	// TODO: escrow
	return nil
}

func (am AppModule) OnChanOpenTry(
	ctx sdk.Context,
	order channelexported.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capability.Capability,
	counterparty channeltypes.Counterparty,
	version,
	counterpartyVersion string,
) error {
	// TODO: Enforce ordering, currently relayers use ORDERED channels

	if counterparty.PortID != types.PortID {
		return sdkerrors.Wrapf(porttypes.ErrInvalidPort, "counterparty has invalid portid. expected: %s, got %s", types.PortID, counterparty.PortID)
	}

	if version != types.Version {
		return sdkerrors.Wrapf(porttypes.ErrInvalidPort, "invalid version: %s, expected %s", version, types.Version)
	}

	if counterpartyVersion != types.Version {
		return sdkerrors.Wrapf(porttypes.ErrInvalidPort, "invalid counterparty version: %s, expected %s", counterpartyVersion, types.Version)
	}

	// Claim channel capability passed back by IBC module
	if err := am.keeper.ClaimCapability(ctx, chanCap, ibctypes.ChannelCapabilityPath(portID, channelID)); err != nil {
		return sdkerrors.Wrap(channel.ErrChannelCapabilityNotFound, err.Error()+"by cross chanOpenTry")
	}

	// TODO: escrow
	return nil
}

func (am AppModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyVersion string,
) error {
	if counterpartyVersion != types.Version {
		return sdkerrors.Wrapf(porttypes.ErrInvalidPort, "invalid counterparty version: %s, expected %s", counterpartyVersion, types.Version)
	}
	return nil
}

func (am AppModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

func (am AppModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

func (am AppModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

func (am AppModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
) (*sdk.Result, error) {
	return am.packetReceiver(ctx, packet)
}

func (am AppModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledment []byte,
) (*sdk.Result, error) {
	return nil, nil
}

func (am AppModule) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
) (*sdk.Result, error) {
	return nil, nil
}
