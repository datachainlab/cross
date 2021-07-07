package packets

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/modules/core/24-host"
	"github.com/gogo/protobuf/proto"
)

type PacketSendKeeper struct {
	cdc codec.Codec

	channelKeeper ChannelKeeper
	portKeeper    PortKeeper
	scopedKeeper  capabilitykeeper.ScopedKeeper
}

func NewPacketSendKeeper(
	cdc codec.Codec,
	channelKeeper ChannelKeeper,
	portKeeper PortKeeper,
	scopedKeeper capabilitykeeper.ScopedKeeper,
) PacketSendKeeper {
	return PacketSendKeeper{
		cdc:           cdc,
		channelKeeper: channelKeeper,
		portKeeper:    portKeeper,
		scopedKeeper:  scopedKeeper,
	}
}

func (k PacketSendKeeper) SendPacket(
	ctx sdk.Context,
	packetSender PacketSender,
	payload PacketDataPayload,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) error {
	pd := NewPacketData(nil, payload)
	bz, err := proto.Marshal(&pd)
	if err != nil {
		return err
	}

	// get the next sequence
	seq, found := k.channelKeeper.GetNextSequenceSend(ctx, sourcePort, sourceChannel)
	if !found {
		return channeltypes.ErrSequenceSendNotFound
	}
	packet := channeltypes.NewPacket(
		bz,
		seq,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
		timeoutHeight,
		timeoutTimestamp,
	)
	channelCap, ok := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(sourcePort, sourceChannel))
	if !ok {
		return sdkerrors.Wrapf(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability '%v:%v'", sourcePort, sourceChannel)
	}

	if err := packetSender.SendPacket(ctx, channelCap, NewOutgoingPacket(packet, pd, payload)); err != nil {
		return err
	}
	return nil
}
