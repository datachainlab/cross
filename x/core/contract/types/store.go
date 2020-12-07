package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CommitMode indicates the type of the commit.
// It is used to tell the underlying store that manages the state.
// It is also expected to be set correctly by the transaction processor.
type CommitMode = uint8

const (
	// UnspecifiedMode indicates that nothing is specified.
	UnspecifiedMode CommitMode = iota + 1

	// BaseMode expects store operations to be committed in a single local transaction.
	// However, depending on the implementation of the store may need to prevent conflicts with concurrent transactions.
	BasicMode

	// AtomicMode expects store operations to be committed in cross-chain transaction.
	AtomicMode
)

// CommitStoreI defines the expected commit store in ContractManager
type CommitStoreI interface {
	Precommit(ctx sdk.Context, id []byte) error
	Abort(ctx sdk.Context, id []byte) error
	Commit(ctx sdk.Context, id []byte) error
	CommitImmediately(ctx sdk.Context)
}
