package contract_test

import (
	"testing"

	"github.com/bluele/crossccc/example/simapp"
	"github.com/bluele/crossccc/x/ibc/contract"
	"github.com/bluele/crossccc/x/ibc/crossccc"
	lock "github.com/bluele/crossccc/x/ibc/store/lock"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type HandlerTestSuite struct {
	suite.Suite

	cdc    *codec.Codec
	ctx    sdk.Context
	app    *simapp.SimApp
	valSet *tmtypes.ValidatorSet
}

func (suite *HandlerTestSuite) SetupTest() {
	lock.RegisterCodec(crossccc.ModuleCdc)

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
	cstk := sdk.NewKVStoreKey("contract")
	k := contract.NewKeeper(cstk)
	contractHandler := contract.NewContractHandler(k, func(kvs sdk.KVStore) crossccc.State {
		return lock.NewStore(kvs)
	})
	handler := contract.NewHandler(contractHandler)
	msg := contract.NewMsgContractCall()
	_, err := handler(suite.ctx, msg)
	suite.Require().Error(err)
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
