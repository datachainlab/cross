package tpc

import (
	"fmt"

	"github.com/datachainlab/cross/x/ibc/cross/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

func QueryCoordinatorStatus(ctx sdk.Context, k Keeper, req abci.RequestQuery) ([]byte, error) {
	var query types.QueryCoordinatorStatusRequest
	if err := k.cdc.UnmarshalJSON(req.Data, &query); err != nil {
		return nil, err
	}
	ci, ok := k.GetCoordinator(ctx, query.TxID)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrCoordinatorNotFound, fmt.Sprintf("TxID '%X' not found", query.TxID))
	}
	res := types.QueryCoordinatorStatusResponse{
		TxID:            query.TxID,
		Completed:       ci.IsReceivedALLAcks(),
		CoordinatorInfo: *ci,
	}
	return k.cdc.MarshalJSON(res)
}

func QueryUnacknowledgedPackets(ctx sdk.Context, k Keeper) ([]byte, error) {
	packets := []types.UnacknowledgedPacket{}
	k.IterateUnacknowledgedPackets(ctx, func(key []byte) bool {
		p, err := types.ParseUnacknowledgedPacketKey(key)
		if err != nil {
			panic(err)
		}
		packets = append(packets, *p)
		return false
	})
	res := types.QueryUnacknowledgedPacketsResponse{
		Packets: packets,
	}
	return k.cdc.MarshalJSON(res)
}
