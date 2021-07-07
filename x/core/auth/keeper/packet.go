package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/modules/core/04-channel/types"
	"github.com/datachainlab/cross/x/core/auth/types"
	"github.com/datachainlab/cross/x/core/router"
	"github.com/datachainlab/cross/x/packets"
)

var _ router.PacketHandler = (*Keeper)(nil)

func (p Keeper) HandlePacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ip packets.IncomingPacket,
) (*sdk.Result, *packets.PacketAcknowledgementData, error) {
	ctx, _, as, err := p.packetMiddleware.HandlePacket(ctx, ip, packets.NewBasicPacketSender(p.channelKeeper), packets.NewBasicACKSender())
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "failed to handle request: %v", err)
	}

	var (
		status types.IBCSignTxStatus
		log    string
	)
	data := *ip.Payload().(*types.PacketDataIBCSignTx)
	completed, err := p.ReceiveIBCSignTx(
		ctx,
		packet.DestinationPort, packet.DestinationChannel,
		data,
	)
	switch {
	case err == nil && completed:
		status = types.IBC_SIGN_TX_STATUS_OK
		if err := p.txManager.OnPostAuth(ctx, data.TxID); err != nil {
			p.Logger(ctx).Error("failed to call PostAuth", "err", err)
			log = err.Error()
		}
	case err == nil:
		status = types.IBC_SIGN_TX_STATUS_OK
	default:
		status = types.IBC_SIGN_TX_STATUS_FAILED
		log = err.Error()
	}

	ack := packets.NewOutgoingPacketAcknowledgement(
		nil,
		&types.PacketAcknowledgementIBCSignTx{Status: status},
	)
	if err = as.SendACK(ctx, ack); err != nil {
		return nil, nil, err
	}

	ackData := ack.Data()
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents(), Log: log}, &ackData, nil
}

func (p Keeper) HandleACK(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ip packets.IncomingPacket,
	ipa packets.IncomingPacketAcknowledgement,
) (*sdk.Result, error) {
	// TODO handle an acknowledgement data
	return &sdk.Result{Data: nil, Events: ctx.EventManager().ABCIEvents()}, nil
}
