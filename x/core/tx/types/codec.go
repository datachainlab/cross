package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// RegisterInterfaces register the ibc transfer module interfaces to protobuf
// Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*Object)(nil),
		&ConstantValueObject{},
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

func PackObjects(objs []Object) ([]codectypes.Any, error) {
	var anys []codectypes.Any
	for _, obj := range objs {
		var any codectypes.Any
		if err := any.Pack(obj); err != nil {
			return nil, err
		}
		anys = append(anys, any)
	}
	return anys, nil
}

func UnpackObjects(m codec.Marshaler, objects []codectypes.Any) ([]Object, error) {
	var objs []Object
	for _, v := range objects {
		var obj Object
		if err := m.UnpackAny(&v, &obj); err != nil {
			return nil, err
		}
		objs = append(objs, obj)
	}
	return objs, nil
}
