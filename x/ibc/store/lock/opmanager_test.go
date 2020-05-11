package lock

import (
	"fmt"
	"testing"

	"github.com/datachainlab/cross/x/ibc/cross"
	"github.com/stretchr/testify/assert"
)

func TestOPManager(t *testing.T) {
	var R = func(k, v string) ReadOP {
		return ReadOP{[]byte(k), []byte(v)}
	}
	var W = func(k, v string) WriteOP {
		return WriteOP{[]byte(k), []byte(v)}
	}

	var cases = []struct {
		constraintType uint8
		inputs         []OP
		LockOPs        []LockOP
		OPs            cross.OPs
	}{
		// Test for ExactMatchStateConstraint
		{
			cross.ExactMatchStateConstraint,
			[]OP{},
			[]LockOP{},
			cross.OPs{},
		},
		{
			cross.ExactMatchStateConstraint,
			[]OP{R("k1", "v1")},
			[]LockOP{},
			cross.OPs{R("k1", "v1")},
		},
		{
			cross.ExactMatchStateConstraint,
			[]OP{R("k1", "v1"), W("k1", "v1-1")},
			[]LockOP{W("k1", "v1-1")},
			cross.OPs{R("k1", "v1"), W("k1", "v1-1")},
		},
		{
			cross.ExactMatchStateConstraint,
			[]OP{R("k1", "v1"), W("k2", "v2")},
			[]LockOP{W("k2", "v2")},
			cross.OPs{R("k1", "v1"), W("k2", "v2")},
		},
		{
			cross.ExactMatchStateConstraint,
			[]OP{R("k1", "v1"), W("k2", "v2"), W("k1", "v1-1")},
			[]LockOP{W("k2", "v2"), W("k1", "v1-1")},
			cross.OPs{R("k1", "v1"), W("k2", "v2"), W("k1", "v1-1")},
		},
		// Test for PreStateConstraint
		{
			cross.PreStateConstraint,
			[]OP{},
			[]LockOP{},
			cross.OPs{},
		},
		{
			cross.PreStateConstraint,
			[]OP{R("k1", "v1")},
			[]LockOP{},
			cross.OPs{R("k1", "v1")},
		},
		{
			cross.PreStateConstraint,
			[]OP{R("k1", "v1"), W("k1", "v1-1")},
			[]LockOP{W("k1", "v1-1")},
			cross.OPs{R("k1", "v1")},
		},
		{
			cross.PreStateConstraint,
			[]OP{R("k1", "v1"), W("k2", "v2")},
			[]LockOP{W("k2", "v2")},
			cross.OPs{R("k1", "v1")},
		},
		{
			cross.PreStateConstraint,
			[]OP{R("k1", "v1"), W("k2", "v2"), W("k1", "v1-1")},
			[]LockOP{W("k2", "v2"), W("k1", "v1-1")},
			cross.OPs{R("k1", "v1")},
		},
		// Test for PostStateConstraint
		{
			cross.PostStateConstraint,
			[]OP{},
			[]LockOP{},
			cross.OPs{},
		},
		{
			cross.PostStateConstraint,
			[]OP{R("k1", "v1")},
			[]LockOP{},
			cross.OPs{},
		},
		{
			cross.PostStateConstraint,
			[]OP{R("k1", "v1"), W("k1", "v1-1")},
			[]LockOP{W("k1", "v1-1")},
			cross.OPs{W("k1", "v1-1")},
		},
		{
			cross.PostStateConstraint,
			[]OP{R("k1", "v1"), W("k2", "v2")},
			[]LockOP{W("k2", "v2")},
			cross.OPs{W("k2", "v2")},
		},
		{
			cross.PostStateConstraint,
			[]OP{R("k1", "v1"), W("k2", "v2"), W("k1", "v1-1")},
			[]LockOP{W("k2", "v2"), W("k1", "v1-1")},
			cross.OPs{W("k2", "v2"), W("k1", "v1-1")},
		},
		// Test for OP
		{
			cross.ExactMatchStateConstraint,
			[]OP{ReadOP{nil, nil}},
			[]LockOP{},
			cross.OPs{ReadOP{nil, nil}},
		},
		{
			cross.ExactMatchStateConstraint,
			[]OP{WriteOP{nil, nil}},
			[]LockOP{WriteOP{nil, nil}},
			cross.OPs{WriteOP{nil, nil}},
		},
		{
			cross.ExactMatchStateConstraint,
			[]OP{WriteOP{nil, nil}, WriteOP{nil, nil}},
			[]LockOP{WriteOP{nil, nil}},
			cross.OPs{WriteOP{nil, nil}, WriteOP{nil, nil}},
		},
	}

	for i, cs := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			assert := assert.New(t)
			m, err := GetOPManager(cs.constraintType)
			if err != nil {
				assert.FailNow(err.Error())
			}
			for _, in := range cs.inputs {
				if tp := in.Type(); tp == OP_TYPE_READ {
					m.AddRead(in.(ReadOP).K, in.(ReadOP).V)
				} else if tp == OP_TYPE_WRITE {
					m.AddWrite(in.(WriteOP).K, in.(WriteOP).V)
				} else {
					assert.FailNow("fatal error")
				}
			}
			assert.Equal(cs.LockOPs, m.LockOPs())
			assert.True(cs.OPs.Equal(m.OPs()), m.OPs().String())
		})
	}
}
