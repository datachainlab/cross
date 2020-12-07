package ibctesting

import (
	"errors"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// GetPacketsFromEvents parses events to returns packets
func GetPacketsFromEvents(events []abci.Event) ([]channeltypes.Packet, error) {
	var packets []channeltypes.Packet
	sevs := sdk.StringifyEvents(events)
	for _, ev := range sevs {
		if ev.Type == channeltypes.EventTypeSendPacket {
			// NOTE: Attributes of packet are included in one event.
			var packet channeltypes.Packet
			for _, attr := range ev.Attributes {
				switch attr.Key {
				case channeltypes.AttributeKeyData:
					// AttributeKeyData key indicates a start of packet attributes
					packet = channeltypes.Packet{}
					packet.Data = []byte(attr.Value)
				case channeltypes.AttributeKeyTimeoutHeight:
					parts := strings.Split(attr.Value, "-")
					packet.TimeoutHeight = clienttypes.NewHeight(
						strToUint64(parts[0]),
						strToUint64(parts[1]),
					)
				case channeltypes.AttributeKeyTimeoutTimestamp:
					packet.TimeoutTimestamp = strToUint64(attr.Value)
				case channeltypes.AttributeKeySequence:
					packet.Sequence = strToUint64(attr.Value)
				case channeltypes.AttributeKeySrcPort:
					packet.SourcePort = attr.Value
				case channeltypes.AttributeKeySrcChannel:
					packet.SourceChannel = attr.Value
				case channeltypes.AttributeKeyDstPort:
					packet.DestinationPort = attr.Value
				case channeltypes.AttributeKeyDstChannel:
					packet.DestinationChannel = attr.Value
				case channeltypes.AttributeKeyChannelOrdering:
					// AttributeKeyChannelOrdering key indicates the end of packet atrributes
					if err := packet.ValidateBasic(); err != nil {
						return nil, err
					}
					packets = append(packets, packet)
				}
			}
		}
	}
	return packets, nil
}

// GetACKFromEvents parses events to returns ack bytes
func GetACKFromEvents(events []abci.Event) ([]byte, error) {
	sevs := sdk.StringifyEvents(events)
	for _, ev := range sevs {
		if ev.Type == channeltypes.EventTypeWriteAck {
			for _, attr := range ev.Attributes {
				switch attr.Key {
				case channeltypes.AttributeKeyAck:
					return []byte(attr.Value), nil
				}
			}
		}
	}

	return nil, errors.New("ack not found")
}

func strToUint64(s string) uint64 {
	v, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return uint64(v)
}
