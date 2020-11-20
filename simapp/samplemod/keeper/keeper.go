package keeper

import (
	"context"
	"encoding/hex"
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

	exContractCaller crosstypes.ExternalContractCaller
}

func NewKeeper(m codec.Marshaler, storeKey sdk.StoreKey, xstore crosstypes.Store) Keeper {
	return Keeper{
		m:                m,
		storeKey:         storeKey,
		xstore:           xstore,
		exContractCaller: crosstypes.NewExternalContractCaller(),
	}
}

// HandleContractCall is called by ContractModule
func (k Keeper) HandleContractCall(goCtx context.Context, callInfo crosstypes.ContractCallInfo) (*crosstypes.ContractCallResult, error) {
	var req types.ContractCallRequest
	if err := k.m.UnmarshalJSON(callInfo, &req); err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	switch req.Method {
	case "nop":
		return &crosstypes.ContractCallResult{}, nil
	case "counter":
		return k.HandleCounter(ctx, req)
	case "external-call":
		return k.HandleExternalCall(ctx, req)
	case "fail":
		return nil, errors.New("failed to process a contract request")
	default:
		panic(fmt.Sprintf("unknown method '%v'", req.Method))
	}
}

var counterKey = []byte("counter")

func (k Keeper) HandleCounter(ctx sdk.Context, req types.ContractCallRequest) (*crosstypes.ContractCallResult, error) {
	// use the account ID as namespace
	account := crosstypes.ContractSignersFromContext(ctx.Context())[0]
	v := k.getCounter(ctx, account)
	bz := k.setCounter(ctx, account, v+1)
	return &crosstypes.ContractCallResult{Data: bz}, nil
}

func (k Keeper) getCounter(ctx sdk.Context, account crosstypes.AccountID) uint64 {
	var count uint64
	v := k.xstore.Prefix(account).Get(ctx, counterKey)
	if v == nil {
		count = 0
	} else {
		count = sdk.BigEndianToUint64(v)
	}
	return count
}

func (k Keeper) setCounter(ctx sdk.Context, account crosstypes.AccountID, value uint64) []byte {
	bz := sdk.Uint64ToBigEndian(value)
	k.xstore.Prefix(account).Set(ctx, counterKey, bz)
	return bz
}

func (k Keeper) HandleExternalCall(ctx sdk.Context, req types.ContractCallRequest) (*crosstypes.ContractCallResult, error) {
	if len(req.Args) != 2 {
		return nil, fmt.Errorf("the number of arguments must be 2")
	}

	accID, err := hex.DecodeString(req.Args[0])
	if err != nil {
		return nil, err
	}
	channelID := req.Args[1]

	r := types.NewContractCallRequest("counter")
	callInfo := crosstypes.ContractCallInfo(k.m.MustMarshalJSON(&r))

	ret := k.exContractCaller.Call(
		ctx,
		&crosstypes.ChannelInfo{
			Port:    crosstypes.PortID,
			Channel: channelID,
		},
		callInfo,
		[]crosstypes.AccountID{accID},
	)
	return &crosstypes.ContractCallResult{Data: ret}, nil
}
