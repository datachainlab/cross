package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/datachainlab/cross/x/core/auth/types"
)

// ExtAuthSigVerificationDecorator verifies all auth extension signatures for a tx and return an error if any are invalid.
// If a given tx has no such signatures, fallback to the wrapped decorator.
type ExtAuthSigVerificationDecorator struct {
	wrappedDec sdk.AnteDecorator
}

func NewExtAuthSigVerificationDecorator(wrappedDec sdk.AnteDecorator) ExtAuthSigVerificationDecorator {
	return ExtAuthSigVerificationDecorator{wrappedDec: wrappedDec}
}

func (dec ExtAuthSigVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// no need to verify signatures on recheck tx
	if ctx.IsReCheckTx() {
		return dec.wrappedDec.AnteHandle(ctx, tx, simulate, next)
	}

	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	var (
		signerNumber int
		stdSigExists bool
		extExists    bool
	)
	for _, msg := range tx.GetMsgs() {
		extMsg, ok := msg.(types.ExtAuthMsg)
		if !ok {
			stdSigExists = true
			continue
		}
		for _, signer := range extMsg.GetSignerAccounts() {
			if signer.AuthType.Mode == types.AuthMode_AUTH_MODE_EXTENSION {
				extExists = true
				signerNumber++
				if stdSigExists {
					return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidType, "invalid auth mode")
				}
			} else if extExists {
				return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidType, "invalid auth mode")
			} else {
				stdSigExists = true
			}
		}
	}
	// if all signers are not EXTENSION mode, just call the wrapped decorator
	if !extExists {
		return dec.wrappedDec.AnteHandle(ctx, tx, simulate, next)
	}

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	sigs, err := sigTx.GetSignaturesV2()
	if err != nil {
		return ctx, err
	}
	if len(sigs) != signerNumber {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "invalid number of signer;  expected: %d, got %d", signerNumber, len(sigs))
	}

	var i int = 0
	for _, msg := range tx.GetMsgs() {
		extMsg := msg.(types.ExtAuthMsg)
		for _, signer := range extMsg.GetSignerAccounts() {
			ext, ok := signer.AuthType.Option.GetCachedValue().(types.AuthExtensionVerifier)
			if !ok {
				return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidType, "unexpected extension type")
			}
			if err := ext.Verify(ctx, signer, sigs[i], tx); err != nil {
				return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, err.Error())
			}
			i++
		}
	}
	return next(ctx, tx, simulate)
}
