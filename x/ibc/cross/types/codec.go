package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// ModuleCdc is the codec for the module
var ModuleCdc = codec.New()

func init() {
	RegisterCodec(ModuleCdc)
}

// RegisterCodec registers concrete types on the Amino codec
func RegisterCodec(cdc *codec.Codec) {
	// TODO These registrations should be included on ibc/channel module.
	cdc.RegisterConcrete(channeltypes.MsgAcknowledgement{}, "ibc/channel/MsgAcknowledgement", nil)

	cdc.RegisterConcrete(MsgInitiate{}, "cross/MsgInitiate", nil)
	cdc.RegisterConcrete(ContractTransaction{}, "cross/ContractTransaction", nil)
	cdc.RegisterConcrete(ContractTransactions{}, "cross/ContractTransactions", nil)
	cdc.RegisterConcrete(ChannelInfo{}, "cross/ChannelInfo", nil)
	cdc.RegisterConcrete(PacketDataPrepare{}, "cross/PacketDataPrepare", nil)
	cdc.RegisterConcrete(PacketDataPrepareResult{}, "cross/PacketDataPrepareResult", nil)
	cdc.RegisterConcrete(PacketDataCommit{}, "cross/PacketDataCommit", nil)
	cdc.RegisterConcrete(AckDataCommit{}, "cross/AckDataCommit", nil)
	cdc.RegisterInterface((*OP)(nil), nil)
}
