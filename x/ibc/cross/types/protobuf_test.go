package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	tmtypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tendermint/version"
)

// A copy from
// https://github.com/tendermint/tendermint/blob/master/types/protobuf_test.go#L84
func TestABCIToTM(t *testing.T) {
	assert := assert.New(t)

	// build a full header
	var height int64 = 5
	header := newHeader(height, []byte("lastCommitHash"), []byte("dataHash"), []byte("evidenceHash"))
	protocolVersion := version.Consensus{Block: 7, App: 8}
	timestamp := time.Now()
	lastBlockID := tmtypes.BlockID{
		Hash: []byte("hash"),
		PartsHeader: tmtypes.PartSetHeader{
			Total: 10,
			Hash:  []byte("hash"),
		},
	}
	header.Populate(
		protocolVersion, "chainID", timestamp, lastBlockID,
		[]byte("valHash"), []byte("nextValHash"),
		[]byte("consHash"), []byte("appHash"), []byte("lastResultsHash"),
		[]byte("proposerAddress"),
	)
	pbHeader := tmtypes.TM2PB.Header(header)
	h := MakeHashFromABCIHeader(pbHeader)
	assert.Equal(header, h)
}

func newHeader(
	height int64, commitHash, dataHash, evidenceHash []byte,
) *tmtypes.Header {
	return &tmtypes.Header{
		Height:         height,
		LastCommitHash: commitHash,
		DataHash:       dataHash,
		EvidenceHash:   evidenceHash,
	}
}
