package utils

import (
	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

func PackAny(msg proto.Message) (*types.Any, error) {
	any := &types.Any{}
	err := any.Pack(msg)
	if err != nil {
		return nil, err
	}
	return any, nil
}
