package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/datachainlab/cross/x/core/types"
	"github.com/datachainlab/cross/x/packets"
)

var _ types.MsgServer = Keeper{}

// InitiateTx defines a rpc handler method for MsgInitiateTx.
func (k Keeper) InitiateTx(goCtx context.Context, msg *types.MsgInitiateTx) (*types.MsgInitiateTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validations

	if ctx.ChainID() != msg.ChainId {
		return nil, fmt.Errorf("unexpected chainID: '%v' != '%v'", ctx.ChainID(), msg.ChainId)
	} else if ctx.BlockHeight() >= int64(msg.TimeoutHeight.GetVersionHeight()) {
		return nil, fmt.Errorf("this msg is already timeout: current=%v timeout=%v", ctx.BlockHeight(), msg.TimeoutHeight)
	}

	// Check if all participants sign the tx

	txID, ok, err := k.verifyTx(ctx, *msg)
	if !ok {
		return &types.MsgInitiateTxResponse{Status: types.INITIATE_TX_STATUS_PENDING}, nil
	}

	// Run packet middlewares

	ctx, ps, err := k.packetMiddleware.HandleMsg(ctx, msg, packets.NewBasicPacketSender(k.ChannelKeeper()))
	if err != nil {
		return nil, err
	}

	// Initiate a transaction

	switch msg.CommitProtocol {
	case types.CommitProtocolSimple:
		err := k.SimpleKeeper().SendCall(ctx, ps, txID, msg.ContractTransactions, msg.TimeoutHeight, msg.TimeoutTimestamp)
		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrFailedInitiateTx, err.Error())
		}
	default:
		return nil, fmt.Errorf("unknown Commit protocol '%v'", msg.CommitProtocol)
	}

	return &types.MsgInitiateTxResponse{Status: types.INITIATE_TX_STATUS_VERIFIED}, nil
}

// SignTx defines a rpc handler method for MsgSignTx.
func (k Keeper) SignTx(goCtx context.Context, msg *types.MsgSignTx) (*types.MsgSignTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var accounts []types.Account
	for _, addr := range msg.Signers {
		accounts = append(accounts, types.NewLocalAccount(addr))
	}
	status, err := k.signTx(ctx, msg.TxID, accounts)
	if err != nil {
		return nil, err
	}
	return &types.MsgSignTxResponse{Status: status}, nil
}
