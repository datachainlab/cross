package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLockManager(t *testing.T) {
	type input struct {
		K []byte
		V []byte
	}

	var W = func(k, v string) input {
		return input{[]byte(k), []byte(v)}
	}
	var L = func(k, v string) LockOP {
		return LockOP{K: []byte(k), V: []byte(v)}
	}

	var cases = []struct {
		inputs  []input
		LockOPs []LockOP
	}{
		{
			[]input{},
			[]LockOP{},
		},
		{
			[]input{W("k1", "v1-1")},
			[]LockOP{L("k1", "v1-1")},
		},
		{
			[]input{W("k1", "v1-1"), W("k2", "v2-1")},
			[]LockOP{L("k1", "v1-1"), L("k2", "v2-1")},
		},
		{
			[]input{W("k1", "v1-1"), W("k1", "v1-2")},
			[]LockOP{L("k1", "v1-2")},
		},
	}

	for i, cs := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			require := require.New(t)
			m := NewLockManager()
			for _, in := range cs.inputs {
				require.NoError(m.AddWrite(in.K, in.V))
			}
			require.Equal(LockOPs{Ops: cs.LockOPs}, m.LockOPs())
		})
	}
}
