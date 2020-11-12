package core

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"

	simplekeeper "github.com/datachainlab/cross/x/atomic/simple/keeper"
	simpletypes "github.com/datachainlab/cross/x/atomic/simple/types"
	"github.com/datachainlab/cross/x/core/keeper"
	"github.com/datachainlab/cross/x/core/types"
	"github.com/datachainlab/cross/x/packets"
)

// NewHandler ...
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case *types.MsgInitiate:
			res, err := k.Initiate(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}

// PacketReceiver is a receiver that handles a packet
type PacketReceiver func(ctx sdk.Context, packet channeltypes.Packet) (*sdk.Result, *packets.PacketAcknowledgementData, error)

// NewPacketReceiver returns a new PacketReceiver
func NewPacketReceiver(cdc codec.Marshaler, keeper keeper.Keeper, packetMiddleware packets.PacketMiddleware) PacketReceiver {
	return func(ctx sdk.Context, packet channeltypes.Packet) (*sdk.Result, *packets.PacketAcknowledgementData, error) {
		ip, err := packets.UnmarshalJSONIncomingPacket(cdc, packet)
		if err != nil {
			return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC packet type: %T: %v", packet, err)
		}
		ctx, _, as, err := packetMiddleware.HandlePacket(ctx, ip, packets.NewBasicPacketSender(keeper.ChannelKeeper()), packets.NewBasicACKSender())
		if err != nil {
			return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "failed to handle request: %v", err)
		}
		var resData []byte
		var ack packets.OutgoingPacketAcknowledgement
		switch payload := ip.Payload().(type) {
		case *simpletypes.PacketDataCall:
			resData, ack, err = handlePacketDataCall(ctx, cdc, keeper.SimpleKeeper(), packet, *payload)
		default:
			return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC packet payload type: %T", payload)
		}
		if err != nil {
			return nil, nil, err
		}
		if err = as.SendACK(ctx, ack); err != nil {
			return nil, nil, err
		}
		ackData := ack.Data()
		return &sdk.Result{Data: resData, Events: ctx.EventManager().ABCIEvents()}, &ackData, nil
	}
}

func handlePacketDataCall(ctx sdk.Context, cdc codec.Marshaler, k simplekeeper.Keeper, packet channeltypes.Packet, payload simpletypes.PacketDataCall) ([]byte, packets.OutgoingPacketAcknowledgement, error) {
	res, ackData, err := k.ReceiveCallPacket(
		ctx,
		packet.DestinationPort, packet.DestinationChannel,
		payload,
	)
	if err != nil {
		return nil, nil, err
	}

	ack := packets.NewOutgoingPacketAcknowledgement(
		cdc,
		nil,
		ackData,
	)
	return res.Data, ack, nil
}

// PacketAcknowledgementReceiver is a packet acknowledgement receiver
type PacketAcknowledgementReceiver func(ctx sdk.Context, packet channeltypes.Packet, ack packets.IncomingPacketAcknowledgement) (*sdk.Result, error)

// NewPacketAcknowledgementReceiver returns a PacketAcknowledgementReceiver
func NewPacketAcknowledgementReceiver(cdc codec.Marshaler, keeper keeper.Keeper, packetMiddleware packets.PacketMiddleware) PacketAcknowledgementReceiver {
	return func(ctx sdk.Context, packet channeltypes.Packet, ack packets.IncomingPacketAcknowledgement) (*sdk.Result, error) {
		pi, err := packets.UnmarshalJSONIncomingPacket(cdc, packet)
		if err != nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC packet type: %T: %v", packet, err)
		}
		ctx, ps, err := packetMiddleware.HandleACK(ctx, pi, ack, packets.NewBasicPacketSender(keeper.ChannelKeeper()))
		if err != nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "failed to handle request: %v", err)
		}
		_ = ps
		var resData []byte
		switch ack := ack.Payload().(type) {
		case *simpletypes.PacketCallAcknowledgement:
			payload := pi.Payload().(*simpletypes.PacketDataCall)
			resData, err = handlePacketCallAcknowledgement(ctx, keeper.SimpleKeeper(), packet, *ack, *payload)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC ack type: %T", ack)
		}
		if err != nil {
			return nil, err
		}

		return &sdk.Result{Data: resData, Events: ctx.EventManager().ABCIEvents()}, nil
	}
}

func handlePacketCallAcknowledgement(ctx sdk.Context, k simplekeeper.Keeper, packet channeltypes.Packet, ack simpletypes.PacketCallAcknowledgement, payload simpletypes.PacketDataCall) ([]byte, error) {
	isCommitable, err := k.ReceiveCallAcknowledgement(ctx, packet.SourcePort, packet.SourceChannel, ack, payload.TxId)
	if err != nil {
		return nil, err
	}
	res, err := k.TryCommit(ctx, payload.TxId, isCommitable)
	if err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvents(res.GetEvents())
	return res.Data, nil
}
