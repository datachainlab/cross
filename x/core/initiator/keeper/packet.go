package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	"github.com/datachainlab/cross/x/core/initiator/types"
	"github.com/datachainlab/cross/x/core/router"
	"github.com/datachainlab/cross/x/packets"
)

var _ router.PacketHandler = (*Keeper)(nil)

func (p Keeper) HandlePacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ip packets.IncomingPacket,
) (*sdk.Result, *packets.PacketAcknowledgementData, error) {
	ctx, _, as, err := p.packetMiddleware.HandlePacket(ctx, ip, packets.NewBasicPacketSender(p.ChannelKeeper()), packets.NewBasicACKSender())
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "failed to handle request: %v", err)
	}

	_, err = p.ReceiveIBCSignTx(
		ctx,
		packet.DestinationPort, packet.DestinationChannel,
		*ip.Payload().(*types.PacketDataIBCSignTx),
	)
	if err != nil {
		return nil, nil, err
	}

	// TODO fix status code
	ack := packets.NewOutgoingPacketAcknowledgement(
		p.m,
		nil,
		&types.PacketAcknowledgementIBCSignTx{Status: 0},
	)

	if err = as.SendACK(ctx, ack); err != nil {
		return nil, nil, err
	}

	ackData := ack.Data()
	return &sdk.Result{Data: nil, Events: ctx.EventManager().ABCIEvents()}, &ackData, nil
}

func (p Keeper) HandleACK(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ip packets.IncomingPacket,
	ipa packets.IncomingPacketAcknowledgement,
) (*sdk.Result, error) {
	return &sdk.Result{Data: nil, Events: ctx.EventManager().ABCIEvents()}, nil
}
