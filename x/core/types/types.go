package types

import (
	"github.com/cosmos/ibc-go/modules/core/exported"
	"github.com/gogo/protobuf/proto"
)

type (
	// TxID represents a ID of transaction. This value must be unique in a chain
	TxID = []byte
	// TxIndex represents an index of an array of contract transactions
	TxIndex = uint32
)

var _ exported.Acknowledgement = (*Acknowledgement)(nil)

func NewAcknowledgement(success bool, result []byte) *Acknowledgement {
	return &Acknowledgement{
		IsSuccess: success,
		Result:    result,
	}
}

func (ack Acknowledgement) Success() bool {
	return ack.IsSuccess
}

func (ack Acknowledgement) Acknowledgement() []byte {
	bz, err := proto.Marshal(&ack)
	if err != nil {
		panic(err)
	}
	return bz
}

func UnmarshalAcknowledgement(bz []byte) (*Acknowledgement, error) {
	var ack Acknowledgement
	if err := proto.Unmarshal(bz, &ack); err != nil {
		return nil, err
	}
	return &ack, nil
}
