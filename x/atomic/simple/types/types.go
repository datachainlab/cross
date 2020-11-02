package types

import "github.com/datachainlab/cross/x/core/types"

func NewPacketDataCall(
	txID types.TxID,
	txInfo types.ContractTransactionInfo,
) PacketDataCall {
	return PacketDataCall{TxId: txID, TxInfo: txInfo}
}

func (p PacketDataCall) ValidateBasic() error {
	if err := p.TxInfo.ValidateBasic(); err != nil {
		return err
	}
	return nil
}
