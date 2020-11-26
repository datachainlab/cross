package types

import (
	"errors"

	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	accounttypes "github.com/datachainlab/cross/x/account/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	"github.com/datachainlab/cross/x/packets"
)

const (
	PacketType = "cross/core"
)

var _ packets.PacketDataPayload = (*PacketDataIBCSignTx)(nil)

// NewPacketDataIBCSignTx creates a new instance of PacketDataIBCSignTx
func NewPacketDataIBCSignTx(
	txID txtypes.TxID,
	signers []accounttypes.AccountID,
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

func (PacketDataIBCSignTx) Type() string {
	return PacketType
}

var _ packets.PacketAcknowledgementPayload = (*PacketAcknowledgementIBCSignTx)(nil)

func (p PacketAcknowledgementIBCSignTx) ValidateBasic() error {
	return nil
}

func (PacketAcknowledgementIBCSignTx) Type() string {
	return PacketType
}
