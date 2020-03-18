package contract

import (
	"encoding/binary"
	"fmt"
)

func Bool(v []byte) bool {
	if l := len(v); l == 0 {
		return false
	} else if l > 1 {
		panic(fmt.Sprintf("length of v should be 0 or 1, but got %v", l))
	}

	if v[0] == 0 {
		return false
	} else {
		return true
	}
}

func Int8(v []byte) int8 {
	if l := len(v); l != 1 {
		panic(fmt.Sprintf("length of v should be 1, but got %v", l))
	}
	return int8(v[0])
}

func Int16(v []byte) int16 {
	return int16(binary.BigEndian.Uint16(v))
}

func Int32(v []byte) int32 {
	return int32(binary.BigEndian.Uint32(v))
}

func Int64(v []byte) int64 {
	return int64(binary.BigEndian.Uint64(v))
}

func UInt8(v []byte) uint8 {
	if l := len(v); l != 1 {
		panic(fmt.Sprintf("length of v should be 1, but got %v", l))
	}
	return v[0]
}

func UInt16(v []byte) uint16 {
	return binary.BigEndian.Uint16(v)
}

func UInt32(v []byte) uint32 {
	return binary.BigEndian.Uint32(v)
}

func UInt64(v []byte) uint64 {
	return binary.BigEndian.Uint64(v)
}

func ToBytes(v interface{}) []byte {
	switch v := v.(type) {
	case bool:
		if v {
			return []byte{1}
		} else {
			return []byte{0}
		}
	case int8:
		return []byte{uint8(v)}
	case int16:
		var bz [2]byte
		binary.BigEndian.PutUint16(bz[:], uint16(v))
		return bz[:]
	case int32:
		var bz [4]byte
		binary.BigEndian.PutUint32(bz[:], uint32(v))
		return bz[:]
	case int64:
		var bz [8]byte
		binary.BigEndian.PutUint64(bz[:], uint64(v))
		return bz[:]
	case uint8:
		return []byte{v}
	case uint16:
		var bz [2]byte
		binary.BigEndian.PutUint16(bz[:], uint16(v))
		return bz[:]
	case uint32:
		var bz [4]byte
		binary.BigEndian.PutUint32(bz[:], v)
		return bz[:]
	case uint64:
		var bz [8]byte
		binary.BigEndian.PutUint64(bz[:], v)
		return bz[:]
	default:
		panic(fmt.Sprintf("unknown type: %T", v))
	}
}
