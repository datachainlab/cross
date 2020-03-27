package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrFailedInitiateTx             = sdkerrors.Register(ModuleName, 2, "failed to initiate a transaction")
	ErrFailedPrepare                = sdkerrors.Register(ModuleName, 3, "failed to prepare a commit")
	ErrFailedRecievePrepareResult   = sdkerrors.Register(ModuleName, 4, "failed to receive a PrepareResult")
	ErrFailedMulticastCommitPacket  = sdkerrors.Register(ModuleName, 5, "failed to multicast a CommitPacket")
	ErrFailedReceiveCommitPacket    = sdkerrors.Register(ModuleName, 6, "failed to receive a CommitPacket")
	ErrFailedSendAckCommitPacket    = sdkerrors.Register(ModuleName, 7, "failed to send an AckCommitPacket")
	ErrFailedReceiveAckCommitPacket = sdkerrors.Register(ModuleName, 8, "failed to receive an AckCommitPacket")

	ErrCoordinatorNotFound = sdkerrors.Register(ModuleName, 100, "coordinator not found")
)
