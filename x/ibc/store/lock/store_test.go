package lock

import (
	"testing"

	"github.com/bluele/crossccc/x/ibc/crossccc"
	sdkstore "github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	db "github.com/tendermint/tm-db"
)

func makeCMStore(t *testing.T, key sdk.StoreKey) types.CommitMultiStore {
	assert := assert.New(t)
	d := db.NewDB("test", db.MemDBBackend, "")

	cms := sdkstore.NewCommitMultiStore(d)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, d)
	assert.NoError(cms.LoadLatestVersion())
	return cms
}

func TestStore(t *testing.T) {
	assert := assert.New(t)

	stk := sdk.NewKVStoreKey("main")

	{
		cms := makeCMStore(t, stk)
		st := NewStore(cms.GetKVStore(stk))
		k0, v0 := []byte("k0"), []byte("v0")
		st.Set(k0, v0)
		assert.Equal(st.OPs(), crossccc.OPs{
			Write{k0, v0},
		})
		v := st.Get(k0)
		assert.Equal(v0, v)
		txID := []byte("txID")
		assert.NoError(st.Precommit(txID))
		assert.True(st.OPs()[0].Equal(Write{k0, v0}))
		cms.Commit()

		st = NewStore(cms.GetKVStore(stk))
		assert.Panics(func() {
			st.Get(k0)
		})
		assert.NoError(st.Commit(txID))
		cms.Commit()

		st = NewStore(cms.GetKVStore(stk))
		v = st.Get(k0)
		assert.Equal(v0, v)
	}

}
