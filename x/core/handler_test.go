package core_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	transfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	"github.com/stretchr/testify/suite"

	"github.com/datachainlab/cross/x/core/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
	ibctesting "github.com/datachainlab/cross/x/ibc/testing"
)

type CrossTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain

	queryClient transfertypes.QueryClient
}

func (suite *CrossTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))

	queryHelper := baseapp.NewQueryServerTestHelper(suite.chainA.GetContext(), suite.chainA.App.InterfaceRegistry())
	transfertypes.RegisterQueryServer(queryHelper, suite.chainA.App.TransferKeeper)
	suite.queryClient = transfertypes.NewQueryClient(queryHelper)
}

func (suite *CrossTestSuite) TestHandleMsgInitiate() {
	// setup

	_, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, ibctesting.Tendermint)
	suite.chainB.CreatePortCapability(crosstypes.PortID)
	channelA, _ := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, crosstypes.PortID, crosstypes.PortID, channeltypes.UNORDERED)
	_ = channelA

	msg := crosstypes.NewMsgInitiate(
		suite.chainA.SenderAccount.GetAddress().Bytes(),
		suite.chainA.ChainID,
		0,
		types.CommitProtocolSimple,
		[]types.ContractTransaction{},
		clienttypes.NewHeight(0, uint64(suite.chainA.CurrentHeader.Height)+100),
		0,
	)

	err := suite.coordinator.SendMsg(suite.chainA, suite.chainB, clientB, msg)
	suite.Require().NoError(err)
}

func TestCrossTestSuite(t *testing.T) {
	suite.Run(t, new(CrossTestSuite))
}
