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
	k1 := []byte("k1")

	_, locked := st.HasAnyLocked(k0)
	assert.False(locked)

	st.Lock(LOCK_TYPE_READ, k0)
	_, locked = st.HasAnyLocked(k0)
	assert.True(locked)
	st.Lock(LOCK_TYPE_READ, k0)
	assert.Panics(func() {
		st.Lock(LOCK_TYPE_WRITE, k0)
	})

	st.Lock(LOCK_TYPE_READ, k1)
	st.Unlock(LOCK_TYPE_READ, k1)

	assert.Panics(func() {
		st.Unlock(LOCK_TYPE_WRITE, k0)
	})

	st.Unlock(LOCK_TYPE_READ, k0)
	_, locked = st.HasAnyLocked(k0)
	assert.True(locked)
	st.Unlock(LOCK_TYPE_READ, k0)
	_, locked = st.HasAnyLocked(k0)
	assert.False(locked)

	st.Lock(LOCK_TYPE_WRITE, k0)
	_, locked = st.HasAnyLocked(k0)
	assert.True(locked)

	assert.Panics(func() {
		st.Lock(LOCK_TYPE_READ, k0)
	})
	assert.Panics(func() {
		st.Lock(LOCK_TYPE_WRITE, k0)
	})

	st.Unlock(LOCK_TYPE_WRITE, k0)
	_, locked = st.HasAnyLocked(k0)
	assert.False(locked)
}

func TestTxLock(t *testing.T) {
	assert := assert.New(t)

	{
		lock := txLock(nil)
		assert.Equal(LOCK_TYPE_NONE, lock.Type())

		assert.NoError(lock.Append(LOCK_TYPE_READ))
		assert.Equal(LOCK_TYPE_READ, lock.Type())

		assert.NoError(lock.Append(LOCK_TYPE_READ))
		assert.Equal(LOCK_TYPE_READ, lock.Type())

		assert.Error(lock.Remove(LOCK_TYPE_WRITE))
		assert.NoError(lock.Remove(LOCK_TYPE_READ))
		assert.Equal(LOCK_TYPE_READ, lock.Type())

		assert.NoError(lock.Remove(LOCK_TYPE_READ))
		assert.Equal(LOCK_TYPE_NONE, lock.Type())

		assert.Error(lock.Remove(LOCK_TYPE_READ))
		assert.Error(lock.Remove(LOCK_TYPE_WRITE))

		assert.NoError(lock.Append(LOCK_TYPE_READ))
		assert.Equal(LOCK_TYPE_READ, lock.Type())

		assert.Error(lock.Append(LOCK_TYPE_WRITE))
	}

	{
		lock := txLock(nil)
		assert.NoError(lock.Append(LOCK_TYPE_WRITE))
		assert.Equal(LOCK_TYPE_WRITE, lock.Type())

		assert.Error(lock.Append(LOCK_TYPE_WRITE))
		assert.Error(lock.Append(LOCK_TYPE_READ))

		assert.Error(lock.Remove(LOCK_TYPE_READ))
		assert.NoError(lock.Remove(LOCK_TYPE_WRITE))
		assert.Equal(LOCK_TYPE_NONE, lock.Type())

		assert.NoError(lock.Append(LOCK_TYPE_READ))
		assert.Equal(LOCK_TYPE_READ, lock.Type())
	}
}
