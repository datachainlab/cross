package utils

import "encoding/binary"

// Uint32ToBigEndian - marshals uint32 to a bigendian byte slice so it can be sorted
func Uint32ToBigEndian(i uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, i)
	return b
}

// BigEndianToUint32 returns an uint32 from big endian encoded bytes. If encoding
// is empty, zero is returned.
func BigEndianToUint32(bz []byte) uint32 {
	if len(bz) == 0 {
		return 0
	}

	return binary.BigEndian.Uint32(bz)
}
