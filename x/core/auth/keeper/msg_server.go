package keeper

import (
	"context"

	"github.com/datachainlab/cross/x/core/auth/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	accounttypes "github.com/datachainlab/cross/x/core/account/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
	"github.com/datachainlab/cross/x/packets"
)

var _ types.MsgServer = (*Keeper)(nil)

// SignTx defines a rpc handler method for MsgSignTx.
func (k Keeper) SignTx(goCtx context.Context, msg *types.MsgSignTx) (*types.MsgSignTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var accounts []accounttypes.Account
	for _, addr := range msg.Signers {
		accounts = append(accounts, accounttypes.NewAccount(k.xccResolver.GetSelfCrossChainChannel(ctx), addr))
	}
	completed, err := k.Sign(ctx, msg.TxID, accounts)
	if err != nil {
		return nil, err
	}
	return &types.MsgSignTxResponse{TxAuthCompleted: completed}, nil
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
