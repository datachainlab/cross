package store

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestLockStore(t *testing.T) {
	require := require.New(t)
	stk := sdk.NewKVStoreKey("main")

	cms := makeCMStore(t, stk)
	ctx := sdk.NewContext(cms, tmproto.Header{}, false, tmlog.NewNopLogger())
	st := newLockStore(newKVStore(stk))
	k0 := []byte("k0")

	require.False(st.IsLocked(ctx, k0))

	st.Lock(ctx, k0)
	require.True(st.IsLocked(ctx, k0))
	require.Panics(func() {
		st.Lock(ctx, k0)
	})
	st.Unlock(ctx, k0)
	require.False(st.IsLocked(ctx, k0))

	st.Lock(ctx, k0)
	require.True(st.IsLocked(ctx, k0))
	st.Unlock(ctx, k0)
}
