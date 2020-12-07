package keeper

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	accounttypes "github.com/datachainlab/cross/x/core/account/types"
	"github.com/datachainlab/cross/x/core/auth/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
	"github.com/datachainlab/cross/x/packets"
	"github.com/tendermint/tendermint/libs/log"
)

type Keeper struct {
	m        codec.Marshaler
	storeKey sdk.StoreKey

	channelKeeper    types.ChannelKeeper
	packetMiddleware packets.PacketMiddleware
	xccResolver      xcctypes.XCCResolver
	txManager        types.TxManager
	packets.PacketSendKeeper
}

var _ types.TxAuthenticator = (*Keeper)(nil)

func NewKeeper(
	m codec.Marshaler,
	storeKey sdk.StoreKey,
	channelKeeper types.ChannelKeeper,
	packetSendKeeper packets.PacketSendKeeper,
	packetMiddleware packets.PacketMiddleware,
	xccResolver xcctypes.XCCResolver,
) Keeper {
	return Keeper{
		m:        m,
		storeKey: storeKey,

		channelKeeper:    channelKeeper,
		xccResolver:      xccResolver,
		packetMiddleware: packetMiddleware,
		PacketSendKeeper: packetSendKeeper,
	}
}

// SetTxManager sets the keeper to txManager
func (k *Keeper) SetTxManager(txm types.TxManager) {
	k.txManager = txm
}

// InitAuthState implements the TxAuthenticator interface
func (k Keeper) InitAuthState(ctx sdk.Context, id []byte, signers []accounttypes.Account) error {
	_, err := k.getTxAuthState(ctx, id)
	if err == nil {
		return fmt.Errorf("id '%x' already exists", id)
	} else if !errors.Is(err, types.ErrIDNotFound{}) {
		return err
	}

	return k.setTxAuthState(ctx, id, types.TxAuthState{RemainingSigners: signers})
}

// IsCompletedAuth implements the TxAuthenticator interface
func (k Keeper) IsCompletedAuth(ctx sdk.Context, id []byte) (bool, error) {
	state, err := k.getTxAuthState(ctx, id)
	if err != nil {
		return false, err
	}
	return state.IsCompleted(), nil
}

// Sign implements the TxAuthenticator interface
func (k Keeper) Sign(ctx sdk.Context, id []byte, signers []accounttypes.Account) (bool, error) {
	state, err := k.getTxAuthState(ctx, id)
	if err != nil {
		return false, err
	}
	if state.IsCompleted() {
		return false, fmt.Errorf("id '%x' is already completed", id)
	}
	state.ConsumeSigners(signers)
	if err := k.setTxAuthState(ctx, id, *state); err != nil {
		return false, err
	}
	return state.IsCompleted(), nil
}

func (k Keeper) getTxAuthState(ctx sdk.Context, id []byte) (*types.TxAuthState, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyTxAuthState())
	bz := store.Get(id)
	if bz == nil {
		return nil, types.NewErrIDNotFound(id)
	}
	var state types.TxAuthState
	if err := k.m.UnmarshalBinaryBare(bz, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func (k Keeper) setTxAuthState(ctx sdk.Context, id []byte, state types.TxAuthState) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyTxAuthState())
	bz, err := k.m.MarshalBinaryBare(&state)
	if err != nil {
		return err
	}
	store.Set(id, bz)
	return nil
}

// Logger returns a logger instance
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s-%s", crosstypes.ModuleName, types.SubModuleName))
}
