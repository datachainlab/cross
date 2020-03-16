package contract

import (
	"fmt"
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/assert"
)

func TestType(t *testing.T) {
	var cases = []struct {
		f interface{}
	}{
		{
			f: func(v bool) bool {
				return v == Bool(ToBytes(v))
			},
		},
		{
			f: func(v int8) bool {
				return v == Int8(ToBytes(v))
			},
		},
		{
			f: func(v int16) bool {
				return v == Int16(ToBytes(v))
			},
		},
		{
			f: func(v int32) bool {
				return v == Int32(ToBytes(v))
			},
		},
		{
			f: func(v int64) bool {
				return v == Int64(ToBytes(v))
			},
		},
		{
			f: func(v uint8) bool {
				return v == UInt8(ToBytes(v))
			},
		},
		{
			f: func(v uint16) bool {
				return v == UInt16(ToBytes(v))
			},
		},
		{
			f: func(v uint32) bool {
				return v == UInt32(ToBytes(v))
			},
		},
		{
			f: func(v uint64) bool {
				return v == UInt64(ToBytes(v))
			},
		},
	}

	for i, cs := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			assert := assert.New(t)
			err := quick.Check(cs.f, &quick.Config{MaxCount: 1000})
			assert.NoError(err)
		})
	}
}
