package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/core/initiator/types"
)

var _ types.MsgServer = (*Keeper)(nil)

// InitiateTx defines a rpc handler method for MsgInitiateTx.
func (k Keeper) InitiateTx(goCtx context.Context, msg *types.MsgInitiateTx) (*types.MsgInitiateTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validations

	if ctx.ChainID() != msg.ChainId {
		return nil, fmt.Errorf("unexpected chainID: '%v' != '%v'", ctx.ChainID(), msg.ChainId)
	} else if !msg.TimeoutHeight.IsZero() && ctx.BlockHeight() >= int64(msg.TimeoutHeight.GetRevisionHeight()) {
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
