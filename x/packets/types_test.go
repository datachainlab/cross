package packets

var _ PacketDataPayload = (*TestPacketDataPayload)(nil)

func (TestPacketDataPayload) ValidateBasic() error {
	return nil
}
