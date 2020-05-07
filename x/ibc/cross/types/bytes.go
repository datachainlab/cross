package types

import (
	"encoding/hex"
	"fmt"
	"strings"
)

type HexByteArray32 [32]byte

// Marshal needed for protobuf compatibility
func (bz HexByteArray32) Marshal() ([]byte, error) {
	return bz[:], nil
}

// Unmarshal needed for protobuf compatibility
func (bz *HexByteArray32) Unmarshal(data []byte) error {
	if l := len(data); l != 32 {
		return fmt.Errorf("%v != %v", l, 32)
	}
	copy(bz[:], data)
	return nil
}

// This is the point of Bytes.
func (bz HexByteArray32) MarshalJSON() ([]byte, error) {
	s := strings.ToUpper(hex.EncodeToString(bz[:]))
	jbz := make([]byte, len(s)+2)
	jbz[0] = '"'
	copy(jbz[1:], []byte(s))
	jbz[len(jbz)-1] = '"'
	return jbz, nil
}

// This is the point of Bytes.
func (bz *HexByteArray32) UnmarshalJSON(data []byte) error {
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("invalid hex string: %s", data)
	}
	bz2, err := hex.DecodeString(string(data[1 : len(data)-1]))
	if err != nil {
		return err
	}
	copy(bz[:], bz2)
	return nil
}

// Allow it to fulfill various interfaces in light-client, etc...
func (bz HexByteArray32) Bytes() []byte {
	return bz[:]
}

func (bz HexByteArray32) String() string {
	return strings.ToUpper(hex.EncodeToString(bz[:]))
}

func (bz *HexByteArray32) FromString(s string) error {
	b, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	copy(bz[:], b)
	return nil
}

func (bz HexByteArray32) Format(s fmt.State, verb rune) {
	switch verb {
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", bz)))
	default:
		s.Write([]byte(fmt.Sprintf("%X", []byte(bz[:]))))
	}
}
