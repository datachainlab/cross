package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	accounttypes "github.com/datachainlab/cross/x/core/account/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
)

// TxAuthenticator defines the expected interface of cross-chain authenticator
type TxAuthenticator interface {
	// InitAuthState initializes the state of the tx corresponding to a given txID
	InitAuthState(ctx sdk.Context, txID txtypes.TxID, signers []accounttypes.Account) error
	// IsCompletedAuth returns a boolean whether the tx corresponding a given txID is completed
	IsCompletedAuth(ctx sdk.Context, txID txtypes.TxID) (bool, error)
	// Sign executes
	Sign(ctx sdk.Context, txID txtypes.TxID, signers []accounttypes.Account) (bool, error)
}

// TxManager defines the expected interface of transaction manager
type TxManager interface {
	// IsActive returns a boolean whether the tx corresponding to a given txID is still active
	IsActive(ctx sdk.Context, txID txtypes.TxID) (bool, error)
	// PostAuth represents a callback function is called at post authentication
	PostAuth(ctx sdk.Context, txID txtypes.TxID) error
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
