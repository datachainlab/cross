package types

import (
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
)

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
	return nil
}

func (p PacketDataIBCSignTx) GetAccounts() []Account {
	var accs []Account
	for _, id := range p.Signers {
		accs = append(accs, NewLocalAccount(AccountID(id)))
	}
	return accs
}
