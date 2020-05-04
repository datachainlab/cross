package keeper

import (
	"fmt"

	"github.com/datachainlab/cross/x/ibc/cross/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryCoordinatorStatus:
			return queryCoordinatorStatus(ctx, keeper, req)
		case types.QueryUnacknowledgedPackets:
			return queryUnacknowledgedPackets(ctx, keeper)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown IBC %s query endpoint", types.ModuleName)
		}
	}
}

func queryCoordinatorStatus(ctx sdk.Context, k Keeper, req abci.RequestQuery) ([]byte, error) {
	var query types.QueryCoordinatorStatusRequest
	if err := k.cdc.UnmarshalBinaryLengthPrefixed(req.Data, &query); err != nil {
		return nil, err
	}
	ci, ok := k.GetCoordinator(ctx, query.TxID)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrCoordinatorNotFound, fmt.Sprintf("Coordinator for TxID '%X' not found", query.TxID))
	}
	res := types.QueryCoordinatorStatusResponse{
		TxID:            query.TxID,
		Completed:       ci.IsReceivedALLAcks(),
		CoordinatorInfo: *ci,
	}
	return k.cdc.MarshalBinaryLengthPrefixed(res)
}

func queryUnacknowledgedPackets(ctx sdk.Context, k Keeper) ([]byte, error) {
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
	return k.cdc.MarshalBinaryLengthPrefixed(res)
}
