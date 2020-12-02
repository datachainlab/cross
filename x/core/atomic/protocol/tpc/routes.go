package tpc

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"

	tpckeeper "github.com/datachainlab/cross/x/core/atomic/protocol/tpc/keeper"
	"github.com/datachainlab/cross/x/core/atomic/protocol/tpc/types"
	"github.com/datachainlab/cross/x/core/router"
	"github.com/datachainlab/cross/x/packets"
)

type PacketHandler struct {
	packetMiddleware packets.PacketMiddleware

	cdc    codec.Marshaler
	keeper tpckeeper.Keeper
}

var _ router.PacketHandler = (*PacketHandler)(nil)

func NewPacketHandler(cdc codec.Marshaler, k tpckeeper.Keeper) PacketHandler {
	return PacketHandler{cdc: cdc, keeper: k}
}

func (h PacketHandler) HandlePacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ip packets.IncomingPacket,
) (*sdk.Result, *packets.PacketAcknowledgementData, error) {
	ctx, _, as, err := h.packetMiddleware.HandlePacket(ctx, ip, packets.NewBasicPacketSender(h.keeper.ChannelKeeper()), packets.NewBasicACKSender())
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "failed to handle request: %v", err)
	}

	// TODO add multiple packet type support
	res, ap, err := h.keeper.ReceivePreparePacket(
		ctx,
		packet.DestinationPort, packet.DestinationChannel,
		*ip.Payload().(*types.PacketDataPrepare),
	)
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "failed to ReceivePreparePacket: %v", err)
	}

	ack := packets.NewOutgoingPacketAcknowledgement(
		h.cdc,
		nil,
		ap,
	)
	if err = as.SendACK(ctx, ack); err != nil {
		return nil, nil, err
	}
	ackData := ack.Data()
	return &sdk.Result{Data: res.Data, Events: ctx.EventManager().ABCIEvents()}, &ackData, nil
}

func (h PacketHandler) HandleACK(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ip packets.IncomingPacket,
	ipa packets.IncomingPacketAcknowledgement,
) (*sdk.Result, error) {
	panic("not implemented error")
}
