package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/rand"
)

// This is a trivial test for protobuf compatibility.
func TestMarshal(t *testing.T) {
	var bz [32]byte
	copy(bz[:], rand.NewRand().Bytes(32))
	dataB := HexByteArray32(bz)
	bz2, err := dataB.Marshal()
	assert.Nil(t, err)
	assert.Equal(t, bz[:], bz2)

	var dataB2 HexByteArray32
	err = (&dataB2).Unmarshal(bz[:])
	assert.Nil(t, err)
	assert.Equal(t, dataB, dataB2)
}

// Test that the hex encoding works.
func TestJSONMarshal(t *testing.T) {
	var bz [32]byte
	copy(bz[:], rand.NewRand().Bytes(32))
	dataB := HexByteArray32(bz)
	bz2, err := dataB.MarshalJSON()
	assert.Nil(t, err)

	var dataB2 HexByteArray32
	err = (&dataB2).UnmarshalJSON(bz2[:])
	assert.Nil(t, err)
	assert.Equal(t, dataB, dataB2)
}
