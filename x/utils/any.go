package utils

import (
	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

func PackAny(msg proto.Message) (*types.Any, error) {
	return types.NewAnyWithValue(msg)
}
