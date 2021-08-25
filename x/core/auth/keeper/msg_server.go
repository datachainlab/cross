package keeper

import (
	"bytes"
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	accounttypes "github.com/datachainlab/cross/x/core/account/types"
	"github.com/datachainlab/cross/x/core/auth/types"
	initiatorkeeper "github.com/datachainlab/cross/x/core/initiator/keeper"
	initiatortypes "github.com/datachainlab/cross/x/core/initiator/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
	"github.com/datachainlab/cross/x/packets"
)

var _ types.MsgServer = (*Keeper)(nil)

// SignTx defines a rpc handler method for MsgSignTx.
func (k Keeper) SignTx(goCtx context.Context, msg *types.MsgSignTx) (*types.MsgSignTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	initiateKeeper, ok := k.txManager.(initiatorkeeper.Keeper)
	if !ok {
		return nil, fmt.Errorf("unrecognized txManager type: %T, msg", k.txManager)
	}
	txState, found := initiateKeeper.GetTxState(ctx, msg.TxID)
	if !found {
		return nil, fmt.Errorf("txState '%x' not found", msg.TxID)
	} else if txState.IsVerified() {
		return nil, fmt.Errorf("txState '%x' is already verified", msg.TxID)
	}

	xcc, err := func (ctx sdk.Context, msg *types.MsgSignTx, txState *initiatortypes.InitiateTxState) (xcctypes.XCC, error) {
		for _, msgSigner := range msg.Signers {
			for _, tx := range txState.Msg.ContractTransactions {
				for _, txSigner := range tx.Signers {
					if bytes.Equal(msgSigner, txSigner) {
						if xcc, err := xcctypes.UnpackCrossChainChannel(k.m, *tx.CrossChainChannel); err != nil {
							return nil, err
						} else {
							return k.xccResolver.ResolveCrossChainChannel(ctx, xcc)
						}
					}
				}
			}
		}

		return nil, fmt.Errorf("the same signer does not exist. %v:%v", msg.Signers, txState.Msg.Signers)
	}(ctx, msg, txState)
	if err != nil {
		return nil, err
	}

	var accounts []accounttypes.Account
	for _, addr := range msg.Signers {
		accounts = append(accounts, accounttypes.NewAccount(xcc, addr))
	}

	completed, err := k.Sign(ctx, msg.TxID, accounts)
	if err != nil {
		return nil, err
	}

	res := &types.MsgSignTxResponse{TxAuthCompleted: completed}
	if completed {
		if err := k.txManager.OnPostAuth(ctx, msg.TxID); err != nil {
			k.Logger(ctx).Error("failed to call PostAuth", "err", err)
			res.Log = err.Error()
		}
	}
	return res, nil
}

// IBCSignTx defines a rpc handler method for MsgIBCSignTx.
func (k Keeper) IBCSignTx(goCtx context.Context, msg *types.MsgIBCSignTx) (*types.MsgIBCSignTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	xcc, err := xcctypes.UnpackCrossChainChannel(k.m, *msg.CrossChainChannel)
	if err != nil {
		return nil, err
	}

	// Run packet middlewares

	ctx, ps, err := k.packetMiddleware.HandleMsg(ctx, msg, packets.NewBasicPacketSender(k.channelKeeper))
	if err != nil {
		return nil, err
	}

	var accounts []accounttypes.Account
	for _, addr := range msg.Signers {
		accounts = append(accounts, accounttypes.NewAccount(k.xccResolver.GetSelfCrossChainChannel(ctx), addr))
	}

	err = k.SendIBCSignTx(
		ctx,
		ps,
		xcc,
		msg.TxID,
		msg.Signers,
		msg.TimeoutHeight,
		msg.TimeoutTimestamp,
	)
	if err != nil {
		return nil, err
	}
	return &types.MsgIBCSignTxResponse{}, nil
}
