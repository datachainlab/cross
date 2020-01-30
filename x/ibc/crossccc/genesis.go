package crossccc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// GenesisState is genesis state
type GenesisState struct {
}

// NewGenesisState is a constructor of GenesisState
func NewGenesisState(master string) GenesisState {
	return GenesisState{}
}

// ValidateGenesis checks the Genesis
func ValidateGenesis(data GenesisState) error {
	return nil
}

// DefaultGenesisState returns default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{}
}

// InitGenesis inits genesis
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports genesis
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	return GenesisState{}
}
