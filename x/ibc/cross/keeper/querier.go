package keeper

import (
	"github.com/datachainlab/cross/x/ibc/cross/keeper/tpc"
	"github.com/datachainlab/cross/x/ibc/cross/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryCoordinatorStatus:
			return tpc.QueryCoordinatorStatus(ctx, keeper.TPCKeeper(), req)
		case types.QueryUnacknowledgedPackets:
			return tpc.QueryUnacknowledgedPackets(ctx, keeper.TPCKeeper())
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown IBC %s query endpoint", types.ModuleName)
		}
	}
}
