package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/datachainlab/cross/x/packets"
)

// RegisterInterfaces register the ibc transfer module interfaces to protobuf
// Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*packets.PacketDataPayload)(nil),
		&PacketDataPrepare{},
		&PacketDataCommit{},
	)
	registry.RegisterImplementations(
		(*packets.PacketAcknowledgementPayload)(nil),
		&PacketAcknowledgementPrepare{},
		&PacketAcknowledgementCommit{},
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
