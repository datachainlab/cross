package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type signersKey struct{}

func WithSigners(ctx sdk.Context, signers []sdk.AccAddress) sdk.Context {
	return ctx.WithValue(signersKey{}, signers)
}

func SignersFromContext(ctx sdk.Context) ([]sdk.AccAddress, bool) {
	signers, ok := ctx.Value(signersKey{}).([]sdk.AccAddress)
	return signers, ok
}
