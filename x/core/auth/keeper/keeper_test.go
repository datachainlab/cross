package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	transfertypes "github.com/cosmos/ibc-go/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/modules/core/exported"
	"github.com/stretchr/testify/suite"

	authtypes "github.com/datachainlab/cross/x/core/auth/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
	ibctesting "github.com/datachainlab/cross/x/ibc/testing"
)

type KeeperTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
	chainC *ibctesting.TestChain

	queryClient transfertypes.QueryClient
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 3)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))
	suite.chainC = suite.coordinator.GetChain(ibctesting.GetChainID(2))

	queryHelper := baseapp.NewQueryServerTestHelper(suite.chainA.GetContext(), suite.chainA.App.InterfaceRegistry())
	transfertypes.RegisterQueryServer(queryHelper, suite.chainA.App.TransferKeeper)
	suite.queryClient = transfertypes.NewQueryClient(queryHelper)
}

func (suite *KeeperTestSuite) TestAuth() {
	// setup channels
	_, _, connAB, connBA := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint, ibctesting.CrossVersion)
	channelAB, _ := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connAB, connBA, crosstypes.PortID, crosstypes.PortID, channeltypes.UNORDERED)
	chAB := xcctypes.ChannelInfo{Port: channelAB.PortID, Channel: channelAB.ID}

	_, _, connAC, connCA := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainC, exported.Tendermint, ibctesting.CrossVersion)
	channelAC, _ := suite.coordinator.CreateChannel(suite.chainA, suite.chainC, connAC, connCA, crosstypes.PortID, crosstypes.PortID, channeltypes.UNORDERED)
	chAC := xcctypes.ChannelInfo{Port: channelAC.PortID, Channel: channelAC.ID}

	akA := suite.chainA.App.CrossKeeper.AuthKeeper()

	accB := authtypes.NewAccount(&chAB, suite.chainB.SenderAccount.GetAddress().Bytes(), authtypes.NewAuthTypeChannel())
	accC := authtypes.NewAccount(&chAC, suite.chainC.SenderAccount.GetAddress().Bytes(), authtypes.NewAuthTypeChannel())

	var txID = []byte("tx0")
	var signers = []authtypes.Account{accB, accC}

	suite.Require().NoError(
		akA.InitAuthState(
			suite.chainA.GetContext(),
			txID,
			signers,
		),
	)

	{
		var accounts = []authtypes.AccountID{accB.Id}
		completed, err := akA.ReceiveIBCSignTx(
			suite.chainA.GetContext(),
			channelAB.PortID, chAB.Channel,
			authtypes.NewPacketDataIBCSignTx(txID, accounts, clienttypes.NewHeight(0, 100), 0),
		)
		suite.Require().NoError(err)
		suite.Require().False(completed)
	}
	{
		var accounts = []authtypes.AccountID{accC.Id}
		completed, err := akA.ReceiveIBCSignTx(
			suite.chainA.GetContext(),
			channelAC.PortID, chAC.Channel,
			authtypes.NewPacketDataIBCSignTx(txID, accounts, clienttypes.NewHeight(0, 100), 0),
		)
		suite.Require().NoError(err)
		suite.Require().True(completed)
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
