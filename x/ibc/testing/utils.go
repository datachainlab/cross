package ibctesting

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/datachainlab/cross/x/packets"
)

type capturePacketSender struct {
	inner   packets.PacketSender
	packets []packets.OutgoingPacket
}

var _ packets.PacketSender = (*capturePacketSender)(nil)

func NewCapturePacketSender(ps packets.PacketSender) *capturePacketSender {
	return &capturePacketSender{inner: ps}
}

func (ps *capturePacketSender) SendPacket(
	ctx sdk.Context,
	channelCap *capabilitytypes.Capability,
	packet packets.OutgoingPacket,
) error {
	if err := ps.inner.SendPacket(ctx, channelCap, packet); err != nil {
		return err
	}
	ps.packets = append(ps.packets, packet)
	return nil
}

func (ps *capturePacketSender) Packets() []packets.OutgoingPacket {
	return ps.packets
}
