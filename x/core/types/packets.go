package types

import (
	"errors"

	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/datachainlab/cross/x/packets"
)

var _ packets.PacketDataPayload = (*PacketDataIBCSignTx)(nil)

// NewPacketDataIBCSignTx creates a new instance of PacketDataIBCSignTx
func NewPacketDataIBCSignTx(
	txID TxID,
	signers []AccountID,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) PacketDataIBCSignTx {
	return PacketDataIBCSignTx{
		TxID:             txID,
		Signers:          signers,
		TimeoutHeight:    timeoutHeight,
		TimeoutTimestamp: timeoutTimestamp,
	}
}

func (p PacketDataIBCSignTx) ValidateBasic() error {
	if len(p.TxID) == 0 {
		return errors.New("txID must not be empty")
	}
	if len(p.Signers) == 0 {
		return errors.New("signers are required")
	}
	return nil
}

var _ packets.PacketAcknowledgementPayload = (*PacketAcknowledgementIBCSignTx)(nil)

func (p PacketAcknowledgementIBCSignTx) ValidateBasic() error {
	return nil
}
