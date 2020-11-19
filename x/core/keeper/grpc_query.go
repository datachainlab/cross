package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/core/types"
)

var _ types.QueryServer = (*Keeper)(nil)

func (q Keeper) SelfXCC(c context.Context, req *types.QuerySelfXCCRequest) (*types.QuerySelfXCCResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	xcc := q.CrossChainChannelResolver().GetSelfCrossChainChannel(ctx)

	anyXCC, err := types.PackCrossChainChannel(xcc)
	if err != nil {
		return nil, err
	}
	return &types.QuerySelfXCCResponse{Xcc: anyXCC}, nil
}
