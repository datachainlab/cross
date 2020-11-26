package tpc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	tpckeeper "github.com/datachainlab/cross/x/core/atomic/protocol/tpc/keeper"
	"github.com/datachainlab/cross/x/core/router"
	"github.com/datachainlab/cross/x/packets"
)

type PacketHandler struct {
	packetMiddleware packets.PacketMiddleware

	keeper tpckeeper.Keeper
}

var _ router.PacketHandler = (*PacketHandler)(nil)

func NewPacketHandler(k tpckeeper.Keeper) PacketHandler {
	return PacketHandler{keeper: k}
}

func (h PacketHandler) HandlePacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ip packets.IncomingPacket,
) (*sdk.Result, *packets.PacketAcknowledgementData, error) {
	panic("not implemented error")
}

func (h PacketHandler) HandleACK(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ip packets.IncomingPacket,
	ipa packets.IncomingPacketAcknowledgement,
) (*sdk.Result, error) {
	panic("not implemented error")
}
