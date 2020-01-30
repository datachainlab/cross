package lock

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var cdc *codec.Codec

func init() {
	cdc = codec.New()
	RegisterCodec(cdc)
}

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*OP)(nil), nil)
	cdc.RegisterConcrete(Read{}, "store/lock/Read", nil)
	cdc.RegisterConcrete(Write{}, "store/lock/Write", nil)
}
