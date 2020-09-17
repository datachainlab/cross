package types

import (
	"encoding/json"

	sptypes "github.com/bluele/interchain-simple-packet/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
)

type (
	Header                    = sptypes.Header
	PacketData                = sptypes.PacketData
	PacketAcknowledgementData = PacketData
)

// NewPacketData returns a new packet data
func NewPacketData(h *Header, payload []byte) PacketData {
	if h == nil {
		h = &Header{}
	}
	return sptypes.NewSimplePacketData(*h, payload)
}

type PacketDataPayload interface {
	ValidateBasic() error
	GetBytes() []byte
	Type() string
}

type PacketAcknowledgementPayload interface {
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

func MarshalPacketAcknowledgementData(data PacketAcknowledgementData) ([]byte, error) {
	bz, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return bz, nil
}

func UnmarshalPacketAcknowledgementData(bz []byte, ad *PacketAcknowledgementData) error {
	return json.Unmarshal(bz, ad)
}

func NewPacketAcknowledgementData(h *Header, payload PacketAcknowledgementPayload) PacketAcknowledgementData {
	if h == nil {
		h = new(Header)
	}
	return PacketAcknowledgementData{
		Header:  *h,
		Payload: payload.GetBytes(),
	}
}

type IncomingPacketAcknowledgement interface {
	Data() PacketAcknowledgementData
	Header() Header
	Payload() PacketAcknowledgementPayload
}

type incomingPacketAcknowledgement struct {
	data    PacketAcknowledgementData
	payload PacketAcknowledgementPayload
}

var _ IncomingPacketAcknowledgement = (*incomingPacketAcknowledgement)(nil)

func NewIncomingPacketAcknowledgement(h *Header, payload PacketAcknowledgementPayload) IncomingPacketAcknowledgement {
	return incomingPacketAcknowledgement{data: NewPacketAcknowledgementData(h, payload), payload: payload}
}

func (a incomingPacketAcknowledgement) Data() PacketAcknowledgementData {
	return a.data
}

func (a incomingPacketAcknowledgement) Header() Header {
	return a.data.Header
}

func (a incomingPacketAcknowledgement) Payload() PacketAcknowledgementPayload {
	return a.payload
}

func UnmarshalIncomingPacketAcknowledgement(cdc *codec.Codec, bz []byte) (IncomingPacketAcknowledgement, error) {
	var pd PacketAcknowledgementData
	var payload PacketAcknowledgementPayload
	if err := UnmarshalPacketDataPayload(cdc, bz, &pd, &payload); err != nil {
		return nil, err
	}
	return NewIncomingPacketAcknowledgement(&pd.Header, payload), nil
}

type OutgoingPacketAcknowledgement interface {
	IncomingPacketAcknowledgement
	SetData(header Header, payload PacketAcknowledgementPayload)
}

type outgoingPacketAcknowledgement struct {
	data    PacketAcknowledgementData
	payload PacketAcknowledgementPayload
}

func NewOutgoingPacketAcknowledgement(h *Header, payload PacketAcknowledgementPayload) OutgoingPacketAcknowledgement {
	return &outgoingPacketAcknowledgement{
		data:    NewPacketAcknowledgementData(h, payload),
		payload: payload,
	}
}

var _ OutgoingPacketAcknowledgement = (*outgoingPacketAcknowledgement)(nil)

func (a outgoingPacketAcknowledgement) Data() PacketAcknowledgementData {
	return a.data
}

func (a outgoingPacketAcknowledgement) Header() Header {
	return a.data.Header
}

func (a outgoingPacketAcknowledgement) Payload() PacketAcknowledgementPayload {
	return a.payload
}

func (a outgoingPacketAcknowledgement) SetData(header Header, payload PacketAcknowledgementPayload) {
	a.data = NewPacketAcknowledgementData(&header, payload)
	a.payload = payload
}
