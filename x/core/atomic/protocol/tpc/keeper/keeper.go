package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	basekeeper "github.com/datachainlab/cross/x/core/atomic/protocol/base/keeper"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
)

const (
	TypeName = "tpc"
)

type Keeper struct {
	cdc codec.Marshaler

	cm          txtypes.ContractManager
	xccResolver xcctypes.XCCResolver

	basekeeper.Keeper
}

func NewKeeper(
	cdc codec.Marshaler,
	cm txtypes.ContractManager,
	xccResolver xcctypes.XCCResolver,
	baseKeeper basekeeper.Keeper,
) Keeper {
	return Keeper{
		cdc:         cdc,
		cm:          cm,
		xccResolver: xccResolver,
		Keeper:      baseKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("cross/core/atomic/%s", TypeName))
}
