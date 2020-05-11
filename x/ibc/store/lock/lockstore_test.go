package lock

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestLockStore(t *testing.T) {
	assert := assert.New(t)
	stk := sdk.NewKVStoreKey("main")

	cms := makeCMStore(t, stk)
	st := newLockStore(cms.GetKVStore(stk))
	k0 := []byte("k0")

	assert.False(st.IsLocked(k0))

	st.Lock(k0)
	assert.True(st.IsLocked(k0))
	assert.Panics(func() {
		st.Lock(k0)
	})
	st.Unlock(k0)
	assert.False(st.IsLocked(k0))

	st.Lock(k0)
	assert.True(st.IsLocked(k0))
	st.Unlock(k0)
}
