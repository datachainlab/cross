package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/datachainlab/cross/x/utils"
)

// RegisterInterfaces register the ibc transfer module interfaces to protobuf
// Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*XCC)(nil),
		&ChannelInfo{},
	)
}

var (
	// ModuleCdc references the global x/ibc-transfer module codec. Note, the codec
	// should ONLY be used in certain instances of tests and for JSON encoding.
	//
	// The actual codec used for serialization should be provided to x/ibc-transfer and
	// defined at the application level.
	ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
)

func PackCrossChainChannel(xcc XCC) (*codectypes.Any, error) {
	return utils.PackAny(xcc)
}

func UnpackCrossChainChannel(m codec.Codec, anyXCC codectypes.Any) (XCC, error) {
	var xcc XCC
	if err := m.UnpackAny(&anyXCC, &xcc); err != nil {
		return nil, err
	}
	return xcc, nil
}
