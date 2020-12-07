package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/core/auth/types"
)

var _ types.QueryServer = (*Keeper)(nil)

func (q Keeper) TxAuthState(c context.Context, req *types.QueryTxAuthStateRequest) (*types.QueryTxAuthStateResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	state, err := q.getAuthState(ctx, req.TxID)
	if err != nil {
		return nil, err
	}
	return &types.QueryTxAuthStateResponse{TxAuthState: state}, nil
}
