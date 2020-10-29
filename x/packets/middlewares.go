package packets

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
)

// PacketMiddleware defines middleware interface of handling packets
type PacketMiddleware interface {
	// HandleMsg handles sdk.Msg
	HandleMsg(ctx sdk.Context, msg sdk.Msg, ps PacketSender) (sdk.Context, PacketSender, error)
	// HandlePacket handles a packet
	HandlePacket(ctx sdk.Context, packet IncomingPacket, ps PacketSender, as ACKSender) (sdk.Context, PacketSender, ACKSender, error)
	// HandleACK handles a packet and packet ack
	HandleACK(ctx sdk.Context, packet IncomingPacket, ack IncomingPacketAcknowledgement, ps PacketSender) (sdk.Context, PacketSender, error)
}

// nopPacketMiddleware is middleware that does nothing
type nopPacketMiddleware struct{}

var _ PacketMiddleware = (*nopPacketMiddleware)(nil)

// NewNOPPacketMiddleware returns nopPacketMiddleware
func NewNOPPacketMiddleware() PacketMiddleware {
	return nopPacketMiddleware{}
}

// HandleMsg implements PacketMiddleware.HandleMsg
func (m nopPacketMiddleware) HandleMsg(ctx sdk.Context, msg sdk.Msg, ps PacketSender) (sdk.Context, PacketSender, error) {
	return ctx, ps, nil
}

// HandlePacket implements PacketMiddleware.HandlePacket
func (m nopPacketMiddleware) HandlePacket(ctx sdk.Context, packet IncomingPacket, ps PacketSender, as ACKSender) (sdk.Context, PacketSender, ACKSender, error) {
	return ctx, ps, as, nil
}

// HandlePacket implements PacketMiddleware.HandleACK
func (m nopPacketMiddleware) HandleACK(ctx sdk.Context, packet IncomingPacket, ack IncomingPacketAcknowledgement, ps PacketSender) (sdk.Context, PacketSender, error) {
	return ctx, ps, nil
}

// PacketSender defines packet-sender interface
type PacketSender interface {
	SendPacket(
		ctx sdk.Context,
		channelCap *capabilitytypes.Capability,
		packet OutgoingPacket,
	) error
}

type basicPacketSender struct {
	channelKeeper ChannelKeeper
}

// NewBasicPacketSender returns a new PacketSender instance
func NewBasicPacketSender(channelKeeper ChannelKeeper) PacketSender {
	return basicPacketSender{channelKeeper: channelKeeper}
}

func (s basicPacketSender) SendPacket(
	ctx sdk.Context,
	channelCap *capabilitytypes.Capability,
	packet OutgoingPacket,
) error {
	return s.channelKeeper.SendPacket(ctx, channelCap, packet)
}

// ACKSender defines packet acknowledgement sender interface
type ACKSender interface {
	SendACK(
		ctx sdk.Context,
		ack OutgoingPacketAcknowledgement,
	) error
}

type basicACKSender struct{}

var _ ACKSender = (*basicACKSender)(nil)

// NewBasicACKSender returns a new ACKSender instance
func NewBasicACKSender() ACKSender {
	return &basicACKSender{}
}

func (s *basicACKSender) SendACK(ctx sdk.Context, ack OutgoingPacketAcknowledgement) error {
	return nil
}
