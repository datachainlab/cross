package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/datachainlab/cross/x/core/auth/types"
)

var _ authtypes.AuthExtensionVerifier = (*SampleAuthExtension)(nil)

func (SampleAuthExtension) Verify(signer authtypes.Account, signature signing.SignatureV2, tx sdk.Tx) error {
	// always returns nil
	return nil
}
