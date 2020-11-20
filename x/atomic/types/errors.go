package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/datachainlab/cross/x/core/host"
)

var (
	ErrFailedInitiateTx = sdkerrors.Register(host.ModuleName, 1200, "failed to initiate a cross-chain transaction")
)
