package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
	ibctesting "github.com/datachainlab/cross/x/ibc/testing"
	"github.com/datachainlab/cross/x/packets"
)

type KeeperTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain

	queryClient types.QueryClient
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))

	queryHelper := baseapp.NewQueryServerTestHelper(suite.chainA.GetContext(), suite.chainA.App.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.chainA.App.TransferKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

func (suite *KeeperTestSuite) TestCall() {
	_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, ibctesting.Tendermint)
	suite.chainB.CreatePortCapability(crosstypes.PortID)
	channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, crosstypes.PortID, crosstypes.PortID, channeltypes.UNORDERED)
	_ = channelB
	chAB := crosstypes.ChannelInfo{Port: channelA.PortID, Channel: channelA.ID}
	selfCh := crosstypes.ChannelInfo{}
	cidB, err := crosstypes.PackChainID(&chAB)
	suite.Require().NoError(err)
	selfCid, err := crosstypes.PackChainID(&selfCh)

	k := suite.chainA.App.CrossKeeper.SimpleKeeper()
	txs := []crosstypes.ContractTransaction{
		{
			ChainId: *selfCid,
			Signers: []crosstypes.AccountAddress{suite.chainA.SenderAccount.GetAddress().Bytes()},
		},
		{
			ChainId: *cidB,
			Signers: []crosstypes.AccountAddress{suite.chainB.SenderAccount.GetAddress().Bytes()},
		},
	}
	suite.Require().NoError(
		k.SendCall(
			suite.chainA.GetContext(),
			packets.NewBasicPacketSender(suite.chainA.App.IBCKeeper.ChannelKeeper),
			[]byte("txid0"),
			txs,
		),
	)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
