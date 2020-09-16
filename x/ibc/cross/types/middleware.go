package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
)

type PacketSender interface {
	SendPacket(
		ctx sdk.Context,
		channelCap *capabilitytypes.Capability,
		packet exported.PacketI,
	) error
}

type PacketMiddleware interface {
	HandleMsg(ctx sdk.Context, msg sdk.Msg, sender PacketSender) (sdk.Context, PacketSender, error)
	HandlePacket(ctx sdk.Context, packet exported.PacketI, sender PacketSender) (sdk.Context, PacketSender, error)
	HandleACK(ctx sdk.Context, packet exported.PacketI, ack []byte, sender PacketSender) (sdk.Context, PacketSender, error)
}

type nopPacketMiddleware struct{}

var _ PacketMiddleware = (*nopPacketMiddleware)(nil)

func NewNOPPacketMiddleware() PacketMiddleware {
	return nopPacketMiddleware{}
}

func (m nopPacketMiddleware) HandleMsg(ctx sdk.Context, msg sdk.Msg, sender PacketSender) (sdk.Context, PacketSender, error) {
	return ctx, sender, nil
}

func (m nopPacketMiddleware) HandlePacket(ctx sdk.Context, packet exported.PacketI, sender PacketSender) (sdk.Context, PacketSender, error) {
	return ctx, sender, nil
}

func (m nopPacketMiddleware) HandleACK(ctx sdk.Context, packet exported.PacketI, ack []byte, sender PacketSender) (sdk.Context, PacketSender, error) {
	return ctx, sender, nil
}
