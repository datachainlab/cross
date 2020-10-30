package types

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
)

type ContractHandler interface {
	Handle(ctx context.Context, callInfo ContractCallInfo) error
}

type OP interface {
	proto.Message
	Equal(OP) bool
	String() string
}

type OPs []OP

func (ops OPs) Equal(other OPs) bool {
	if len(ops) != len(other) {
		return false
	}
	for i, op := range ops {
		if !op.Equal(other[i]) {
			return false
		}
	}
	return true
}

func (ops OPs) String() string {
	opStrs := make([]string, len(ops))
	for i, op := range ops {
		opStrs[i] = op.String()
	}
	return fmt.Sprintf("OPs{%v}", strings.Join(opStrs, ","))
}
