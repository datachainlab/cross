package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	accounttypes "github.com/datachainlab/cross/x/core/account/types"
)

type TxAuthenticator interface {
	InitTxAuthState(ctx sdk.Context, id []byte, signers []accounttypes.Account) error
	IsCompletedTxAuth(ctx sdk.Context, id []byte) (bool, error)
	SignTx(ctx sdk.Context, id []byte, signers []accounttypes.Account) (bool, error)
}

// IsCompleted returns a boolean whether the required authentication is completed
func (s TxAuthState) IsCompleted() bool {
	return len(s.RemainingSigners) == 0
}

// ConsumeSigners removes the signers from required signers
func (s *TxAuthState) ConsumeSigners(signers []accounttypes.Account) {
	s.RemainingSigners = getRemainingAccounts(signers, s.RemainingSigners)
}

func getRemainingAccounts(signers, required []accounttypes.Account) []accounttypes.Account {
	var state = make([]bool, len(required))
	for i, acc := range required {
		for _, s := range signers {
			if acc.Equal(s) {
				state[i] = true
			}
		}
	}
	var remaining []accounttypes.Account
	for i, acc := range required {
		if !state[i] {
			remaining = append(remaining, acc)
		}
	}
	return remaining
}
