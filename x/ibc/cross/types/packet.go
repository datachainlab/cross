package types

import (
	"encoding/json"

	sptypes "github.com/bluele/interchain-simple-packet/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
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

func UnmarshalIncomingPacket(cdc *codec.Codec, raw exported.PacketI) (IncomingPacket, error) {
	var pd PacketData
	var payload PacketDataPayload
	if err := UnmarshalPacketDataPayload(cdc, raw.GetData(), &pd, &payload); err != nil {
		return nil, err
	}
	return NewIncomingPacket(raw, pd, payload), nil
}

type IncomingPacket interface {
	exported.PacketI
	PacketData() PacketData
	Header() Header
	Payload() PacketDataPayload
}

var _ IncomingPacket = (*incomingPacket)(nil)

type incomingPacket struct {
	exported.PacketI
	packetData PacketData
	payload    PacketDataPayload
}

func NewIncomingPacket(raw exported.PacketI, packetData PacketData, payload PacketDataPayload) IncomingPacket {
	return &incomingPacket{
		PacketI:    raw,
		packetData: packetData,
		payload:    payload,
	}
}

func (p incomingPacket) PacketData() PacketData {
	return p.packetData
}

func (p incomingPacket) Header() Header {
	return p.packetData.Header
}

func (p incomingPacket) Payload() PacketDataPayload {
	return p.payload
}

type OutgoingPacket interface {
	IncomingPacket
	SetPacketData(header Header, payload PacketDataPayload)
}

var _ OutgoingPacket = (*outgoingPacket)(nil)

type outgoingPacket struct {
	exported.PacketI
	packetData PacketData
	payload    PacketDataPayload
}

func NewOutgoingPacket(raw exported.PacketI, packetData PacketData, payload PacketDataPayload) OutgoingPacket {
	return &outgoingPacket{
		PacketI:    raw,
		packetData: packetData,
		payload:    payload,
	}
}

func (p outgoingPacket) PacketData() PacketData {
	return p.packetData
}

func (p outgoingPacket) Header() Header {
	return p.packetData.Header
}

func (p outgoingPacket) Payload() PacketDataPayload {
	return p.payload
}

func (p *outgoingPacket) SetPacketData(header Header, payload PacketDataPayload) {
	p.payload = payload
	p.packetData = NewPacketData(&header, payload.GetBytes())
}

func (p outgoingPacket) GetData() []byte {
	bz, err := MarshalPacketData(p.packetData)
	if err != nil {
		panic(err)
	}
	return bz
}
