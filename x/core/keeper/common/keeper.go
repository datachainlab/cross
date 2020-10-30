package common

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	"github.com/datachainlab/cross/x/core/types"
)

type Keeper struct {
	cdc      codec.Marshaler
	storeKey sdk.StoreKey

	channelKeeper types.ChannelKeeper
	portKeeper    types.PortKeeper
	scopedKeeper  capabilitykeeper.ScopedKeeper

	contractHandler  types.ContractHandler
	resolverProvider types.ObjectResolverProvider
	// channelResolver  types.ChannelResolver
}

func NewKeeper(
	cdc codec.Marshaler,
	storeKey sdk.StoreKey,
	channelKeeper types.ChannelKeeper,
	portKeeper types.PortKeeper,
	scopedKeeper capabilitykeeper.ScopedKeeper,
	contractHandler types.ContractHandler,
) Keeper {
	return Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		channelKeeper:   channelKeeper,
		portKeeper:      portKeeper,
		scopedKeeper:    scopedKeeper,
		contractHandler: contractHandler,
	}
}

func (k Keeper) ChannelKeeper() types.ChannelKeeper {
	return k.channelKeeper
}

func (k Keeper) PrepareCommit(
	ctx sdk.Context,
	txID types.TxID,
	txIndex types.TxIndex,
	tx types.ContractTransaction,
	links []types.Object,
) error {
	res, err := k.processTransaction(ctx, txIndex, tx, links)
	if err != nil {
		return err
	}
	k.SetContractResult(ctx, txID, txIndex, *res)
	ctxID := makeContractTransactionID(txID, txIndex)
	_ = ctxID
	// return store.Precommit(ctxID)
	panic("not implemented error")
}

func (k Keeper) processTransaction(
	ctx sdk.Context,
	txIndex types.TxIndex,
	tx types.ContractTransaction,
	links []types.Object,
) (res *types.ContractHandlerResult, err error) {
	rs, err := k.resolverProvider(links)
	if err != nil {
		return nil, err
	}

	// Build a context
	goCtx := sdk.WrapSDKContext(ctx)
	runtimeInfo := types.ContractRuntimeInfo{
		StateConstraintType: tx.StateConstraint.Type, ExternalObjectResolver: rs,
	}
	// TODO set context to this
	_ = runtimeInfo
	if err := k.contractHandler.Handle(goCtx, tx.CallInfo); err != nil {
		return nil, err
	}

	if rv := tx.ReturnValue; !rv.IsNil() && !rv.Equal(types.NewReturnValue(res.Data)) {
		return nil, fmt.Errorf("unexpected return-value: expected='%X' actual='%X'", *rv, res.Data)
	}

	// TODO
	// if ops := store.OPs(); !ops.Equal(tx.StateConstraint.OPs) {
	// 	return nil, nil, fmt.Errorf("unexpected ops: actual(%v) != expected(%v)", ops.String(), tx.StateConstraint.OPs.String())
	// }

	return res, nil
}
