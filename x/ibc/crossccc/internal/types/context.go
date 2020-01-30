package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type signerKey struct{}

func WithSigner(ctx sdk.Context, signer sdk.AccAddress) sdk.Context {
	return ctx.WithValue(signerKey{}, signer)
}

func SignerFromContext(ctx sdk.Context) (sdk.AccAddress, bool) {
	signer, ok := ctx.Value(signerKey{}).(sdk.AccAddress)
	return signer, ok
}
