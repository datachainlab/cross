package types

import (
	"encoding/json"

	sptypes "github.com/bluele/interchain-simple-packet/types"
	"github.com/cosmos/cosmos-sdk/codec"
)

type (
	Header     = sptypes.Header
	PacketData = sptypes.PacketData
)

// NewPacketData returns a new packet data
func NewPacketData(h *sptypes.Header, payload []byte) PacketData {
	if h == nil {
		h = &sptypes.Header{}
	}
	return sptypes.NewSimplePacketData(*h, payload)
}

type PacketDataPayload interface {
	ValidateBasic() error
	GetBytes() []byte
	Type() string
}

type PacketAcknowledgement interface {
	ValidateBasic() error
	GetBytes() []byte
	Type() string
}

func MarshalPacketData(data PacketData) ([]byte, error) {
	bz, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return bz, nil
}

func UnmarshalPacketData(bz []byte, pd *PacketData) error {
	return json.Unmarshal(bz, pd)
}

func UnmarshalPacketDataPayload(cdc *codec.Codec, bz []byte, pd *PacketData, ptr interface{}) error {
	if err := UnmarshalPacketData(bz, pd); err != nil {
		return err
	}
	return cdc.UnmarshalJSON(pd.Payload, ptr)
}
