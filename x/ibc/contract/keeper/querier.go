package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/datachainlab/cross/x/ibc/contract/types"
	"github.com/datachainlab/cross/x/ibc/cross"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(handler sdk.Handler, keeper Keeper, contractHandler cross.ContractHandler) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QuerySimulation:
			return querySimulation(ctx, handler, keeper, req)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown contract %s query endpoint", types.ModuleName)
		}
	}
}

func querySimulation(ctx sdk.Context, handler sdk.Handler, k Keeper, req abci.RequestQuery) ([]byte, error) {
	var msg types.MsgContractCall
	if err := k.cdc.UnmarshalJSON(req.Data, &msg); err != nil {
		return nil, err
	}
	res, err := handler(withSimulation(ctx), msg)
	if err != nil {
		return nil, err
	}
	return k.cdc.MarshalJSON(res)
}

type simulationKey struct{}

func withSimulation(ctx sdk.Context) sdk.Context {
	return ctx.WithValue(simulationKey{}, true)
}

func IsSimulation(ctx sdk.Context) bool {
	v, ok := ctx.Value(simulationKey{}).(bool)
	if !ok {
		return false
	}
	return v
}
