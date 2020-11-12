package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/core/types"
	"github.com/gogo/protobuf/proto"
)

func (k Keeper) verifyTx(ctx sdk.Context, msg types.MsgInitiateTx) (types.TxID, bool, error) {
	txID := types.MakeTxID(&msg)

	_, found := k.getTxState(ctx, txID)
	if found {
		return nil, false, fmt.Errorf("txID '%X' already exists", txID)
	}

	signers := msg.GetAccounts()
	required := msg.GetRequiredAccounts()
	remaining := getRemainingAccounts(signers, required)

	state, err := k.initTxState(ctx, txID, msg, remaining)
	if err != nil {
		return nil, false, err
	}

	return txID, state.Status == types.INITIATE_TX_STATUS_VERIFIED, nil
}

func (k Keeper) initTxState(ctx sdk.Context, txID types.TxID, msg types.MsgInitiateTx, remainingSigners []types.Account) (*types.InitiateTxState, error) {
	bz, err := proto.Marshal(&msg)
	if err != nil {
		return nil, err
	}
	prefix.NewStore(k.store(ctx), types.KeyInitiateTx()).Set(txID, bz)

	state := types.NewInitiateTxState(remainingSigners)
	if err := k.setTxState(ctx, txID, state); err != nil {
		return nil, err
	}
	return &state, nil
}

func (k Keeper) setTxState(ctx sdk.Context, txID types.TxID, state types.InitiateTxState) error {
	bz, err := proto.Marshal(&state)
	if err != nil {
		return err
	}
	prefix.NewStore(k.store(ctx), types.KeyInitiateTxState()).Set(txID, bz)
	return nil
}

func (k Keeper) getTxState(ctx sdk.Context, txID types.TxID) (*types.InitiateTxState, bool) {
	var state types.InitiateTxState
	bz := prefix.NewStore(k.store(ctx), types.KeyInitiateTxState()).Get(txID)

	if bz == nil {
		return nil, false
	}

	if err := proto.Unmarshal(bz, &state); err != nil {
		panic(err)
	}
	return &state, true
}

// This method will be used in:
// 1. Signing asynchronously on same chain
// 2. Remote signing from other chain
func (k Keeper) signTx(ctx sdk.Context, txID types.TxID, signers []types.Account) (types.InitiateTxStatus, error) {
	state, found := k.getTxState(ctx, txID)
	if !found {
		return 0, fmt.Errorf("txID '%X' not found", txID)
	} else if state.Status != types.INITIATE_TX_STATUS_PENDING {
		return 0, fmt.Errorf("status must be %v", types.INITIATE_TX_STATUS_PENDING)
	}
	remaining := getRemainingAccounts(signers, state.RemainingSigners)
	if len(remaining) == 0 {
		state.Status = types.INITIATE_TX_STATUS_VERIFIED
	} else {
		state.Status = types.INITIATE_TX_STATUS_PENDING
	}
	if err := k.setTxState(ctx, txID, *state); err != nil {
		return 0, err
	}
	return state.Status, nil
}

func getRemainingAccounts(signers, required []types.Account) []types.Account {
	var state = make([]bool, len(required))
	for i, acc := range required {
		for _, s := range signers {
			if acc.Equal(s) {
				state[i] = true
			}
		}
	}
	var remaining []types.Account
	for i, acc := range required {
		if !state[i] {
			remaining = append(remaining, acc)
		}
	}
	return remaining
}
