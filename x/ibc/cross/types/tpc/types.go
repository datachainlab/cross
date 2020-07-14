package tpc

import (
	"math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/ibc/cross/types"
)

const (
	TypePrepare       = "cross_tpc_prepare"
	TypePrepareResult = "cross_tpc_prepare_result"
	TypeCommit        = "cross_tpc_commit"
	TypeAckCommit     = "cross_tpc_ack_commit"
)

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(PacketDataPrepare{}, "cross/tpc/PacketDataPrepare", nil)
	cdc.RegisterConcrete(PacketPrepareAcknowledgement{}, "cross/tpc/PacketPrepareAcknowledgement", nil)
	cdc.RegisterConcrete(PacketDataCommit{}, "cross/tpc/PacketDataCommit", nil)
	cdc.RegisterConcrete(PacketCommitAcknowledgement{}, "cross/tpc/PacketCommitAcknowledgement", nil)
}

func init() {
	RegisterCodec(types.ModuleCdc)
}

var _ types.PacketData = (*PacketDataPrepare)(nil)

type PacketDataPrepare struct {
	Sender  sdk.AccAddress
	TxID    types.TxID
	TxIndex types.TxIndex
	TxInfo  types.ContractTransactionInfo
}

func NewPacketDataPrepare(
	sender sdk.AccAddress,
	txID types.TxID,
	txIndex types.TxIndex,
	txInfo types.ContractTransactionInfo,
) PacketDataPrepare {
	return PacketDataPrepare{Sender: sender, TxID: txID, TxIndex: txIndex, TxInfo: txInfo}
}

func (p PacketDataPrepare) ValidateBasic() error {
	if err := p.TxInfo.ValidateBasic(); err != nil {
		return err
	}
	return nil
}

func (p PacketDataPrepare) GetBytes() []byte {
	return sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(p))
}

func (p PacketDataPrepare) GetTimeoutHeight() uint64 {
	return math.MaxUint64
}

func (p PacketDataPrepare) GetTimeoutTimestamp() uint64 {
	return 0
}

func (p PacketDataPrepare) Type() string {
	return TypePrepare
}

var _ types.PacketAcknowledgement = (*PacketPrepareAcknowledgement)(nil)

type PacketPrepareAcknowledgement struct {
	Status uint8
}

func NewPacketPrepareAcknowledgement(status uint8) PacketPrepareAcknowledgement {
	return PacketPrepareAcknowledgement{Status: status}
}

func (p PacketPrepareAcknowledgement) ValidateBasic() error {
	return nil
}

func (p PacketPrepareAcknowledgement) GetBytes() []byte {
	return sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(p))
}

func (p PacketPrepareAcknowledgement) Type() string {
	return TypePrepareResult
}

func (p PacketPrepareAcknowledgement) IsOK() bool {
	return p.Status == types.PREPARE_RESULT_OK
}

var _ types.PacketData = (*PacketDataCommit)(nil)

type PacketDataCommit struct {
	TxID          types.TxID
	TxIndex       types.TxIndex
	IsCommittable bool
}

func NewPacketDataCommit(txID types.TxID, txIndex types.TxIndex, isCommittable bool) PacketDataCommit {
	return PacketDataCommit{TxID: txID, TxIndex: txIndex, IsCommittable: isCommittable}
}

func (p PacketDataCommit) ValidateBasic() error {
	return nil
}

func (p PacketDataCommit) GetBytes() []byte {
	return sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(p))
}

func (p PacketDataCommit) GetTimeoutHeight() uint64 {
	return math.MaxUint64
}

func (p PacketDataCommit) GetTimeoutTimestamp() uint64 {
	return 0
}

func (p PacketDataCommit) Type() string {
	return TypeCommit
}

var _ types.PacketAcknowledgement = (*PacketCommitAcknowledgement)(nil)

type PacketCommitAcknowledgement struct{}

func NewPacketCommitAcknowledgement() PacketCommitAcknowledgement {
	return PacketCommitAcknowledgement{}
}

func (p PacketCommitAcknowledgement) ValidateBasic() error {
	return nil
}

func (p PacketCommitAcknowledgement) GetBytes() []byte {
	return sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(p))
}

func (p PacketCommitAcknowledgement) Type() string {
	return TypeAckCommit
}
