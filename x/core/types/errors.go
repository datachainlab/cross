package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// cross module sentinel errors
var (
	ErrInvalidVersion         = sdkerrors.Register(ModuleName, 1100, "invalid cross module version")
	ErrInvalidAcknowledgement = sdkerrors.Register(ModuleName, 1101, "invalid acknowledgement")
	ErrAcknowledgementExists  = sdkerrors.Register(ModuleName, 1102, "acknowledgement for packet already exists")

	ErrFailedInitiateTx = sdkerrors.Register(ModuleName, 1103, "failed to initiate a cross-chain transaction")
)
