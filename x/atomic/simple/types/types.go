package types

import (
	crosstypes "github.com/datachainlab/cross/x/core/types"
)

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

// NewPacketCallAcknowledgement creates a new instance of PacketCallAcknowledgement
func NewPacketCallAcknowledgement(status CommitStatus) *PacketCallAcknowledgement {
	return &PacketCallAcknowledgement{Status: status}
}
