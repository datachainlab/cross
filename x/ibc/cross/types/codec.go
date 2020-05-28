package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// ModuleCdc is the codec for the module
var ModuleCdc = codec.New()

func init() {
	RegisterCodec(ModuleCdc)
}

// RegisterCodec registers concrete types on the Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgInitiate{}, "cross/MsgInitiate", nil)
	cdc.RegisterConcrete(ContractTransaction{}, "cross/ContractTransaction", nil)
	cdc.RegisterConcrete(ContractTransactions{}, "cross/ContractTransactions", nil)
	cdc.RegisterConcrete(ChannelInfo{}, "cross/ChannelInfo", nil)
	cdc.RegisterConcrete(ReturnValue{}, "cross/ReturnValue", nil)
	cdc.RegisterConcrete(CallResultLink{}, "cross/CallResultLink", nil)
	cdc.RegisterInterface((*Link)(nil), nil)
	cdc.RegisterInterface((*Object)(nil), nil)
	cdc.RegisterConcrete(ConstantValueObject{}, "cross/ConstantValueObject", nil)
	cdc.RegisterInterface((*PacketData)(nil), nil)
	cdc.RegisterInterface((*PacketAcknowledgement)(nil), nil)
	cdc.RegisterConcrete(PacketDataPrepare{}, "cross/PacketDataPrepare", nil)
	cdc.RegisterConcrete(PacketPrepareAcknowledgement{}, "cross/PacketPrepareAcknowledgement", nil)
	cdc.RegisterConcrete(PacketDataCommit{}, "cross/PacketDataCommit", nil)
	cdc.RegisterConcrete(PacketCommitAcknowledgement{}, "cross/PacketCommitAcknowledgement", nil)
	cdc.RegisterInterface((*OP)(nil), nil)
	cdc.RegisterInterface((*ContractHandlerResult)(nil), nil)
}
