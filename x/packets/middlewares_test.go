package packets

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/ibc-go/modules/core/exported"
	"github.com/stretchr/testify/require"
)

var codecm codec.Codec

func TestMiddleware(t *testing.T) {
	require := require.New(t)
	m := NewCounterPacketMiddleware()
	ctx := makeMockContext()

	mps := &memPacketSender{}
	_, ps, err := m.HandleMsg(ctx, nil, mps)
	require.NoError(err)
	outp := newTestPacket(Header{}, &TestPacketDataPayload{})
	require.NoError(ps.SendPacket(ctx, nil, outp))
	count, ok := getCount(outp.pd.Header)
	require.True(ok)
	require.Equal(uint32(1), count)

	mps = &memPacketSender{}
	as := NewBasicACKSender()
	_, ps, as, err = m.HandlePacket(ctx, outp, mps, as)
	require.NoError(err)
	require.NoError(ps.SendPacket(ctx, nil, outp))
	count, ok = getCount(outp.pd.Header)
	require.True(ok)
	require.Equal(uint32(2), count)
}

func makeMockContext() sdk.Context {
	return sdk.Context{}
}

func (p TestPacketDataPayload) Type() string {
	return "cross/packets/test"
}

type testPacket struct {
	exported.PacketI
	pd      PacketData
	payload PacketDataPayload
}

func newTestPacket(header Header, payload PacketDataPayload) *testPacket {
	return &testPacket{
		pd:      NewPacketData(&header, payload),
		payload: payload,
	}
}

var _ IncomingPacket = (*testPacket)(nil)
var _ OutgoingPacket = (*testPacket)(nil)

func (p testPacket) PacketData() PacketData {
	return p.pd
}

func (p testPacket) Header() Header {
	return p.pd.Header
}

func (p testPacket) Payload() PacketDataPayload {
	return p.payload
}

func (p *testPacket) SetPacketData(header Header, payload PacketDataPayload) {
	*p = *newTestPacket(header, payload)
}

type memPacketSender struct {
	packet *OutgoingPacket
}

func (s *memPacketSender) SendPacket(
	ctx sdk.Context,
	channelCap *capabilitytypes.Capability,
	packet OutgoingPacket,
) error {
	s.packet = &packet
	return nil
}

type counterPacketMiddleware struct{}

var _ PacketMiddleware = (*counterPacketMiddleware)(nil)

// NewCounterPacketMiddleware returns counterPacketMiddleware
func NewCounterPacketMiddleware() PacketMiddleware {
	return counterPacketMiddleware{}
}

// HandleMsg implements PacketMiddleware.HandleMsg
func (m counterPacketMiddleware) HandleMsg(ctx sdk.Context, msg sdk.Msg, ps PacketSender) (sdk.Context, PacketSender, error) {
	return ctx, newPacketSender(1, ps), nil
}

// HandlePacket implements PacketMiddleware.HandlePacket
func (m counterPacketMiddleware) HandlePacket(ctx sdk.Context, ip IncomingPacket, ps PacketSender, as ACKSender) (sdk.Context, PacketSender, ACKSender, error) {
	var next uint32
	count, found := getCount(ip.Header())
	if found {
		next = count + 1
	} else {
		next = 1
	}
	return ctx, newPacketSender(next, ps), newACKSender(next, as), nil
}

// HandlePacket implements PacketMiddleware.HandleACK
func (m counterPacketMiddleware) HandleACK(ctx sdk.Context, ip IncomingPacket, ack IncomingPacketAcknowledgement, ps PacketSender) (sdk.Context, PacketSender, error) {
	return ctx, ps, nil
}

type packetSender struct {
	count uint32
	next  PacketSender
}

var _ PacketSender = (*packetSender)(nil)

func newPacketSender(count uint32, next PacketSender) PacketSender {
	return packetSender{count: count, next: next}
}

func (ps packetSender) SendPacket(
	ctx sdk.Context,
	channelCap *capabilitytypes.Capability,
	packet OutgoingPacket,
) error {
	h := packet.Header()
	setCount(&h, ps.count)
	packet.SetPacketData(h, packet.Payload())
	return ps.next.SendPacket(ctx, channelCap, packet)
}

type ackSender struct {
	count uint32
	next  ACKSender
}

var _ ACKSender = (*ackSender)(nil)

func newACKSender(count uint32, next ACKSender) ACKSender {
	return &ackSender{count: count, next: next}
}

func (as ackSender) SendACK(ctx sdk.Context, ack OutgoingPacketAcknowledgement) error {
	h := ack.Header()
	setCount(&h, as.count)
	ack.SetData(h, ack.Payload())
	return nil
}

const testHeaderKey = "count"

func setCount(h *Header, count uint32) {
	h.Set(testHeaderKey, []byte(fmt.Sprint(count)))
}

func getCount(h Header) (uint32, bool) {
	v, ok := h.Get(testHeaderKey)
	if !ok {
		return 0, false
	}
	i, err := strconv.Atoi(string(v))
	if err != nil {
		panic(err)
	}
	return uint32(i), true
}
