package keeper

import (
	"context"

	authtypes "github.com/datachainlab/cross/x/core/auth/types"
	initiatortypes "github.com/datachainlab/cross/x/core/initiator/types"
)

var _ initiatortypes.QueryServer = (*Keeper)(nil)

func (q Keeper) SelfXCC(c context.Context, req *initiatortypes.QuerySelfXCCRequest) (*initiatortypes.QuerySelfXCCResponse, error) {
	return q.initiatorKeeper.SelfXCC(c, req)
}

func (q Keeper) TxAuthState(c context.Context, req *authtypes.QueryTxAuthStateRequest) (*authtypes.QueryTxAuthStateResponse, error) {
	return q.authKeeper.TxAuthState(c, req)
}
