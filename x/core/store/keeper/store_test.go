package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdkstore "github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	contracttypes "github.com/datachainlab/cross/x/core/contract/types"
	"github.com/datachainlab/cross/x/core/store/types"
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

	s1.Set(ctx, key1, value1)
	require.Equal(value1, s1.Get(ctx, key1))
	require.Equal(value1, s.Get(ctx, []byte("/1/key1")))
	s1.Delete(ctx, key1)
	require.Nil(s1.Get(ctx, key1))
	require.Nil(s.Get(ctx, []byte("/1/key1")))
}

func TestStore(t *testing.T) {
	require := require.New(t)

	stk := sdk.NewKVStoreKey("state")
	var m codec.Marshaler
	s := NewStore(m, stk)

	cms := makeCMStore(t, stk)
	ctx := sdk.NewContext(cms, tmproto.Header{}, false, tmlog.NewNopLogger())

	key0, value0 := []byte("key0"), []byte("value0")

	s.Set(ctx, key0, value0)
	require.Equal(value0, s.Get(ctx, key0))

	s1 := s.Prefix([]byte("/1/"))
	require.Nil(s1.Get(ctx, key0))

	key1, value1 := []byte("key1"), []byte("value1")

	s1.Set(ctx, key1, value1)
	require.Equal(value1, s1.Get(ctx, key1))
	require.Equal(value1, s.Get(ctx, []byte("/1/key1")))
	s1.Delete(ctx, key1)
	require.Nil(s1.Get(ctx, key1))
	require.Nil(s.Get(ctx, []byte("/1/key1")))
}

func TestCommitStore(t *testing.T) {
	require := require.New(t)

	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	types.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	// var unpackOPItem = func(anyOPItem codectypes.Any) types.OP {
	// 	opItem, err := types.UnpackOPItem(cdc, anyOPItem)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	return opItem
	// }

	stk := sdk.NewKVStoreKey("main")
	k0, v0 := []byte("k0"), []byte("v0")
	k1, v1 := []byte("k1"), []byte("v1")
	cms := makeCMStore(t, stk)
	{
		lkmgr := types.NewLockManager()
		ctx := makeAtomicModeContext(cms, lkmgr)
		id0 := []byte("id0")
		st := NewStore(cdc, stk)
		st.Set(ctx, k0, v0)
		require.NoError(st.Precommit(ctx, id0))
		require.Equal(1, len(lkmgr.LockOPs().Ops))
		require.Equal(types.LockOP{K: k0, V: v0}, lkmgr.LockOPs().Ops[0])
		cms.Commit()

		// check if concurrent access is failed
		require.Panics(func() {
			ctx, _ := makeContext(cms).CacheContext()
			_ = st.Get(ctx, k0)
		})

		require.NoError(st.Commit(ctx, id0))
		cms.Commit()

		require.NotPanics(func() {
			ctx, _ := makeContext(cms).CacheContext()
			_ = st.Get(ctx, k0)
		})
	}
	{
		ctx, _ := makeContext(cms).CacheContext()
		st := NewStore(cdc, stk)
		require.Equal(v0, st.Get(ctx, k0))
	}
	{
		lkmgr := types.NewLockManager()
		ctx := makeAtomicModeContext(cms, lkmgr)
		id1 := []byte("id1")
		st := NewStore(cdc, stk)
		require.Equal(v0, st.Get(ctx, k0))
		st.Set(ctx, k1, v1)
		require.NoError(st.Precommit(ctx, id1))
		require.Equal(1, len(lkmgr.LockOPs().Ops))
		require.Equal(types.LockOP{K: k1, V: v1}, lkmgr.LockOPs().Ops[0])
		cms.Commit()

		// check if concurrent access is failed
		require.Panics(func() {
			ctx, _ := makeContext(cms).CacheContext()
			_ = st.Get(ctx, k1)
		})

		// check if concurrent read access is success
		require.NotPanics(func() {
			ctx, _ := makeContext(cms).CacheContext()
			_ = st.Get(ctx, k0)
		})
		require.NoError(st.Commit(ctx, id1))
		cms.Commit()

		require.NotPanics(func() {
			ctx, _ := makeContext(cms).CacheContext()
			_ = st.Get(ctx, k1)
			_ = st.Get(ctx, k0)
		})
	}
}

func makeContext(cms sdk.CommitMultiStore) sdk.Context {
	return sdk.NewContext(cms, tmproto.Header{}, false, tmlog.NewNopLogger())
}

func makeAtomicModeContext(cms sdk.CommitMultiStore, lkmgr types.LockManager) sdk.Context {
	ctx := sdk.NewContext(cms, tmproto.Header{}, false, tmlog.NewNopLogger())
	ctx = ctx.WithContext(types.ContextWithLockManager(ctx.Context(), lkmgr))
	return ctx.WithContext(
		contracttypes.ContextWithContractRuntimeInfo(
			ctx.Context(),
			contracttypes.ContractRuntimeInfo{CommitMode: contracttypes.AtomicMode},
		),
	)
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
