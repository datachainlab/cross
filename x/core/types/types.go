package types

import "github.com/gogo/protobuf/proto"

func NewAcknowledgement(success bool, result []byte) *Acknowledgement {
	return &Acknowledgement{
		Success: success,
		Result:  result,
	}
}

func (ack Acknowledgement) GetBytes() []byte {
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
