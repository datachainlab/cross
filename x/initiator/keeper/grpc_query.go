package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/initiator/types"
	xcctypes "github.com/datachainlab/cross/x/xcc/types"
)

var _ types.QueryServer = (*Keeper)(nil)

func (q Keeper) SelfXCC(c context.Context, req *types.QuerySelfXCCRequest) (*types.QuerySelfXCCResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	xcc := q.xccResolver.GetSelfCrossChainChannel(ctx)

	anyXCC, err := xcctypes.PackCrossChainChannel(xcc)
	if err != nil {
		return nil, err
	}
	return &types.QuerySelfXCCResponse{Xcc: anyXCC}, nil
}
