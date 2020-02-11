package lock

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptimizeOPs(t *testing.T) {
	var R = func(k string) Read {
		return Read{K: []byte(k)}
	}
	var W = func(k, v string) Write {
		return Write{K: []byte(k), V: []byte(v)}
	}

	var cases = []struct {
		ops      OPs
		expected OPs
	}{
		{
			OPs{R("k0"), R("k1")},
			OPs{R("k0"), R("k1")},
		},
		{
			OPs{R("k0"), W("k0", "v0")},
			OPs{W("k0", "v0")},
		},
		{
			OPs{W("k0", "v0"), R("k0")},
			OPs{W("k0", "v0")},
		},
		{
			OPs{R("k0"), W("k0", "v0"), R("k0")},
			OPs{W("k0", "v0")},
		},
		{
			OPs{R("k0"), W("k0", "v0"), W("k0", "v1"), R("k0")},
			OPs{W("k0", "v1")},
		},
		{
			OPs{W("k0", "v0"), W("k1", "v1")},
			OPs{W("k0", "v0"), W("k1", "v1")},
		},
	}

	for i, cs := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			assert := assert.New(t)
			actual := OptimizeOPs(cs.ops)
			assert.Equal(cs.expected, actual)
		})
	}
}
