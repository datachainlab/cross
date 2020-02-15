package contract

import "github.com/cosmos/cosmos-sdk/codec"

var cdc *codec.Codec

func init() {
	cdc = codec.New()
	RegisterCodec(cdc)
}

// RegisterCodec registers concrete types on the Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(ContractInfo{}, "contract/ContractInfo", nil)
}
