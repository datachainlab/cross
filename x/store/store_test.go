package store

import (
	"testing"

	sdkstore "github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	db "github.com/tendermint/tm-db"
)

func TestKVStore(t *testing.T) {
	require := require.New(t)

	stk := sdk.NewKVStoreKey("state")
	s := newKVStore(stk)

	cms := makeCMStore(t, stk)
	ctx := sdk.NewContext(cms, tmproto.Header{}, false, tmlog.NewNopLogger())

	key0, value0 := []byte("key0"), []byte("value0")

	s.Set(ctx, key0, value0)
	require.Equal(value0, s.Get(ctx, key0))

	s1 := s.Prefix([]byte("/1/"))
	require.Nil(s1.Get(ctx, key0))

	key1, value1 := []byte("key1"), []byte("value1")

	s.Set(ctx, key1, value1)
	require.Equal(value1, s.Get(ctx, key1))
}

func makeCMStore(t *testing.T, key sdk.StoreKey) sdk.CommitMultiStore {
	require := require.New(t)
	d, err := db.NewDB("test", db.MemDBBackend, "")
	if err != nil {
		panic(err)
	}
	cms := sdkstore.NewCommitMultiStore(d)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, d)
	require.NoError(cms.LoadLatestVersion())
	return cms
}
