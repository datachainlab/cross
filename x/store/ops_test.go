package store

import (
	"fmt"
	"testing"

	crosstypes "github.com/datachainlab/cross/x/core/types"
	"github.com/stretchr/testify/assert"
)

func TestOPManager(t *testing.T) {
	var R = func(k, v string) *ReadOP {
		return &ReadOP{[]byte(k), []byte(v)}
	}
	var RNIL = func() *ReadOP {
		return &ReadOP{nil, nil}
	}
	var W = func(k, v string) *WriteOP {
		return &WriteOP{[]byte(k), []byte(v)}
	}
	var WNIL = func() *WriteOP {
		return &WriteOP{nil, nil}
	}
	var OPs = func(items ...OP) crosstypes.OPs {
		ops, err := convertOPItemsToOPs(items)
		if err != nil {
			panic(err)
		}
		return *ops
	}

	var cases = []struct {
		constraintType crosstypes.StateConstraintType
		inputs         []OP
		LockOPs        []LockOP
		OPs            crosstypes.OPs
	}{
		// Test for ExactMatchStateConstraint
		{
			crosstypes.ExactMatchStateConstraint,
			[]OP{},
			[]LockOP{},
			OPs(),
		},
		{
			crosstypes.ExactMatchStateConstraint,
			[]OP{R("k1", "v1")},
			[]LockOP{},
			OPs(R("k1", "v1")),
		},
		{
			crosstypes.ExactMatchStateConstraint,
			[]OP{R("k1", "v1"), W("k1", "v1-1")},
			[]LockOP{W("k1", "v1-1")},
			OPs(R("k1", "v1"), W("k1", "v1-1")),
		},
		{
			crosstypes.ExactMatchStateConstraint,
			[]OP{R("k1", "v1"), W("k2", "v2")},
			[]LockOP{W("k2", "v2")},
			OPs(R("k1", "v1"), W("k2", "v2")),
		},
		{
			crosstypes.ExactMatchStateConstraint,
			[]OP{R("k1", "v1"), W("k2", "v2"), W("k1", "v1-1")},
			[]LockOP{W("k2", "v2"), W("k1", "v1-1")},
			OPs(R("k1", "v1"), W("k2", "v2"), W("k1", "v1-1")),
		},
		// Test for PreStateConstraint
		{
			crosstypes.PreStateConstraint,
			[]OP{},
			[]LockOP{},
			OPs(),
		},
		{
			crosstypes.PreStateConstraint,
			[]OP{R("k1", "v1")},
			[]LockOP{},
			OPs(R("k1", "v1")),
		},
		{
			crosstypes.PreStateConstraint,
			[]OP{R("k1", "v1"), W("k1", "v1-1")},
			[]LockOP{W("k1", "v1-1")},
			OPs(R("k1", "v1")),
		},
		{
			crosstypes.PreStateConstraint,
			[]OP{R("k1", "v1"), W("k2", "v2")},
			[]LockOP{W("k2", "v2")},
			OPs(R("k1", "v1")),
		},
		{
			crosstypes.PreStateConstraint,
			[]OP{R("k1", "v1"), W("k2", "v2"), W("k1", "v1-1")},
			[]LockOP{W("k2", "v2"), W("k1", "v1-1")},
			OPs(R("k1", "v1")),
		},
		// Test for PostStateConstraint
		{
			crosstypes.PostStateConstraint,
			[]OP{},
			[]LockOP{},
			OPs(),
		},
		{
			crosstypes.PostStateConstraint,
			[]OP{R("k1", "v1")},
			[]LockOP{},
			OPs(),
		},
		{
			crosstypes.PostStateConstraint,
			[]OP{R("k1", "v1"), W("k1", "v1-1")},
			[]LockOP{W("k1", "v1-1")},
			OPs(W("k1", "v1-1")),
		},
		{
			crosstypes.PostStateConstraint,
			[]OP{R("k1", "v1"), W("k2", "v2")},
			[]LockOP{W("k2", "v2")},
			OPs(W("k2", "v2")),
		},
		{
			crosstypes.PostStateConstraint,
			[]OP{R("k1", "v1"), W("k2", "v2"), W("k1", "v1-1")},
			[]LockOP{W("k2", "v2"), W("k1", "v1-1")},
			OPs(W("k2", "v2"), W("k1", "v1-1")),
		},
		// Test for NoStateConstraint
		{
			crosstypes.NoStateConstraint,
			[]OP{},
			[]LockOP{},
			OPs(),
		},
		{
			crosstypes.NoStateConstraint,
			[]OP{R("k1", "v1")},
			[]LockOP{},
			OPs(),
		},
		{
			crosstypes.NoStateConstraint,
			[]OP{R("k1", "v1"), W("k1", "v1-1")},
			[]LockOP{W("k1", "v1-1")},
			OPs(),
		},
		{
			crosstypes.NoStateConstraint,
			[]OP{R("k1", "v1"), W("k2", "v2")},
			[]LockOP{W("k2", "v2")},
			OPs(),
		},
		{
			crosstypes.NoStateConstraint,
			[]OP{R("k1", "v1"), W("k2", "v2"), W("k1", "v1-1")},
			[]LockOP{W("k2", "v2"), W("k1", "v1-1")},
			OPs(),
		},
		// Test for OP
		{
			crosstypes.ExactMatchStateConstraint,
			[]OP{RNIL()},
			[]LockOP{},
			OPs(RNIL()),
		},
		{
			crosstypes.ExactMatchStateConstraint,
			[]OP{WNIL()},
			[]LockOP{WNIL()},
			OPs(WNIL()),
		},
		{
			crosstypes.ExactMatchStateConstraint,
			[]OP{WNIL(), WNIL()},
			[]LockOP{WNIL()},
			OPs(WNIL(), WNIL()),
		},
	}

	for i, cs := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			assert := assert.New(t)
			m, err := getOPManager(cs.constraintType)
			if err != nil {
				assert.FailNow(err.Error())
			}
			for _, in := range cs.inputs {
				if tp := in.Type(); tp == OpTypeRead {
					m.AddRead(in.(*ReadOP).K, in.(*ReadOP).V)
				} else if tp == OpTypeWrite {
					m.AddWrite(in.(*WriteOP).K, in.(*WriteOP).V)
				} else {
					assert.FailNow("fatal error")
				}
			}
			assert.Equal(cs.LockOPs, m.LockOPs())
			assert.True(cs.OPs.Equal(m.OPs()))
		})
	}
}
