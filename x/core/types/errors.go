package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/datachainlab/cross/x/core/host"
)

// cross module sentinel errors
var (
	ErrInvalidVersion         = sdkerrors.Register(host.ModuleName, 1100, "invalid cross module version")
	ErrInvalidAcknowledgement = sdkerrors.Register(host.ModuleName, 1101, "invalid acknowledgement")
	ErrAcknowledgementExists  = sdkerrors.Register(host.ModuleName, 1102, "acknowledgement for packet already exists")
)
