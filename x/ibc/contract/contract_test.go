package contract

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/bluele/crossccc/x/ibc/crossccc"
	lock "github.com/bluele/crossccc/x/ibc/store/lock"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkstore "github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/types"
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
	k := NewKeeper(stk)
	h := NewContractHandler(k, func(kvs sdk.KVStore) crossccc.State {
		return lock.NewStore(kvs)
	})
	c := makeContract()
	contractID := "contract0"
	h.AddRoute(contractID, c)

	cms := makeCMStore(t, stk)
	header := abci.Header{}
	ctx := sdk.NewContext(cms, header, false, log.NewNopLogger())
	ctx = crossccc.WithSigners(ctx, []sdk.AccAddress{[]byte("user0")})

	{
		contractInfo := NewContractInfo(contractID, "issue", [][]byte{
			[]byte("mycoin"),
			[]byte("100"),
		})
		bz, _ := EncodeContractSignature(contractInfo)
		state, err := h.Handle(ctx, bz)
		if err != nil {
			assert.FailNow(err.Error())
		}
		tx0 := tmhash.Sum([]byte("tx0"))
		// Do a Precommit phase
		assert.NoError(state.Precommit(tx0))
		cms.Commit()

		{
			contractInfo := NewContractInfo(contractID, "test-balance", [][]byte{
				[]byte("mycoin"),
				[]byte("100"),
			})
			bz, _ := EncodeContractSignature(contractInfo)
			_, err := h.Handle(ctx, bz)
			if err == nil {
				assert.FailNow("expected an error")
			}
		}

		// Do a Commit phase
		assert.NoError(state.Commit(tx0))
		cms.Commit()

		{
			contractInfo := NewContractInfo(contractID, "test-balance", [][]byte{
				[]byte("mycoin"),
				[]byte("100"),
			})
			bz, _ := EncodeContractSignature(contractInfo)
			_, err := h.Handle(ctx, bz)
			if err != nil {
				assert.FailNow(err.Error())
			}
		}

		{
			contractInfo := NewContractInfo(contractID, "issue", [][]byte{
				[]byte("mycoin2"),
				[]byte("50"),
			})
			bz, _ := EncodeContractSignature(contractInfo)
			state, err := h.Handle(ctx, bz)
			if err != nil {
				assert.FailNow(err.Error())
			}
			// Commit immediately
			assert.NoError(state.CommitImmediately())
			cms.Commit()
		}

		{
			contractInfo := NewContractInfo(contractID, "test-balance", [][]byte{
				[]byte("mycoin2"),
				[]byte("50"),
			})
			bz, _ := EncodeContractSignature(contractInfo)
			_, err := h.Handle(ctx, bz)
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

func getBalanceOf(store crossccc.Store, address sdk.AccAddress) sdk.Coins {
	bz := store.Get(address)
	if bz == nil {
		return sdk.NewCoins()
	}
	return unmarshalCoin(bz)
}

func setBalance(store crossccc.Store, address sdk.AccAddress, balance sdk.Coins) {
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
			F: func(ctx Context, store crossccc.Store) error {
				coin, err := parseCoin(ctx, 0, 1)
				if err != nil {
					return err
				}
				balance := getBalanceOf(store, ctx.Signers()[0])
				balance = balance.Add(coin)
				setBalance(store, ctx.Signers()[0], balance)
				return nil
			},
		},
		{
			Name: "transfer",
			F: func(ctx Context, store crossccc.Store) error {
				coin, err := parseCoin(ctx, 0, 1)
				if err != nil {
					return err
				}
				rem := sdk.NewCoins(coin)

				var recipient sdk.AccAddress = ctx.Args()[2]

				signerBalance := getBalanceOf(store, ctx.Signers()[0])
				if !signerBalance.IsAllGT(rem) {
					return fmt.Errorf("balance is insufficent")
				}
				signerBalance = signerBalance.Sub(rem)
				setBalance(store, ctx.Signers()[0], signerBalance)

				recipientBalance := getBalanceOf(store, recipient)
				recipientBalance.Add(rem...)
				setBalance(store, recipient, recipientBalance)

				return nil
			},
		},
		{
			Name: "test-balance",
			F: func(ctx Context, store crossccc.Store) error {
				coin, err := parseCoin(ctx, 0, 1)
				if err != nil {
					return err
				}
				balance := getBalanceOf(store, ctx.Signers()[0])
				if !balance.AmountOf(coin.Denom).Equal(coin.Amount) {
					return errors.New("amount is unexpected")
				}
				return nil
			},
		},
	})
	return c
}

func makeCMStore(t *testing.T, key sdk.StoreKey) types.CommitMultiStore {
	assert := assert.New(t)
	d := db.NewDB("test", db.MemDBBackend, "")

	cms := sdkstore.NewCommitMultiStore(d)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, d)
	assert.NoError(cms.LoadLatestVersion())
	return cms
}
