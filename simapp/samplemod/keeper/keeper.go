package keeper

import (
	"context"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/simapp/samplemod/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
)

type Keeper struct {
	m        codec.Marshaler
	storeKey sdk.StoreKey
	xstore   crosstypes.Store
}

func NewKeeper(m codec.Marshaler, storeKey sdk.StoreKey, xstore crosstypes.Store) Keeper {
	return Keeper{
		m:        m,
		storeKey: storeKey,
		xstore:   xstore,
	}
}

func (k Keeper) HandleContractCall(ctx context.Context, callInfo crosstypes.ContractCallInfo) (*crosstypes.ContractCallResult, *crosstypes.OPs, error) {
	var req types.ContractCallRequest
	if err := k.m.UnmarshalJSON(callInfo, &req); err != nil {
		return nil, nil, err
	}
	switch req.Method {
	case "nop":
		return &crosstypes.ContractCallResult{}, nil, nil
	case "fail":
		return nil, nil, errors.New("failed to process a contract request")
	default:
		panic(fmt.Sprintf("unknown method '%v'", req.Method))
	}
}
