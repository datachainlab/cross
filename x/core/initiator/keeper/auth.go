package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/datachainlab/cross/x/core/auth/types"
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
	return a.TryRunTx(ctx, txID)
}
