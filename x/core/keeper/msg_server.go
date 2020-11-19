package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/core/types"
	"github.com/datachainlab/cross/x/packets"
)

var _ types.MsgServer = (*Keeper)(nil)

// InitiateTx defines a rpc handler method for MsgInitiateTx.
func (k Keeper) InitiateTx(goCtx context.Context, msg *types.MsgInitiateTx) (*types.MsgInitiateTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validations

	if ctx.ChainID() != msg.ChainId {
		return nil, fmt.Errorf("unexpected chainID: '%v' != '%v'", ctx.ChainID(), msg.ChainId)
	} else if !msg.TimeoutHeight.IsZero() && ctx.BlockHeight() >= int64(msg.TimeoutHeight.GetVersionHeight()) {
		return nil, fmt.Errorf("the Msg is already timeout: current=%v timeout-height=%v", ctx.BlockHeight(), msg.TimeoutHeight)
	} else if msg.TimeoutTimestamp > 0 && uint64(ctx.BlockTime().Unix()) >= msg.TimeoutTimestamp {
		return nil, fmt.Errorf("the Msg is already timeout: current=%v timeout-timestamp=%v", ctx.BlockTime().Unix(), msg.TimeoutTimestamp)
	}

	// Check if all participants sign the tx

	txID, completed, err := k.initTx(ctx, msg)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent("tx", sdk.NewAttribute("id", hex.EncodeToString(txID))))

	if !completed {
		return &types.MsgInitiateTxResponse{TxID: txID, Status: types.INITIATE_TX_STATUS_PENDING}, nil
	}

	// Run a transaction

	// FIXME can this method returns an error? we should cleanup txState and txMsg.
	if err := k.runTx(ctx, txID, msg); err != nil {
		return nil, err
	}
	return &types.MsgInitiateTxResponse{TxID: txID, Status: types.INITIATE_TX_STATUS_VERIFIED}, nil
}

// SignTx defines a rpc handler method for MsgSignTx.
func (k Keeper) SignTx(goCtx context.Context, msg *types.MsgSignTx) (*types.MsgSignTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var accounts []types.Account
	for _, addr := range msg.Signers {
		accounts = append(accounts, types.NewAccount(k.CrossChainChannelResolver().GetSelfCrossChainChannel(ctx), addr))
	}
	status, err := k.signTx(ctx, msg.TxID, accounts)
	if err != nil {
		return nil, err
	}
	return &types.MsgSignTxResponse{Status: status}, nil
}

// IBCSignTx defines a rpc handler method for MsgIBCSignTx.
func (k Keeper) IBCSignTx(goCtx context.Context, msg *types.MsgIBCSignTx) (*types.MsgIBCSignTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	xcc, err := types.UnpackCrossChainChannel(k.m, *msg.CrossChainChannel)
	if err != nil {
		return nil, err
	}

	// Run packet middlewares

	ctx, ps, err := k.packetMiddleware.HandleMsg(ctx, msg, packets.NewBasicPacketSender(k.ChannelKeeper()))
	if err != nil {
		return nil, err
	}

	var accounts []types.Account
	for _, addr := range msg.Signers {
		accounts = append(accounts, types.NewAccount(k.CrossChainChannelResolver().GetSelfCrossChainChannel(ctx), addr))
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
	return &types.MsgIBCSignTxResponse{Status: 0}, nil
}
