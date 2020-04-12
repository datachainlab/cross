package cross

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/ibc/cross/types"
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
	_, err := keeper.BindPort(ctx, types.PortID)
	if err != nil {
		panic(fmt.Sprintf("could not claim port capability: %v", err))
	}
	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports genesis
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	return GenesisState{}
}
