package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/ibc/atomic module sentinel errors
var (
	ErrInvalidVersion         = sdkerrors.Register(ModuleName, 1100, "invalid atomic module version")
	ErrInvalidAcknowledgement = sdkerrors.Register(ModuleName, 1101, "invalid acknowledgement")
	ErrAcknowledgementExists  = sdkerrors.Register(ModuleName, 1102, "acknowledgement for packet already exists")
)
