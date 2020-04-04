package contract_test

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/datachainlab/cross/example/simapp"
	"github.com/datachainlab/cross/x/ibc/contract"
	"github.com/datachainlab/cross/x/ibc/cross"
	lock "github.com/datachainlab/cross/x/ibc/store/lock"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

type HandlerTestSuite struct {
	suite.Suite

	cdc    *codec.Codec
	ctx    sdk.Context
	app    *simapp.SimApp
	valSet *tmtypes.ValidatorSet
}

func (suite *HandlerTestSuite) SetupTest() {
	isCheckTx := false
	app := simapp.Setup(isCheckTx)

	suite.cdc = app.Codec()
	suite.ctx = app.BaseApp.NewContext(isCheckTx, abci.Header{})
	suite.app = app

	privVal := tmtypes.NewMockPV()

	validator := tmtypes.NewValidator(privVal.GetPubKey(), 1)
	suite.valSet = tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})
}

func (suite *HandlerTestSuite) TestHandleContractCall() {
	contractHandler := contract.NewContractHandler(suite.app.ContractKeeper, func(kvs sdk.KVStore) cross.State {
		return lock.NewStore(kvs)
	})
	var methods []contract.Method
	methods = append(methods, []contract.Method{
		{
			Name: "f0",
			F: func(ctx contract.Context, store cross.Store) ([]byte, error) {
				return nil, nil
			},
		},
		{
			Name: "issue",
			F: func(ctx contract.Context, store cross.Store) ([]byte, error) {
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
			Name: "test-balance",
			F: func(ctx contract.Context, store cross.Store) ([]byte, error) {
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
	}...,
	)
	contractHandler.AddRoute("first", contract.NewContract(methods))
	handler := contract.NewHandler(suite.app.ContractKeeper, contractHandler)

	acc0 := sdk.AccAddress("acc0")
	acc1 := sdk.AccAddress("acc1")

	// Ensure that validation of handler is correct
	{
		{
			msg := contract.NewMsgContractCall(acc0, []sdk.AccAddress{acc0}, nil)
			_, err := handler(suite.ctx, msg)
			suite.Require().Error(err)
		}

		{
			c := contract.NewContractCallInfo("first", "f0", nil)
			bz, err := contract.EncodeContractSignature(c)
			suite.Require().NoError(err)
			msg := contract.NewMsgContractCall(acc0, []sdk.AccAddress{acc0}, bz)
			res, err := handler(suite.ctx, msg)
			suite.Require().NoError(err)
			suite.Require().NotNil(res, "%+v", res) // successfully executed
		}

		{
			c := contract.NewContractCallInfo("dummy", "f0", nil)
			bz, err := contract.EncodeContractSignature(c)
			suite.Require().NoError(err)
			msg := contract.NewMsgContractCall(acc0, []sdk.AccAddress{acc0}, bz)
			_, err = handler(suite.ctx, msg)
			suite.Require().Error(err)
		}

		{
			c := contract.NewContractCallInfo("first", "dummy", nil)
			bz, err := contract.EncodeContractSignature(c)
			suite.Require().NoError(err)
			msg := contract.NewMsgContractCall(acc0, []sdk.AccAddress{acc0}, bz)
			_, err = handler(suite.ctx, msg)
			suite.Require().Error(err)
		}
	}

	var commitApp = func() {
		suite.app.Commit()

		suite.app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: suite.app.LastBlockHeight() + 1}})
		suite.ctx = suite.app.BaseApp.NewContext(false, abci.Header{})
	}

	// Ensure that commit store state successfully
	{
		{
			c := contract.NewContractCallInfo("first", "issue", [][]byte{[]byte("tone"), []byte("80")})
			bz, err := contract.EncodeContractSignature(c)
			suite.Require().NoError(err)
			msg := contract.NewMsgContractCall(acc0, []sdk.AccAddress{acc0}, bz)
			ctx, write := suite.ctx.CacheContext()
			res, err := handler(ctx, msg)
			suite.Require().NoError(err)
			suite.Require().NotNil(res, "%+v", res) // successfully executed
			write()
			commitApp()
		}

		{
			c := contract.NewContractCallInfo("first", "test-balance", [][]byte{[]byte("tone"), []byte("80")})
			bz, err := contract.EncodeContractSignature(c)
			suite.Require().NoError(err)
			msg := contract.NewMsgContractCall(acc0, []sdk.AccAddress{acc0}, bz)
			ctx, _ := suite.ctx.CacheContext()
			res, err := handler(ctx, msg)
			suite.Require().NoError(err)
			suite.Require().NotNil(res, "%+v", res) // successfully executed
		}

		{
			c := contract.NewContractCallInfo("first", "test-balance", [][]byte{[]byte("tone"), []byte("80")})
			bz, err := contract.EncodeContractSignature(c)
			suite.Require().NoError(err)
			msg := contract.NewMsgContractCall(acc1, []sdk.AccAddress{acc0}, bz)
			ctx, _ := suite.ctx.CacheContext()
			_, err = handler(ctx, msg)
			suite.Require().Error(err)
		}
	}
}

func parseCoin(ctx contract.Context, denomIdx, amountIdx int) (sdk.Coin, error) {
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

var testcdc *codec.Codec

func init() {
	testcdc = codec.New()

	testcdc.RegisterConcrete(sdk.Coin{}, "sdk/Coin", nil)
	testcdc.RegisterConcrete(sdk.Coins{}, "sdk/Coins", nil)
}
