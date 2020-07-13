package naive

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/ibc/cross/types"
)

const (
	TypeCall = "cross_naive_call"
)

var _ types.PacketData = (*PacketDataCall)(nil)

type PacketDataCall struct {
	Sender sdk.AccAddress
	TxID   types.TxID
	TxInfo types.ContractTransactionInfo
}

func NewPacketDataCall(
	sender sdk.AccAddress,
	txID types.TxID,
	txInfo types.ContractTransactionInfo,
) PacketDataCall {
	return PacketDataCall{Sender: sender, TxID: txID, TxInfo: txInfo}
}

func (p PacketDataCall) ValidateBasic() error {
	if err := p.TxInfo.ValidateBasic(); err != nil {
		return err
	}
	return nil
}

func (p PacketDataCall) GetBytes() []byte {
	return sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(p))
}

func (p PacketDataCall) GetTimeoutHeight() uint64 {
	return math.MaxUint64
}

func (p PacketDataCall) GetTimeoutTimestamp() uint64 {
	return 0
}

func (p PacketDataCall) Type() string {
	return TypeCall
}
