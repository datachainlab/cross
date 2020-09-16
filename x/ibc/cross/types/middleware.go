package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
)

// PacketMiddleware defines middleware interface of handling packets
type PacketMiddleware interface {
	// HandleMsg handles sdk.Msg
	HandleMsg(ctx sdk.Context, msg sdk.Msg, sender PacketSender) (sdk.Context, PacketSender, error)
	// HandlePacket handles a packet
	HandlePacket(ctx sdk.Context, packet exported.PacketI, sender PacketSender) (sdk.Context, PacketSender, error)
	// HandleACK handles a packet and packet ack
	HandleACK(ctx sdk.Context, packet exported.PacketI, ack []byte, sender PacketSender) (sdk.Context, PacketSender, error)
}

// PacketSender defines packet-sender interface
type PacketSender interface {
	SendPacket(
		ctx sdk.Context,
		channelCap *capabilitytypes.Capability,
		packet exported.PacketI,
	) error
}

// nopPacketMiddleware is middleware that does nothing
type nopPacketMiddleware struct{}

var _ PacketMiddleware = (*nopPacketMiddleware)(nil)

// NewNOPPacketMiddleware returns nopPacketMiddleware
func NewNOPPacketMiddleware() PacketMiddleware {
	return nopPacketMiddleware{}
}

// HandleMsg implements PacketMiddleware.HandleMsg
func (m nopPacketMiddleware) HandleMsg(ctx sdk.Context, msg sdk.Msg, sender PacketSender) (sdk.Context, PacketSender, error) {
	return ctx, sender, nil
}

// HandlePacket implements PacketMiddleware.HandlePacket
func (m nopPacketMiddleware) HandlePacket(ctx sdk.Context, packet exported.PacketI, sender PacketSender) (sdk.Context, PacketSender, error) {
	return ctx, sender, nil
}

// HandlePacket implements PacketMiddleware.HandleACK
func (m nopPacketMiddleware) HandleACK(ctx sdk.Context, packet exported.PacketI, ack []byte, sender PacketSender) (sdk.Context, PacketSender, error) {
	return ctx, sender, nil
}
