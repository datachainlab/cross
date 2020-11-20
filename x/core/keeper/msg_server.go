package keeper

import (
	"context"

	initiatortypes "github.com/datachainlab/cross/x/initiator/types"
)

var _ initiatortypes.MsgServer = (*Keeper)(nil)

func (k Keeper) InitiateTx(ctx context.Context, msg *initiatortypes.MsgInitiateTx) (*initiatortypes.MsgInitiateTxResponse, error) {
	return k.initiatorKeeper.InitiateTx(ctx, msg)
}

func (k Keeper) SignTx(ctx context.Context, msg *initiatortypes.MsgSignTx) (*initiatortypes.MsgSignTxResponse, error) {
	return k.initiatorKeeper.SignTx(ctx, msg)
}

func (k Keeper) IBCSignTx(ctx context.Context, msg *initiatortypes.MsgIBCSignTx) (*initiatortypes.MsgIBCSignTxResponse, error) {
	return k.initiatorKeeper.IBCSignTx(ctx, msg)
}
