package lock

import (
	"testing"

	sdkstore "github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/ibc/cross"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto/tmhash"
	db "github.com/tendermint/tm-db"
)

func TestStore(t *testing.T) {
	assert := assert.New(t)

	stk := sdk.NewKVStoreKey("main")
	txID0 := tmhash.Sum([]byte("tx0"))
	txID1 := tmhash.Sum([]byte("tx1"))
	txID2 := tmhash.Sum([]byte("tx2"))

	{
		cms := makeCMStore(t, stk)

		k0, v0 := []byte("k0"), []byte("v0")
		{
			st := NewStore(cms.GetKVStore(stk), cross.ExactStateCondition)
			st.Set(k0, v0)
			assert.Equal(st.OPs(), cross.OPs{
				WriteOP{k0, v0},
			})
			v := st.Get(k0)
			assert.Equal(v0, v)
			assert.NoError(st.Precommit(txID0))
			assert.True(st.OPs()[0].Equal(WriteOP{k0, v0}))
			cms.Commit()

			{ // In other tx, try to access locked entry, but it will be failed
				st = NewStore(cms.GetKVStore(stk), cross.ExactStateCondition)
				assert.Panics(func() {
					st.Get(k0)
				})
				assert.Panics(func() {
					st.Set(k0, v0)
				})
			}

			assert.NoError(st.Commit(txID0))
			cms.Commit()
		}

		{
			st := NewStore(cms.GetKVStore(stk), cross.ExactStateCondition)
			assert.Equal(v0, st.Get(k0))
		}

		v1 := []byte("v1")
		{
			st := NewStore(cms.GetKVStore(stk), cross.ExactStateCondition)
			st.Set(k0, v1)
			assert.NoError(st.Precommit(txID1))
			cms.Commit()
			assert.NoError(st.Commit(txID1))
			cms.Commit()

			st = NewStore(cms.GetKVStore(stk), cross.ExactStateCondition)
			assert.Equal(v1, st.Get(k0))
		}

		{
			st := NewStore(cms.GetKVStore(stk), cross.ExactStateCondition)
			st.Delete(k0)
			assert.NoError(st.Precommit(txID2))
			cms.Commit()
			assert.NoError(st.Commit(txID2))
			cms.Commit()
			st = NewStore(cms.GetKVStore(stk), cross.ExactStateCondition)
			assert.True(st.Get(k0) == nil)
		}

		k1 := []byte("k1")
		{
			st := NewStore(cms.GetKVStore(stk), cross.ExactStateCondition)
			st.Get(k0)
			st.Set(k1, v1)
			assert.NoError(st.Precommit(txID2))
			cms.Commit()

			{
				st = NewStore(cms.GetKVStore(stk), cross.ExactStateCondition)
				assert.NotPanics(func() {
					st.Get(k0)
				})
				assert.Panics(func() {
					st.Get(k1)
				})
				assert.Panics(func() {
					st.Set(k1, []byte("update"))
				})
			}

			assert.NoError(st.Commit(txID2))
			cms.Commit()

			assert.True(st.Get(k0) == nil)
			assert.Equal(st.Get(k1), v1)
		}
	}
}

func makeCMStore(t *testing.T, key sdk.StoreKey) types.CommitMultiStore {
	assert := assert.New(t)
	d := db.NewDB("test", db.MemDBBackend, "")

	cms := sdkstore.NewCommitMultiStore(d)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, d)
	assert.NoError(cms.LoadLatestVersion())
	return cms
}
