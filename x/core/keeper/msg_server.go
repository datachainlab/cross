package keeper

import (
	"context"

	authtypes "github.com/datachainlab/cross/x/core/auth/types"
	initiatortypes "github.com/datachainlab/cross/x/core/initiator/types"
)

var _ initiatortypes.MsgServer = (*Keeper)(nil)

func (k Keeper) InitiateTx(ctx context.Context, msg *initiatortypes.MsgInitiateTx) (*initiatortypes.MsgInitiateTxResponse, error) {
	return k.initiatorKeeper.InitiateTx(ctx, msg)
}

func (k Keeper) SignTx(ctx context.Context, msg *authtypes.MsgSignTx) (*authtypes.MsgSignTxResponse, error) {
	return k.authKeeper.SignTx(ctx, msg)
}

func (k Keeper) IBCSignTx(ctx context.Context, msg *authtypes.MsgIBCSignTx) (*authtypes.MsgIBCSignTxResponse, error) {
	return k.authKeeper.IBCSignTx(ctx, msg)
}
