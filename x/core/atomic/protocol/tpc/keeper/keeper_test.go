package keeper_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	transfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	"github.com/stretchr/testify/suite"

	samplemodtypes "github.com/datachainlab/cross/simapp/samplemod/types"
	accounttypes "github.com/datachainlab/cross/x/core/account/types"
	initiatortypes "github.com/datachainlab/cross/x/core/initiator/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
	ibctesting "github.com/datachainlab/cross/x/ibc/testing"
	"github.com/datachainlab/cross/x/packets"
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

func (suite *KeeperTestSuite) TestTransaction() {
	// setup:
	// A(coordinator) => B(participant) -> Connection: AB, BA, Channel: AB, AB
	// A(coordinator) => C(participant) -> Connection: AC, CA, Channel: AC, AC

	_, _, connAB, connBA := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, ibctesting.Tendermint)
	suite.chainB.CreatePortCapability(crosstypes.PortID)
	channelAB, channelBA := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connAB, connBA, crosstypes.PortID, crosstypes.PortID, channeltypes.UNORDERED)
	chAB := xcctypes.ChannelInfo{Port: channelAB.PortID, Channel: channelAB.ID}
	xccB, err := xcctypes.PackCrossChainChannel(&chAB)
	suite.Require().NoError(err)

	_, _, connAC, connCA := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainC, ibctesting.Tendermint)
	suite.chainC.CreatePortCapability(crosstypes.PortID)
	channelAC, channelCA := suite.coordinator.CreateChannel(suite.chainA, suite.chainC, connAC, connCA, crosstypes.PortID, crosstypes.PortID, channeltypes.UNORDERED)
	chAC := xcctypes.ChannelInfo{Port: channelAC.PortID, Channel: channelAC.ID}
	xccC, err := xcctypes.PackCrossChainChannel(&chAC)
	suite.Require().NoError(err)

	xccSelf, err := xcctypes.PackCrossChainChannel(
		suite.chainA.App.XCCResolver.GetSelfCrossChainChannel(suite.chainA.GetContext()),
	)
	suite.Require().NoError(err)

	_, _, _, _, _ = channelBA, channelCA, xccB, xccC, xccSelf

	var cases = []struct {
		name string
		txs  [2]initiatortypes.ContractTransaction
	}{
		{
			"case0",
			[2]initiatortypes.ContractTransaction{
				{
					CrossChainChannel: xccB,
					Signers: []accounttypes.AccountID{
						accounttypes.AccountID(suite.chainB.SenderAccount.GetAddress()),
					},
					CallInfo: samplemodtypes.NewContractCallRequest("counter").ContractCallInfo(suite.chainB.App.AppCodec()),
				},
				{
					CrossChainChannel: xccC,
					Signers: []accounttypes.AccountID{
						accounttypes.AccountID(suite.chainC.SenderAccount.GetAddress()),
					},
					CallInfo: samplemodtypes.NewContractCallRequest("counter").ContractCallInfo(suite.chainC.App.AppCodec()),
				},
			},
		},
	}

	for i, c := range cases {
		suite.Run(c.name, func() {
			txs, err := suite.chainA.App.CrossKeeper.InitiatorKeeper().ResolveTransactions(
				suite.chainA.GetContext(),
				c.txs[:],
			)
			suite.Require().NoError(err)

			txID := []byte(fmt.Sprintf("txid-%v", i))
			kA := suite.chainA.App.AtomicKeeper.TPCKeeper()

			ps := ibctesting.NewCapturePacketSender(
				packets.NewBasicPacketSender(suite.chainA.App.IBCKeeper.ChannelKeeper),
			)
			err = kA.SendPrepare(
				suite.chainA.GetContext(),
				ps,
				txID,
				txs,
				clienttypes.NewHeight(0, uint64(suite.chainA.CurrentHeader.Height)+100),
				0,
			)
			suite.Require().NoError(err)
		})
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
