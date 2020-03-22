package types

/* TODO a copy from other projects. we should delete or rewrite these.
 */

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// IBC channel sentinel errors
var (
	ErrExample = sdkerrors.Register(ModuleName, 1, "an error example")
)
