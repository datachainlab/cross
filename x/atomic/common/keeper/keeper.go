package keeper

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
	commitStore   types.CommitStore

	contractHandler  types.ContractHandler
	resolverProvider types.ObjectResolverProvider
	channelResolver  types.ChannelResolver
}

func NewKeeper(
	cdc codec.Marshaler,
	storeKey sdk.StoreKey,
	channelKeeper types.ChannelKeeper,
	portKeeper types.PortKeeper,
	scopedKeeper capabilitykeeper.ScopedKeeper,
	contractHandler types.ContractHandler,
	commitStore types.CommitStore,
) Keeper {
	return Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		channelKeeper:   channelKeeper,
		portKeeper:      portKeeper,
		scopedKeeper:    scopedKeeper,
		commitStore:     commitStore,
		contractHandler: contractHandler,
	}
}

func (k Keeper) ChannelKeeper() types.ChannelKeeper {
	return k.channelKeeper
}

func (k Keeper) ChannelResolver() types.ChannelResolver {
	return k.channelResolver
}

func (k Keeper) PrepareCommit(
	ctx sdk.Context,
	txID types.TxID,
	txIndex types.TxIndex,
	tx types.ContractTransaction,
	links []types.Object,
) error {
	res, err := k.processTransaction(ctx, txIndex, tx, links, types.AtomicMode)
	if err != nil {
		return err
	}
	k.SetContractResult(ctx, txID, txIndex, *res)
	return k.commitStore.Precommit(ctx, makeContractTransactionID(txID, txIndex))
}

func (k Keeper) processTransaction(
	ctx sdk.Context,
	txIndex types.TxIndex,
	tx types.ContractTransaction,
	links []types.Object,
	commitMode types.CommitMode,
) (res *types.ContractHandlerResult, err error) {
	// TODO resolverProvider can be moved into contract package?
	rs, err := k.resolverProvider(links)
	if err != nil {
		return nil, err
	}

	// Setup a context
	goCtx := types.SetupContractContext(
		sdk.WrapSDKContext(ctx),
		types.ContractRuntimeInfo{
			CommitMode:             commitMode,
			StateConstraintType:    tx.StateConstraint.Type,
			ExternalObjectResolver: rs,
		},
	)
	ops, err := k.contractHandler(
		goCtx,
		tx.CallInfo,
	)
	if err != nil {
		return nil, err
	}

	if rv := tx.ReturnValue; !rv.IsNil() && !rv.Equal(types.NewReturnValue(res.Data)) {
		return nil, fmt.Errorf("unexpected return-value: expected='%X' actual='%X'", *rv, res.Data)
	}

	if !tx.StateConstraint.Ops.Equal(ops) {
		return nil, fmt.Errorf("unexpected ops: actual(%v) != expected(%v)", ops.String(), tx.StateConstraint.Ops.String())
	}

	return res, nil
}
