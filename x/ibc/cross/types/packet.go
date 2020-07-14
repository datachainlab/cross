package types

type PacketData interface {
	ValidateBasic() error
	GetBytes() []byte
	GetTimeoutHeight() uint64
	GetTimeoutTimestamp() uint64
	Type() string
}

type PacketAcknowledgement interface {
	ValidateBasic() error
	GetBytes() []byte
	Type() string
}
