package contract

import "github.com/cosmos/cosmos-sdk/codec"

var cdc *codec.Codec

func init() {
	cdc = codec.New()

	cdc.RegisterConcrete(ContractInfo{}, "handler/ContractInfo", nil)
}
