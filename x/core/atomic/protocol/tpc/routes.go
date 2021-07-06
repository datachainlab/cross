package tpc

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/modules/core/04-channel/types"

	tpckeeper "github.com/datachainlab/cross/x/core/atomic/protocol/tpc/keeper"
	"github.com/datachainlab/cross/x/core/atomic/protocol/tpc/types"
	"github.com/datachainlab/cross/x/core/router"
	"github.com/datachainlab/cross/x/packets"
)

type PacketHandler struct {
	packetMiddleware packets.PacketMiddleware

	cdc    codec.Codec
	keeper tpckeeper.Keeper
}

var _ router.PacketHandler = (*PacketHandler)(nil)

func NewPacketHandler(cdc codec.Codec, k tpckeeper.Keeper) PacketHandler {
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

	var (
		data []byte
		ack  packets.OutgoingPacketAcknowledgement
	)
	switch payload := ip.Payload().(type) {
	case *types.PacketDataPrepare:
		res, ap, err := h.keeper.ReceivePacketPrepare(
			ctx,
			packet.DestinationPort, packet.DestinationChannel,
			*payload,
		)
		if err != nil {
			return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "failed to ReceivePreparePacket: %v", err)
		}
		ack = packets.NewOutgoingPacketAcknowledgement(nil, ap)
		data = res.GetData()
	case *types.PacketDataCommit:
		res, ap, err := h.keeper.ReceivePacketCommit(
			ctx,
			packet.DestinationPort, packet.DestinationChannel,
			*payload,
		)
		if err != nil {
			return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "failed to ReceivePacketCommit: %v", err)
		}
		ack = packets.NewOutgoingPacketAcknowledgement(nil, ap)
		if res != nil {
			data = res.GetData()
			ctx.EventManager().EmitEvents(res.GetEvents())
		} else {
			data = nil
		}
	default:
		return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized packet type: %T", payload)
	}
	if err = as.SendACK(ctx, ack); err != nil {
		return nil, nil, err
	}
	ackData := ack.Data()
	return &sdk.Result{Data: data, Events: ctx.EventManager().ABCIEvents()}, &ackData, nil
}

func (h PacketHandler) HandleACK(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ip packets.IncomingPacket,
	ipa packets.IncomingPacketAcknowledgement,
) (*sdk.Result, error) {
	ctx, ps, err := h.packetMiddleware.HandleACK(ctx, ip, ipa, packets.NewBasicPacketSender(h.keeper.ChannelKeeper()))
	if err != nil {
		return nil, err
	}

	switch payload := ipa.Payload().(type) {
	case *types.PacketAcknowledgementPrepare:
		pd := ip.Payload().(*types.PacketDataPrepare)
		return h.keeper.HandlePacketAcknowledgementPrepare(
			ctx,
			packet.SourcePort, packet.SourceChannel,
			*payload, pd.TxId, pd.TxIndex, ps,
		)
	case *types.PacketAcknowledgementCommit:
		bz := h.cdc.MustMarshalJSON(payload)
		return &sdk.Result{Data: bz, Events: ctx.EventManager().ABCIEvents()}, nil
	default:
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized ack type: %T", payload)
	}
}
