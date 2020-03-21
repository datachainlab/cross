package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/datachainlab/cross/x/ibc/contract/internal/types"
	"github.com/datachainlab/cross/x/ibc/cross"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(handler sdk.Handler, keeper Keeper, contractHandler cross.ContractHandler) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.ModuleName:
			switch path[1] {
			case types.QuerySimulation:
				return querySimulation(ctx, handler, keeper, req)
			default:
				return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown contract %s query endpoint", types.ModuleName)
			}
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown contract query endpoint")
		}
	}
}

func querySimulation(ctx sdk.Context, handler sdk.Handler, k Keeper, req abci.RequestQuery) ([]byte, error) {
	var msg types.MsgContractCall
	if err := k.cdc.UnmarshalBinaryBare(req.Data, &msg); err != nil {
		return nil, err
	}
	res, err := handler(ctx, msg)
	if err != nil {
		return nil, err
	}
	return res.Data, nil
}
