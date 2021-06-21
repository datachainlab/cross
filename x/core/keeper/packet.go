package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"

	"github.com/datachainlab/cross/x/packets"
)

func (p Keeper) ReceivePacket(ctx sdk.Context, packet channeltypes.Packet) (*sdk.Result, *packets.PacketAcknowledgementData, error) {
	ip, err := packets.UnmarshalIncomingPacket(p.m, packet)
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC packet type: %T: %v", packet, err)
	}
	route, found := p.router.GetRoute(ip.Payload().Type())
	if !found {
		return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "route not found: %v", ip.Payload().Type())
	}
	res, ack, err := route.HandlePacket(ctx, packet, ip)
	if err != nil {
		return nil, nil, err
	}
	return res, ack, nil
}

func (p Keeper) ReceivePacketAcknowledgementfunc(ctx sdk.Context, packet channeltypes.Packet, ack packets.IncomingPacketAcknowledgement) (*sdk.Result, error) {
	ip, err := packets.UnmarshalIncomingPacket(p.m, packet)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC packet type: %T: %v", packet, err)
	}
	route, found := p.router.GetRoute(ip.Payload().Type())
	if !found {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "route not found: %v", ip.Payload().Type())
	}
	res, err := route.HandleACK(ctx, packet, ip, ack)
	if err != nil {
		return nil, err
	}
	return res, nil
}
