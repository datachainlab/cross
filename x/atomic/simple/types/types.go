package types

import (
	crosstypes "github.com/datachainlab/cross/x/core/types"
	"github.com/datachainlab/cross/x/packets"
)

var _ packets.PacketDataPayload = (*PacketDataCall)(nil)

// NewPacketDataCall creates a new instance of PacketDataCall
func NewPacketDataCall(
	txID crosstypes.TxID,
	txInfo crosstypes.ContractTransactionInfo,
) PacketDataCall {
	return PacketDataCall{TxId: txID, TxInfo: txInfo}
}

func (p PacketDataCall) ValidateBasic() error {
	if err := p.TxInfo.ValidateBasic(); err != nil {
		return err
	}
	return nil
}

var _ packets.PacketDataPayload = (*PacketAcknowledgementCall)(nil)

// NewPacketAcknowledgementCall creates a new instance of PacketAcknowledgementCall
func NewPacketAcknowledgementCall(status CommitStatus) *PacketAcknowledgementCall {
	return &PacketAcknowledgementCall{Status: status}
}

func (a PacketAcknowledgementCall) ValidateBasic() error {
	return nil
}
