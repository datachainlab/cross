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
		(*CallResult)(nil),
		&ConstantValueCallResult{},
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

func PackCallResults(results []CallResult) ([]codectypes.Any, error) {
	var anys []codectypes.Any
	for _, result := range results {
		any, err := utils.PackAny(result)
		if err != nil {
			return nil, err
		}
		anys = append(anys, *any)
	}
	return anys, nil
}

func UnpackCallResults(m codec.Codec, anyCallResults []codectypes.Any) ([]CallResult, error) {
	var results []CallResult
	for _, v := range anyCallResults {
		var result CallResult
		if err := m.UnpackAny(&v, &result); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}
