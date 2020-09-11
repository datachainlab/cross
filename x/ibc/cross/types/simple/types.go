package simple

import (
	"math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/ibc/cross/types"
)

const (
	TypeCall    = "cross_simple_call"
	TypeCallAck = "cross_simple_call_ack"
)

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(PacketDataCall{}, "cross/simple/PacketDataCall", nil)
	cdc.RegisterConcrete(PacketCallAcknowledgement{}, "cross/simple/PacketCallAcknowledgement", nil)
}

func init() {
	RegisterCodec(types.ModuleCdc)
}

const (
	COMMIT_OK uint8 = iota + 1
	COMMIT_FAILED
)

var _ types.PacketDataPayload = (*PacketDataCall)(nil)

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

var _ types.PacketAcknowledgement = (*PacketCallAcknowledgement)(nil)

type PacketCallAcknowledgement struct {
	Status uint8
}

func NewPacketCallAcknowledgement(status uint8) PacketCallAcknowledgement {
	return PacketCallAcknowledgement{Status: status}
}

func (p PacketCallAcknowledgement) ValidateBasic() error {
	return nil
}

func (p PacketCallAcknowledgement) GetBytes() []byte {
	return sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(p))
}

func (p PacketCallAcknowledgement) Type() string {
	return TypeCallAck
}

func (p PacketCallAcknowledgement) IsOK() bool {
	return p.Status == COMMIT_OK
}
