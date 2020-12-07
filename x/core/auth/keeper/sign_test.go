package keeper

import (
	"fmt"
	"testing"

	sdkstore "github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	accounttypes "github.com/datachainlab/cross/x/core/account/types"
	"github.com/datachainlab/cross/x/core/auth/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
	"github.com/datachainlab/cross/x/packets"

	"github.com/stretchr/testify/require"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	db "github.com/tendermint/tm-db"
)

func TestSign(t *testing.T) {
	var cases = []struct {
		name            string
		requiredSigners []accounttypes.Account
		signers         []accounttypes.Account
		isCompleted     bool
		wantsConsumed   []bool
	}{
		{
			"case-0",
			[]accounttypes.Account{
				{Id: []byte{0}},
			},
			[]accounttypes.Account{
				{Id: []byte{0}},
			},
			true,
			[]bool{true},
		},
		{
			"case-1",
			[]accounttypes.Account{
				{Id: []byte{0}},
				{Id: []byte{1}},
			},
			[]accounttypes.Account{
				{Id: []byte{0}},
				{Id: []byte{1}},
			},
			true,
			[]bool{true, true},
		},
		{
			"case-2",
			[]accounttypes.Account{
				{Id: []byte{0}},
				{Id: []byte{1}},
			},
			[]accounttypes.Account{
				{Id: []byte{0}},
				{Id: []byte{2}},
			},
			false,
			[]bool{true, false},
		},
		{
			"case-3",
			[]accounttypes.Account{
				{Id: []byte{0}},
				{Id: []byte{1}},
			},
			[]accounttypes.Account{
				{Id: []byte{0}},
				{Id: []byte{2}},
				{Id: []byte{1}},
			},
			true,
			[]bool{true, false, true},
		},
		{
			"case-4",
			[]accounttypes.Account{},
			[]accounttypes.Account{},
			true,
			[]bool{},
		},
	}

	storeKey := sdk.NewKVStoreKey("test")
	for i, cs := range cases {
		txID := []byte(fmt.Sprintf("tx-%v", i))
		t.Run(cs.name, func(t *testing.T) {
			require := require.New(t)

			cms := makeCMStore(t, storeKey)
			ctx := makeContext(cms)

			k := NewKeeper(types.ModuleCdc, storeKey, nil, packets.PacketSendKeeper{}, packets.NewNOPPacketMiddleware(), xcctypes.NewChannelInfoResolver(nil))
			require.NoError(k.InitAuthState(ctx, txID, cs.requiredSigners))

			for j, acc := range cs.signers {
				state, _ := k.getAuthState(ctx, txID)
				before := len(state.RemainingSigners)

				_, err := k.Sign(ctx, txID, []accounttypes.Account{acc})
				require.NoError(err)

				state, _ = k.getAuthState(ctx, txID)
				after := len(state.RemainingSigners)

				if cs.wantsConsumed[j] {
					require.True(before > after)
				} else {
					require.True(before == after)
				}
			}
			state, err := k.getAuthState(ctx, txID)
			require.NoError(err)
			require.Equal(cs.isCompleted, state.IsCompleted())
		})
	}
}

func makeContext(cms sdk.CommitMultiStore) sdk.Context {
	return sdk.NewContext(cms, tmproto.Header{}, false, tmlog.NewNopLogger())
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
