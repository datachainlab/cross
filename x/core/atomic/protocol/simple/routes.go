package simple

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	"github.com/datachainlab/cross/x/core/atomic/protocol/simple/keeper"
	"github.com/datachainlab/cross/x/core/atomic/protocol/simple/types"
	"github.com/datachainlab/cross/x/core/router"
	"github.com/datachainlab/cross/x/packets"
)

type PacketHandler struct {
	packetMiddleware packets.PacketMiddleware

	cdc    codec.Marshaler
	keeper keeper.Keeper
}

var _ router.PacketHandler = (*PacketHandler)(nil)

func NewPacketHandler(cdc codec.Marshaler, k keeper.Keeper, packetMiddleware packets.PacketMiddleware) PacketHandler {
	return PacketHandler{cdc: cdc, keeper: k, packetMiddleware: packetMiddleware}
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
	res, ap, err := h.keeper.ReceiveCallPacket(
		ctx,
		packet.DestinationPort, packet.DestinationChannel,
		*ip.Payload().(*types.PacketDataCall),
	)
	if err != nil {
		return nil, nil, err
	}

	ack := packets.NewOutgoingPacketAcknowledgement(
		nil,
		ap,
	)
	if err = as.SendACK(ctx, ack); err != nil {
		return nil, nil, err
	}
	ackData := ack.Data()
	return &sdk.Result{Data: res.GetData(), Events: ctx.EventManager().ABCIEvents()}, &ackData, nil
}

func (h PacketHandler) HandleACK(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ip packets.IncomingPacket,
	ipa packets.IncomingPacketAcknowledgement,
) (*sdk.Result, error) {
	ctx, _, err := h.packetMiddleware.HandleACK(ctx, ip, ipa, packets.NewBasicPacketSender(h.keeper.ChannelKeeper()))
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "failed to handle request: %v", err)
	}
	payload := ip.Payload().(*types.PacketDataCall)
	isCommitable, err := h.keeper.ReceiveCallAcknowledgement(ctx, packet.SourcePort, packet.SourceChannel, *ipa.Payload().(*types.PacketAcknowledgementCall), payload.TxId)
	if err != nil {
		return nil, err
	}
	res, err := h.keeper.TryCommit(ctx, payload.TxId, isCommitable)
	if err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvents(res.GetEvents())
	return &sdk.Result{Data: res.GetData(), Events: ctx.EventManager().ABCIEvents()}, nil
}
