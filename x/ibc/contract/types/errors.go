package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrFailedContractHandle = sdkerrors.Register(ModuleName, 2, "failed to execute contract handler")
	ErrFailedCommitStore    = sdkerrors.Register(ModuleName, 3, "failed to commit updates")
)
