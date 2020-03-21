package keeper

import (
	"github.com/datachainlab/cross/x/ibc/cross/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		// case types.QueryAddress:
		// 	// return queryAddress(ctx, path[1:], req, keeper)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown IBC %s query endpoint", types.ModuleName)
		}
	}
}

// func queryAddress(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
// 	address, err := sdk.AccAddressFromBech32(string(req.Data))

// 	if err != nil {
// 		return nil, sdk.ErrInvalidAddress(address.String())
// 	}

// 	obj := keeper.Get(ctx, address)
// 	res, _ := json.Marshal(obj)

// 	return res, nil
// }
