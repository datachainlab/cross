package keeper

import (
	"context"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/datachainlab/cross/simapp/samplemod/types"
	authtypes "github.com/datachainlab/cross/x/core/auth/types"
	contracttypes "github.com/datachainlab/cross/x/core/contract/types"
	storetypes "github.com/datachainlab/cross/x/core/store/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
)

type Keeper struct {
	m        codec.Codec
	storeKey sdk.StoreKey
	xstore   storetypes.KVStoreI

	exContractCaller contracttypes.ExternalContractCaller
}

func NewKeeper(m codec.Codec, storeKey sdk.StoreKey, xstore storetypes.KVStoreI) Keeper {
	return Keeper{
		m:                m,
		storeKey:         storeKey,
		xstore:           xstore,
		exContractCaller: contracttypes.NewExternalContractCaller(),
	}
}

// HandleContractCall is called by ContractModule
func (k Keeper) HandleContractCall(goCtx context.Context, signers []authtypes.Account, callInfo txtypes.ContractCallInfo) (*txtypes.ContractCallResult, error) {
	var req types.ContractCallRequest
	if err := k.m.UnmarshalJSON(callInfo, &req); err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	switch req.Method {
	case "nop":
		return &txtypes.ContractCallResult{}, nil
	case "counter":
		return k.HandleCounter(ctx, signers, req)
	case "external-call":
		return k.HandleExternalCall(ctx, req)
	case "fail":
		return nil, errors.New("failed to process a contract request")
	default:
		panic(fmt.Sprintf("unknown method '%v'", req.Method))
	}
}

var counterKey = []byte("counter")

func (k Keeper) HandleCounter(ctx sdk.Context, signers []authtypes.Account, req types.ContractCallRequest) (*txtypes.ContractCallResult, error) {
	// use the account ID as namespace
	account := signers[0]
	v := k.getCounter(ctx, account.Id)
	bz := k.setCounter(ctx, account.Id, v+1)
	return &txtypes.ContractCallResult{Data: bz}, nil
}

func (k Keeper) getCounter(ctx sdk.Context, account authtypes.AccountID) uint64 {
	var count uint64
	v := k.xstore.Prefix(account).Get(ctx, counterKey)
	if v == nil {
		count = 0
	} else {
		count = sdk.BigEndianToUint64(v)
	}
	return count
}

func (k Keeper) setCounter(ctx sdk.Context, account authtypes.AccountID, value uint64) []byte {
	bz := sdk.Uint64ToBigEndian(value)
	k.xstore.Prefix(account).Set(ctx, counterKey, bz)
	return bz
}

func (k Keeper) HandleExternalCall(ctx sdk.Context, req types.ContractCallRequest) (*txtypes.ContractCallResult, error) {
	if len(req.Args) != 2 {
		return nil, fmt.Errorf("the number of arguments must be 2")
	}

	acc, err := authtypes.NewAccountFromHexString(req.Args[0])
	if err != nil {
		return nil, err
	}
	channelID := req.Args[1]

	r := types.NewContractCallRequest("counter")
	callInfo := txtypes.ContractCallInfo(k.m.MustMarshalJSON(&r))

	ret := k.exContractCaller.Call(
		ctx,
		&xcctypes.ChannelInfo{
			Port:    crosstypes.PortID,
			Channel: channelID,
		},
		callInfo,
		[]authtypes.Account{*acc},
	)
	return &txtypes.ContractCallResult{Data: ret}, nil
}
