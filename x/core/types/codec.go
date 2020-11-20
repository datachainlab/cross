package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	initiatortypes "github.com/datachainlab/cross/x/initiator/types"
)

// RegisterInterfaces registers x/ibc interfaces into protobuf Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	initiatortypes.RegisterInterfaces(registry)
}
