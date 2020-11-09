package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
)

// RegisterInterfaces register the ibc transfer module interfaces to protobuf
// Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*LockOP)(nil),
		&WriteOP{},
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

func ConvertOPsToLockOPs(m codec.Marshaler, ops crosstypes.OPs) ([]LockOP, error) {
	var lks []LockOP
	for _, op := range ops.Items {
		var lk LockOP
		if err := m.UnpackAny(&op, &lk); err != nil {
			return nil, err
		}
		lks = append(lks, lk)
	}
	return lks, nil
}

func ConvertLockOPsToOPs(lks []LockOP) (*crosstypes.OPs, error) {
	var ops crosstypes.OPs
	for _, lk := range lks {
		var any codectypes.Any
		if err := any.Pack(lk); err != nil {
			return nil, err
		}
		ops.Items = append(ops.Items, any)
	}
	return &ops, nil
}

func ConvertOPItemsToOPs(items []OP) (*crosstypes.OPs, error) {
	var ops crosstypes.OPs
	for _, it := range items {
		var any codectypes.Any
		if err := any.Pack(it); err != nil {
			return nil, err
		}
		ops.Items = append(ops.Items, any)
	}
	return &ops, nil
}
