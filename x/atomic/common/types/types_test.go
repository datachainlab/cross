package types

import (
	"testing"

	crosstypes "github.com/datachainlab/cross/x/core/types"
	"github.com/stretchr/testify/require"
)

func TestCoordinatorState(t *testing.T) {
	channels := []crosstypes.ChannelInfo{
		{Port: "port0", Channel: "channel0"},
		{Port: "port1", Channel: "channel1"},
	}

	for _, tp := range []crosstypes.CommitProtocol{crosstypes.COMMIT_PROTOCOL_SIMPLE, crosstypes.COMMIT_PROTOCOL_TPC} {
		t.Run(tp.String(), func(t *testing.T) {
			cs := NewCoordinatorState(tp, COORDINATOR_PHASE_PREPARE, channels)
			require.False(t, cs.IsCompleted())
			require.False(t, cs.IsReceivedALLAcks())

			require.NoError(t, cs.Confirm(0, channels[0]))
			require.Error(t, cs.Confirm(0, channels[0]))
			require.False(t, cs.IsCompleted())

			require.NoError(t, cs.Confirm(1, channels[1]))
			require.True(t, cs.IsCompleted())

			cs.AddAck(0)
			require.False(t, cs.IsReceivedALLAcks())
			cs.AddAck(1)
			require.True(t, cs.IsReceivedALLAcks())
		})
	}

	// if `channels` is empty, constructor panics
	{
		require.Panics(t, func() {
			NewCoordinatorState(crosstypes.COMMIT_PROTOCOL_SIMPLE, COORDINATOR_PHASE_PREPARE, nil)
		})
	}
}
