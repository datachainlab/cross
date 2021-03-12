package types

import (
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	"github.com/datachainlab/cross/x/packets"
)

const (
	PacketType = "cross/core/atomic/simple"
)

var _ packets.PacketDataPayload = (*PacketDataCall)(nil)

// NewPacketDataCall creates a new instance of PacketDataCall
func NewPacketDataCall(
	txID txtypes.TxID,
	tx txtypes.ResolvedContractTransaction,
) *PacketDataCall {
	return &PacketDataCall{TxId: txID, Tx: tx}
}

func (p PacketDataCall) ValidateBasic() error {
	if err := p.Tx.ValidateBasic(); err != nil {
		return err
	}
	return nil
}

func (PacketDataCall) Type() string {
	return PacketType
}

var _ packets.PacketAcknowledgementPayload = (*PacketAcknowledgementCall)(nil)

// NewPacketAcknowledgementCall creates a new instance of PacketAcknowledgementCall
func NewPacketAcknowledgementCall(status CommitStatus) *PacketAcknowledgementCall {
	return &PacketAcknowledgementCall{Status: status}
}

func (a PacketAcknowledgementCall) ValidateBasic() error {
	return nil
}

func (PacketAcknowledgementCall) Type() string {
	return PacketType
}
