package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	crosstypes "github.com/datachainlab/cross/x/core/types"
)

var (
	ErrFailedInitiateTx = sdkerrors.Register(crosstypes.ModuleName, 1200, "failed to initiate a cross-chain transaction")
)
