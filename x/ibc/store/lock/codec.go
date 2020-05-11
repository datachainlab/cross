package lock

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/datachainlab/cross/x/ibc/cross"
)

var cdc *codec.Codec

func init() {
	cdc = codec.New()
	RegisterCodec(cdc)
	RegisterOPCodec(cross.ModuleCdc)
}

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*OP)(nil), nil)
	cdc.RegisterInterface((*LockOP)(nil), nil)
	RegisterOPCodec(cdc)
}

func RegisterOPCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(ReadOP{}, "store/lock/ReadOP", nil)
	cdc.RegisterConcrete(WriteOP{}, "store/lock/WriteOP", nil)
}
