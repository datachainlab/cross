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

// PacketDataPayload defines the interface of packet data's payload
type PacketDataPayload interface {
	ValidateBasic() error
	GetBytes() []byte
	Type() string
}

// PacketAcknowledgementPayload defines the interface of packet ack's payload
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

func UnmarshalJSONPacketData(bz []byte, pd *PacketData) error {
	return json.Unmarshal(bz, pd)
}

func UnmarshalJSONPacketDataPayload(cdc *codec.Codec, bz []byte, pd *PacketData, ptr interface{}) error {
	if err := UnmarshalJSONPacketData(bz, pd); err != nil {
		return err
	}
	return cdc.UnmarshalJSON(pd.Payload, ptr)
}

func UnmarshalJSONIncomingPacket(cdc *codec.Codec, raw exported.PacketI) (IncomingPacket, error) {
	var pd PacketData
	var payload PacketDataPayload
	if err := UnmarshalJSONPacketDataPayload(cdc, raw.GetData(), &pd, &payload); err != nil {
		return nil, err
	}
	return NewIncomingPacket(raw, pd, payload), nil
}

// IncomingPacket defines the interface of incoming packet
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

// NewIncomingPacket returns a new IncomingPacket
func NewIncomingPacket(raw exported.PacketI, packetData PacketData, payload PacketDataPayload) IncomingPacket {
	return &incomingPacket{
		PacketI:    raw,
		packetData: packetData,
		payload:    payload,
	}
}

// PacketData implements IncomingPacket.PacketData
func (p incomingPacket) PacketData() PacketData {
	return p.packetData
}

// Header implements IncomingPacket.Header
func (p incomingPacket) Header() Header {
	return p.packetData.Header
}

// Payload implements IncomingPacket.Payload
func (p incomingPacket) Payload() PacketDataPayload {
	return p.payload
}

// OutgoingPacket defines the interface of outgoing packet
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

// NewOutgoingPacket returns a new OutgoingPacket
func NewOutgoingPacket(raw exported.PacketI, packetData PacketData, payload PacketDataPayload) OutgoingPacket {
	return &outgoingPacket{
		PacketI:    raw,
		packetData: packetData,
		payload:    payload,
	}
}

// PacketData implements Outgoing.PacketData
func (p outgoingPacket) PacketData() PacketData {
	return p.packetData
}

// Header implements Outgoing.Header
func (p outgoingPacket) Header() Header {
	return p.packetData.Header
}

// Payload implements Outgoing.Payload
func (p outgoingPacket) Payload() PacketDataPayload {
	return p.payload
}

// SetPacketData implements Outgoing.SetPacketData
func (p *outgoingPacket) SetPacketData(header Header, payload PacketDataPayload) {
	p.payload = payload
	p.packetData = NewPacketData(&header, payload.GetBytes())
}

// GetData implements Outgoing.GetData
func (p outgoingPacket) GetData() []byte {
	bz, err := MarshalPacketData(p.packetData)
	if err != nil {
		panic(err)
	}
	return bz
}

func MarshalJSONPacketAcknowledgementData(data PacketAcknowledgementData) ([]byte, error) {
	bz, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return bz, nil
}

func UnmarshalJSONPacketAcknowledgementData(bz []byte, ad *PacketAcknowledgementData) error {
	return json.Unmarshal(bz, ad)
}

// NewPacketAcknowledgementData returns a new PacketAcknowledgementData
func NewPacketAcknowledgementData(h *Header, payload PacketAcknowledgementPayload) PacketAcknowledgementData {
	if h == nil {
		h = new(Header)
	}
	return PacketAcknowledgementData{
		Header:  *h,
		Payload: payload.GetBytes(),
	}
}

// IncomingPacketAcknowledgement defines the interface of incoming packet acknowledgement
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

// NewIncomingPacketAcknowledgement returns a new IncomingPacketAcknowledgement
func NewIncomingPacketAcknowledgement(h *Header, payload PacketAcknowledgementPayload) IncomingPacketAcknowledgement {
	return incomingPacketAcknowledgement{data: NewPacketAcknowledgementData(h, payload), payload: payload}
}

// Data implements IncomingPacketAcknowledgement.Data
func (a incomingPacketAcknowledgement) Data() PacketAcknowledgementData {
	return a.data
}

// Header implements IncomingPacketAcknowledgement.Header
func (a incomingPacketAcknowledgement) Header() Header {
	return a.data.Header
}

// Payload implements IncomingPacketAcknowledgement.Payload
func (a incomingPacketAcknowledgement) Payload() PacketAcknowledgementPayload {
	return a.payload
}

func UnmarshalJSONIncomingPacketAcknowledgement(cdc *codec.Codec, bz []byte) (IncomingPacketAcknowledgement, error) {
	var pd PacketAcknowledgementData
	var payload PacketAcknowledgementPayload
	if err := UnmarshalJSONPacketDataPayload(cdc, bz, &pd, &payload); err != nil {
		return nil, err
	}
	return NewIncomingPacketAcknowledgement(&pd.Header, payload), nil
}

// OutgoingPacketAcknowledgement defines the interface of outgoing packet acknowledgement
type OutgoingPacketAcknowledgement interface {
	IncomingPacketAcknowledgement
	SetData(header Header, payload PacketAcknowledgementPayload)
}

type outgoingPacketAcknowledgement struct {
	data    PacketAcknowledgementData
	payload PacketAcknowledgementPayload
}

// NewOutgoingPacketAcknowledgement returns a new OutgoingPacketAcknowledgement
func NewOutgoingPacketAcknowledgement(h *Header, payload PacketAcknowledgementPayload) OutgoingPacketAcknowledgement {
	return &outgoingPacketAcknowledgement{
		data:    NewPacketAcknowledgementData(h, payload),
		payload: payload,
	}
}

var _ OutgoingPacketAcknowledgement = (*outgoingPacketAcknowledgement)(nil)

// Data implements OutgoingPacketAcknowledgement.Data
func (a outgoingPacketAcknowledgement) Data() PacketAcknowledgementData {
	return a.data
}

// Header implements OutgoingPacketAcknowledgement.Header
func (a outgoingPacketAcknowledgement) Header() Header {
	return a.data.Header
}

// Payload implements OutgoingPacketAcknowledgement.Payload
func (a outgoingPacketAcknowledgement) Payload() PacketAcknowledgementPayload {
	return a.payload
}

// Payload implements OutgoingPacketAcknowledgement.SetData
func (a outgoingPacketAcknowledgement) SetData(header Header, payload PacketAcknowledgementPayload) {
	a.data = NewPacketAcknowledgementData(&header, payload)
	a.payload = payload
}
