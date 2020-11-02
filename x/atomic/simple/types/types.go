package types

import "github.com/datachainlab/cross/x/core/types"

func NewPacketDataCall(
	sender types.AccountAddress,
	txID types.TxID,
	txInfo types.ContractTransactionInfo,
) PacketDataCall {
	return PacketDataCall{Sender: sender, TxId: txID, TxInfo: txInfo}
}

func (p PacketDataCall) ValidateBasic() error {
	if err := p.TxInfo.ValidateBasic(); err != nil {
		return err
	}
	return nil
}
