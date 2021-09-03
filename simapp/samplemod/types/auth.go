package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/datachainlab/cross/x/core/auth/types"
)

var _ authtypes.AuthExtensionVerifier = (*SampleAuthExtension)(nil)

func (SampleAuthExtension) Verify(ctx sdk.Context, signer authtypes.Account, signature signing.SignatureV2, tx sdk.Tx) error {
	t := tx.(sdk.TxWithMemo)
	if t.GetMemo() == "sample" {
		return nil
	}
	return fmt.Errorf("unexpected memo: %v", t.GetMemo())
}
