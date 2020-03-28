package relayer

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/datachainlab/cross/x/ibc/relayer/client/cli"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

const (
	ModuleName = "relayer"
)

// type check to ensure the interface is properly implemented
var (
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic is an app module Basics object
type AppModuleBasic struct{}

// Name returns module name
func (AppModuleBasic) Name() string {
	return ModuleName
}

// RegisterCodec returns RegisterCodec
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {}

// DefaultGenesis returns default genesis state
func (AppModuleBasic) DefaultGenesis(m codec.JSONMarshaler) json.RawMessage {
	return nil
}

// ValidateGenesis checks the Genesis
func (AppModuleBasic) ValidateGenesis(m codec.JSONMarshaler, bz json.RawMessage) error {
	return nil
}

// RegisterRESTRoutes returns rest routes
func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router) {}

// GetQueryCmd returns the root query command of this module
func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	// return cli.GetQueryCmd(cdc)
	return nil
}

// GetTxCmd returns the root tx command of this module
func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetTxCmd(cdc)
}
