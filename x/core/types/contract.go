package types

import (
	"context"

	"github.com/gogo/protobuf/proto"
)

type ContractHandler interface {
	Handle(ctx context.Context, callInfo ContractCallInfo) error
}

type OP interface {
	proto.Message
}

func (ops OPs) Equal(other OPs) bool {
	if len(ops.Items) != len(other.Items) {
		return false
	}
	for i, op := range ops.Items {
		if !op.Equal(other.Items[i]) {
			return false
		}
	}
	return true
}
