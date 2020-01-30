package keeper_test

import (
	"testing"

	"github.com/bluele/crossccc/simapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

type KeeperTestSuite struct {
	suite.Suite
}

func (suite *KeeperTestSuite) SetupTest() {}

type appContext struct {
	cdc    *codec.Codec
	ctx    sdk.Context
	app    *simapp.SimApp
	valSet *tmtypes.ValidatorSet
}

func (suite *KeeperTestSuite) createApp() *appContext {
	isCheckTx := false
	app := simapp.Setup(isCheckTx)
	ctx := app.BaseApp.NewContext(isCheckTx, abci.Header{})
	privVal := tmtypes.NewMockPV()
	validator := tmtypes.NewValidator(privVal.GetPubKey(), 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})

	actx := &appContext{
		cdc:    app.Codec(),
		ctx:    ctx,
		app:    app,
		valSet: valSet,
	}
	suite.createClient(actx)
	suite.createConnection(actx, connectionexported.OPEN)
	return actx
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
