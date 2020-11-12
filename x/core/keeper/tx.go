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

	if len(remaining) == 0 {
		return txID, true, nil
	}

	// save MsgInitiateTx
	// waiting for other signers
	if err := k.setTxState(ctx, txID, msg, remaining); err != nil {
		return nil, false, err
	}

	return txID, false, nil
}

func (k Keeper) setTxState(ctx sdk.Context, txID types.TxID, msg types.MsgInitiateTx, remainingSigners []types.Account) error {
	bz, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}
	prefix.NewStore(k.store(ctx), types.KeyInitiateTx()).Set(txID, bz)

	state := types.NewInitiateTxState(remainingSigners)
	bz, err = proto.Marshal(&state)
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
func (k Keeper) SignTx(txID types.TxID, signer types.Account) error {
	// TODO implement
	return nil
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
