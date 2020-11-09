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

// Initiate defines a rpc handler method for MsgInitiate.
func (k Keeper) Initiate(goCtx context.Context, msg *types.MsgInitiate) (*types.MsgInitiateResponse, error) {
	ctx, ps, err := k.packetMiddleware.HandleMsg(sdk.UnwrapSDKContext(goCtx), msg, packets.NewBasicPacketSender(k.ChannelKeeper()))
	if err != nil {
		return nil, err
	}

	// Validations

	if ctx.ChainID() != msg.ChainId {
		return nil, fmt.Errorf("unexpected chainID: '%v' != '%v'", ctx.ChainID(), msg.ChainId)
	} else if ctx.BlockHeight() >= int64(msg.TimeoutHeight.GetVersionHeight()) {
		return nil, fmt.Errorf("this msg is already timeout: current=%v timeout=%v", ctx.BlockHeight(), msg.TimeoutHeight)
	}

	// Initiate a transaction

	txID := types.MakeTxID(msg)
	switch msg.CommitProtocol {
	case types.CommitProtocolSimple:
		err := k.SimpleKeeper().SendCall(ctx, ps, txID, msg.ContractTransactions, msg.TimeoutHeight, msg.TimeoutTimestamp)
		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrFailedInitiateTx, err.Error())
		}
	default:
		return nil, fmt.Errorf("unknown Commit protocol '%v'", msg.CommitProtocol)
	}

	return &types.MsgInitiateResponse{}, nil
}
