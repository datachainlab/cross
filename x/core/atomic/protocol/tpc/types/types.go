package types

import (
	atomictypes "github.com/datachainlab/cross/x/core/atomic/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	"github.com/datachainlab/cross/x/packets"
)

const (
	PacketType = "cross/core/atomic/tpc"
)

var _ packets.PacketDataPayload = (*PacketDataPrepare)(nil)

// NewPacketDataPrepare creates a new instance of PacketDataPrepare
func NewPacketDataPrepare(
	txID txtypes.TxID,
	tx txtypes.ResolvedContractTransaction,
	txIndex txtypes.TxIndex,
) PacketDataPrepare {
	return PacketDataPrepare{TxId: txID, TxIndex: txIndex, Tx: tx}
}

func (p PacketDataPrepare) ValidateBasic() error {
	if err := p.Tx.ValidateBasic(); err != nil {
		return err
	}
	return nil
}

func (PacketDataPrepare) Type() string {
	return PacketType
}

var _ packets.PacketAcknowledgementPayload = (*PacketAcknowledgementPrepare)(nil)

func NewPacketAcknowledgementPayload(
	result atomictypes.PrepareResult,
) *PacketAcknowledgementPrepare {
	return &PacketAcknowledgementPrepare{
		Result: result,
	}
}

func (a PacketAcknowledgementPrepare) ValidateBasic() error {
	return nil
}

func (PacketAcknowledgementPrepare) Type() string {
	return PacketType
}