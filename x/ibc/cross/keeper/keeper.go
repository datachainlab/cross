package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/ibc/cross/types"
)

// Keeper maintains the link to storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	cdc               *codec.Codec // The wire codec for binary encoding/decoding.
	storeKey          sdk.StoreKey // Unexposed key to access store from sdk.Context
	boundedCapability sdk.CapabilityKey

	channelKeeper types.ChannelKeeper
}

// NewKeeper creates new instances of the cross Keeper
func NewKeeper(
	cdc *codec.Codec,
	storeKey sdk.StoreKey,
	capKey sdk.CapabilityKey,
	channelKeeper types.ChannelKeeper,
) Keeper {
	return Keeper{
		cdc:               cdc,
		storeKey:          storeKey,
		boundedCapability: capKey,
		channelKeeper:     channelKeeper,
	}
}

func (k Keeper) SetTx(ctx sdk.Context, txID types.TxID, txIndex types.TxIndex, tx types.TxInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(tx)
	store.Set(types.KeyTx(txID, txIndex), bz)
}

func (k Keeper) EnsureTxStatus(ctx sdk.Context, txID types.TxID, txIndex types.TxIndex, status uint8) (*types.TxInfo, error) {
	tx, found := k.GetTx(ctx, txID, txIndex)
	if !found {
		return nil, fmt.Errorf("txID '%x' not found", txID)
	}
	if tx.Status == status {
		return tx, nil
	} else {
		return nil, fmt.Errorf("expected status is %v, but got %v", status, tx.Status)
	}
}

func (k Keeper) UpdateTxStatus(ctx sdk.Context, txID types.TxID, txIndex types.TxIndex, status uint8) error {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyTx(txID, txIndex))
	if bz == nil {
		return fmt.Errorf("txID '%x' not found", txID)
	}
	var tx types.TxInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &tx)
	tx.Status = status
	k.SetTx(ctx, txID, txIndex, tx)
	return nil
}

func (k Keeper) GetTx(ctx sdk.Context, txID types.TxID, txIndex types.TxIndex) (*types.TxInfo, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyTx(txID, txIndex))
	if bz == nil {
		return nil, false
	}
	var tx types.TxInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &tx)
	return &tx, true
}

func (k Keeper) SetCoordinator(ctx sdk.Context, txID types.TxID, ci types.CoordinatorInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(ci)
	store.Set(types.KeyCoordinator(txID), bz)
}

func (k Keeper) GetCoordinator(ctx sdk.Context, txID types.TxID) (*types.CoordinatorInfo, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyCoordinator(txID))
	if bz == nil {
		return nil, false
	}
	var ci types.CoordinatorInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &ci)
	return &ci, true
}

func (k Keeper) UpdateCoordinatorStatus(ctx sdk.Context, txID types.TxID, status uint8) error {
	ci, found := k.GetCoordinator(ctx, txID)
	if !found {
		return fmt.Errorf("txID '%x' not found", txID)
	}
	ci.Status = status
	k.SetCoordinator(ctx, txID, *ci)
	return nil
}

func (k Keeper) EnsureCoordinatorStatus(ctx sdk.Context, txID types.TxID, status uint8) (*types.CoordinatorInfo, error) {
	ci, found := k.GetCoordinator(ctx, txID)
	if !found {
		return nil, fmt.Errorf("txID '%x' not found", txID)
	}
	if ci.Status == status {
		return ci, nil
	} else {
		return nil, fmt.Errorf("expected status is %v, but got %v", status, ci.Status)
	}
}

func (k Keeper) SetContractResult(ctx sdk.Context, txID types.TxID, txIndex types.TxIndex, result types.ContractHandlerResult) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(result)
	store.Set(types.KeyContractResult(txID, txIndex), bz)
}

func (k Keeper) GetContractResult(ctx sdk.Context, txID types.TxID, txIndex types.TxIndex) (types.ContractHandlerResult, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyContractResult(txID, txIndex))
	if bz == nil {
		return nil, false
	}
	var result types.ContractHandlerResult
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &result)
	return result, true
}

func (k Keeper) RemoveContractResult(ctx sdk.Context, txID types.TxID, txIndex types.TxIndex) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyContractResult(txID, txIndex))
}
