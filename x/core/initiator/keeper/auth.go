package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/datachainlab/cross/x/core/auth/types"
	"github.com/datachainlab/cross/x/core/initiator/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
)

var _ authtypes.TxManager = (*Keeper)(nil)

// IsActive implements TxManager interface
func (a Keeper) IsActive(ctx sdk.Context, txID txtypes.TxID) (bool, error) {
	// TODO add timeout support to initiator?
	_, found := a.getTxState(ctx, txID)
	if !found {
		return false, nil
	}
	return true, nil
}

// OnPostAuth implements TxManager interface
func (a Keeper) OnPostAuth(ctx sdk.Context, txID txtypes.TxID) error {
	txState, found := a.getTxState(ctx, txID)
	if !found {
		return fmt.Errorf("txState '%x' not found", txID)
	}
	if txState.IsVerified() {
		return fmt.Errorf("txState '%x' is already verified", txID)
	}
	txState.Status = types.INITIATE_TX_STATUS_VERIFIED
	a.setTxState(ctx, txID, *txState)
	return a.runTx(ctx, txID, &txState.Msg)
}
