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

var _ packets.PacketDataPayload = (*PacketCallAcknowledgement)(nil)

// NewPacketCallAcknowledgement creates a new instance of PacketCallAcknowledgement
func NewPacketCallAcknowledgement(status CommitStatus) *PacketCallAcknowledgement {
	return &PacketCallAcknowledgement{Status: status}
}

func (a PacketCallAcknowledgement) ValidateBasic() error {
	return nil
}
