package keeper

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/datachainlab/cross/x/ibc/contract/types"
	"github.com/datachainlab/cross/x/ibc/cross"
	lock "github.com/datachainlab/cross/x/ibc/store/lock"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkstore "github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	"github.com/tendermint/tendermint/libs/log"
	db "github.com/tendermint/tm-db"
)

var testcdc *codec.Codec

func init() {
	testcdc = codec.New()

	testcdc.RegisterConcrete(sdk.Coin{}, "sdk/Coin", nil)
	testcdc.RegisterConcrete(sdk.Coins{}, "sdk/Coins", nil)
}

func TestContractHandler(t *testing.T) {
	assert := assert.New(t)

	stk := sdk.NewKVStoreKey("main")
	k := NewKeeper(testcdc, stk)
	h := NewContractHandler(k, func(kvs sdk.KVStore, tp cross.StateConstraintType) cross.State {
		return lock.NewStore(kvs, cross.ExactMatchStateConstraint)
	})
	c := makeContract()
	contractID := "contract0"
	h.AddRoute(contractID, c)

	cms := makeCMStore(t, stk)
	header := abci.Header{}
	ctx := sdk.NewContext(cms, header, false, log.NewNopLogger())
	ctx = cross.WithSigners(ctx, []sdk.AccAddress{[]byte("user0")})

	{
		contractInfo := types.NewContractCallInfo(contractID, "issue", [][]byte{
			[]byte("mycoin"),
			[]byte("100"),
		})
		bz, _ := types.EncodeContractSignature(contractInfo)
		state, _, err := h.Handle(ctx, cross.NoStateConstraint, bz)
		if err != nil {
			assert.FailNow(err.Error())
		}
		tx0 := tmhash.Sum([]byte("tx0"))
		// Do a Precommit phase
		assert.NoError(state.Precommit(tx0))
		cms.Commit()

		{
			contractInfo := types.NewContractCallInfo(contractID, "test-balance", [][]byte{
				[]byte("mycoin"),
				[]byte("100"),
			})
			bz, _ := types.EncodeContractSignature(contractInfo)
			_, _, err := h.Handle(ctx, cross.NoStateConstraint, bz)
			if err == nil {
				assert.FailNow("expected an error")
			}
		}

		// Do a Commit phase
		assert.NoError(state.Commit(tx0))
		cms.Commit()

		{
			contractInfo := types.NewContractCallInfo(contractID, "test-balance", [][]byte{
				[]byte("mycoin"),
				[]byte("100"),
			})
			bz, _ := types.EncodeContractSignature(contractInfo)
			_, _, err := h.Handle(ctx, cross.NoStateConstraint, bz)
			if err != nil {
				assert.FailNow(err.Error())
			}
		}

		{
			contractInfo := types.NewContractCallInfo(contractID, "issue", [][]byte{
				[]byte("mycoin2"),
				[]byte("50"),
			})
			bz, _ := types.EncodeContractSignature(contractInfo)
			state, _, err := h.Handle(ctx, cross.NoStateConstraint, bz)
			if err != nil {
				assert.FailNow(err.Error())
			}
			// Commit immediately
			assert.NoError(state.CommitImmediately())
			cms.Commit()
		}

		{
			contractInfo := types.NewContractCallInfo(contractID, "test-balance", [][]byte{
				[]byte("mycoin2"),
				[]byte("50"),
			})
			bz, _ := types.EncodeContractSignature(contractInfo)
			_, _, err := h.Handle(ctx, cross.NoStateConstraint, bz)
			if err != nil {
				assert.FailNow(err.Error())
			}
		}
	}

}

func marshalCoin(coins sdk.Coins) []byte {
	return testcdc.MustMarshalBinaryLengthPrefixed(coins)
}

func unmarshalCoin(bz []byte) sdk.Coins {
	var coins sdk.Coins
	testcdc.MustUnmarshalBinaryLengthPrefixed(bz, &coins)
	return coins
}

func getBalanceOf(store cross.Store, address sdk.AccAddress) sdk.Coins {
	bz := store.Get(address)
	if bz == nil {
		return sdk.NewCoins()
	}
	return unmarshalCoin(bz)
}

func setBalance(store cross.Store, address sdk.AccAddress, balance sdk.Coins) {
	bz := marshalCoin(balance)
	store.Set(address, bz)
}

func makeContract() Contract {
	var parseCoin = func(ctx Context, denomIdx, amountIdx int) (sdk.Coin, error) {
		denom := string(ctx.Args()[denomIdx])
		amount, err := strconv.Atoi(string(ctx.Args()[amountIdx]))
		if err != nil {
			return sdk.Coin{}, err
		}
		if amount < 0 {
			return sdk.Coin{}, fmt.Errorf("amount must be positive number")
		}
		coin := sdk.NewInt64Coin(denom, int64(amount))
		return coin, nil
	}

	c := NewContract([]Method{
		{
			Name: "issue",
			F: func(ctx Context, store cross.Store) ([]byte, error) {
				coin, err := parseCoin(ctx, 0, 1)
				if err != nil {
					return nil, err
				}
				balance := getBalanceOf(store, ctx.Signers()[0])
				balance = balance.Add(coin)
				setBalance(store, ctx.Signers()[0], balance)
				return nil, nil
			},
		},
		{
			Name: "transfer",
			F: func(ctx Context, store cross.Store) ([]byte, error) {
				coin, err := parseCoin(ctx, 0, 1)
				if err != nil {
					return nil, err
				}
				rem := sdk.NewCoins(coin)

				var recipient sdk.AccAddress = ctx.Args()[2]

				signerBalance := getBalanceOf(store, ctx.Signers()[0])
				if !signerBalance.IsAllGT(rem) {
					return nil, fmt.Errorf("balance is insufficent")
				}
				signerBalance = signerBalance.Sub(rem)
				setBalance(store, ctx.Signers()[0], signerBalance)

				recipientBalance := getBalanceOf(store, recipient)
				recipientBalance.Add(rem...)
				setBalance(store, recipient, recipientBalance)

				return nil, nil
			},
		},
		{
			Name: "test-balance",
			F: func(ctx Context, store cross.Store) ([]byte, error) {
				coin, err := parseCoin(ctx, 0, 1)
				if err != nil {
					return nil, err
				}
				balance := getBalanceOf(store, ctx.Signers()[0])
				if !balance.AmountOf(coin.Denom).Equal(coin.Amount) {
					return nil, errors.New("amount is unexpected")
				}
				return nil, nil
			},
		},
	})
	return c
}

func makeCMStore(t *testing.T, key sdk.StoreKey) sdk.CommitMultiStore {
	assert := assert.New(t)
	d := db.NewDB("test", db.MemDBBackend, "")

	cms := sdkstore.NewCommitMultiStore(d)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, d)
	assert.NoError(cms.LoadLatestVersion())
	return cms
}
