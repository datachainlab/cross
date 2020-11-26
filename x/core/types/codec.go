package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	initiatortypes "github.com/datachainlab/cross/x/core/initiator/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
)

// RegisterInterfaces registers x/ibc interfaces into protobuf Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	xcctypes.RegisterInterfaces(registry)
	initiatortypes.RegisterInterfaces(registry)
	txtypes.RegisterInterfaces(registry)
}
