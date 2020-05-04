package types

import (
	"errors"
	"strconv"
	"strings"
)

const (
	QueryCoordinatorStatus     = "coordinator_status"
	QueryUnacknowledgedPackets = "unacknowledged_packets"
)

type QueryCoordinatorStatusRequest struct {
	TxID TxID `json:"tx_id" yaml:"tx_id"`
}

type QueryCoordinatorStatusResponse struct {
	TxID            TxID            `json:"tx_id" yaml:"tx_id"`
	CoordinatorInfo CoordinatorInfo `json:"coordinator_info" yaml:"coordinator_info"`
	Completed       bool            `json:"completed" yaml:"completed"`
}

type QueryUnacknowledgedPacketsRequest struct{}

type QueryUnacknowledgedPacketsResponse struct {
	Packets []UnacknowledgedPacket `json:"packets" yaml:"packets"`
}

type UnacknowledgedPacket struct {
	Sequence      uint64 `json:"sequence" yaml:"sequence"`             // number corresponds to the order of sends and receives, where a Packet with an earlier sequence number must be sent and received before a Packet with a later sequence number.
	SourcePort    string `json:"source_port" yaml:"source_port"`       // identifies the port on the sending chain.
	SourceChannel string `json:"source_channel" yaml:"source_channel"` // identifies the channel end on the sending chain.
}

func ParseUnacknowledgedPacketKey(key []byte) (*UnacknowledgedPacket, error) {
	parts := strings.Split(string(key), "/")
	if len(parts) != 3 {
		return nil, errors.New("length of parts must be 3")
	}
	seq, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, err
	}
	return &UnacknowledgedPacket{
		Sequence:      uint64(seq),
		SourcePort:    string(parts[0]),
		SourceChannel: string(parts[1]),
	}, nil
}
