package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/atomic/common/types"
)

var _ types.QueryServer = (*Keeper)(nil)

func (q Keeper) CoordinatorState(c context.Context, req *types.QueryCoordinatorStateRequest) (*types.QueryCoordinatorStateResponse, error) {
	cs, found := q.GetCoordinatorState(sdk.UnwrapSDKContext(c), req.TxId)
	if !found {
		return nil, fmt.Errorf("txID '%x' not found", req.TxId)
	}
	return &types.QueryCoordinatorStateResponse{CoodinatorState: *cs}, nil
}
