package keeper

import (
	"fmt"

	"github.com/bluele/crossccc/x/ibc/crossccc/internal/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	mapset "github.com/deckarep/golang-set"
)

// Keeper maintains the link to storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	cdc               *codec.Codec // The wire codec for binary encoding/decoding.
	storeKey          sdk.StoreKey // Unexposed key to access store from sdk.Context
	boundedCapability sdk.CapabilityKey

	channelKeeper types.ChannelKeeper
}

// NewKeeper creates new instances of the crossccc Keeper
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

const (
	TX_STATUS_PREPARE uint8 = iota + 1
	TX_STATUS_COMMIT
	TX_STATUS_ABORT
)

type TxInfo struct {
	Status      uint8
	Coordinator types.ChannelInfo
	Contract    []byte
}

func NewTxInfo(status uint8, coordinator types.ChannelInfo, contract []byte) TxInfo {
	return TxInfo{Status: status, Coordinator: coordinator, Contract: contract}
}

func (k Keeper) SetTx(ctx sdk.Context, txID []byte, tx TxInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(tx)
	store.Set(types.KeyTx(txID), bz)
}

func (k Keeper) EnsureTxStatus(ctx sdk.Context, txID []byte, status uint8) (*TxInfo, error) {
	tx, found := k.GetTx(ctx, txID)
	if !found {
		return nil, fmt.Errorf("txID '%x' not found", txID)
	}
	if tx.Status == status {
		return tx, nil
	} else {
		return nil, fmt.Errorf("expected status is %v, but got %v", status, tx.Status)
	}
}

func (k Keeper) UpdateTxStatus(ctx sdk.Context, txID []byte, status uint8) error {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyTx(txID))
	if bz == nil {
		return fmt.Errorf("txID '%x' not found", txID)
	}
	var tx TxInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &tx)
	tx.Status = status
	k.SetTx(ctx, txID, tx)
	return nil
}

func (k Keeper) GetTx(ctx sdk.Context, txID []byte) (*TxInfo, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyTx(txID))
	if bz == nil {
		return nil, false
	}
	var tx TxInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &tx)
	return &tx, true
}

const (
	CO_STATUS_INIT uint8 = iota + 1
	CO_STATUS_COMMIT
)

type CoordinatorInfo struct {
	Status      uint8
	Transitions []string // [ConnectionID]
}

type ConnectionTransitionPair struct {
	ConnectionID string
	TransitionID int
}

func (ci CoordinatorInfo) Set() mapset.Set {
	set := mapset.NewSet()
	for tsID, connID := range ci.Transitions {
		set.Add(ConnectionTransitionPair{connID, tsID})
	}
	return set
}

func NewCoordinatorInfo(status uint8, tss []string) CoordinatorInfo {
	return CoordinatorInfo{Status: status, Transitions: tss}
}

func (k Keeper) SetCoordinator(ctx sdk.Context, txID []byte, ci CoordinatorInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(ci)
	store.Set(types.KeyCoordinator(txID), bz)
}

func (k Keeper) GetCoordinator(ctx sdk.Context, txID []byte) (*CoordinatorInfo, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyCoordinator(txID))
	if bz == nil {
		return nil, false
	}
	var ci CoordinatorInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &ci)
	return &ci, true
}

func (k Keeper) UpdateCoordinatorStatus(ctx sdk.Context, txID []byte, status uint8) error {
	ci, found := k.GetCoordinator(ctx, txID)
	if !found {
		return fmt.Errorf("txID '%x' not found", txID)
	}
	ci.Status = status
	k.SetCoordinator(ctx, txID, *ci)
	return nil
}

func (k Keeper) EnsureCoordinatorStatus(ctx sdk.Context, txID []byte, status uint8) (*CoordinatorInfo, error) {
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
