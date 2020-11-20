package keeper

import (
	"context"

	"github.com/datachainlab/cross/x/initiator/types"
)

var _ types.QueryServer = (*Keeper)(nil)

func (q Keeper) SelfXCC(c context.Context, req *types.QuerySelfXCCRequest) (*types.QuerySelfXCCResponse, error) {
	return q.initiatorKeeper.SelfXCC(c, req)
}
