package keeper_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	transfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	"github.com/stretchr/testify/suite"

	samplemodtypes "github.com/datachainlab/cross/simapp/samplemod/types"
	accounttypes "github.com/datachainlab/cross/x/core/account/types"
	"github.com/datachainlab/cross/x/core/atomic/protocol/tpc/types"
	atomictypes "github.com/datachainlab/cross/x/core/atomic/types"
	initiatortypes "github.com/datachainlab/cross/x/core/initiator/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
	ibctesting "github.com/datachainlab/cross/x/ibc/testing"
	"github.com/datachainlab/cross/x/packets"
	"github.com/datachainlab/cross/x/utils"
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
			kB := suite.chainB.App.AtomicKeeper.TPCKeeper()
			kC := suite.chainC.App.AtomicKeeper.TPCKeeper()

			ps := ibctesting.NewCapturePacketSender(
				packets.NewBasicPacketSender(suite.chainA.App.IBCKeeper.ChannelKeeper),
			)
			suite.Require().NoError(
				kA.SendPrepare(
					suite.chainA.GetContext(),
					ps,
					txID,
					txs,
					clienttypes.NewHeight(0, uint64(suite.chainA.CurrentHeader.Height)+100),
					0,
				),
			)
			suite.chainA.NextBlock()

			// check if coordinator state is expected

			cs, found := suite.chainA.App.AtomicKeeper.TPCKeeper().GetCoordinatorState(suite.chainA.GetContext(), txID)
			suite.Require().True(found)
			suite.Require().Equal(atomictypes.COORDINATOR_PHASE_PREPARE, cs.Phase)
			suite.Require().Equal(atomictypes.COORDINATOR_DECISION_UNKNOWN, cs.Decision)

			// check if ReceiveCallPacket call is expected

			suite.Require().Equal(2, len(ps.Packets()))
			p0, p1 := ps.Packets()[0], ps.Packets()[1]
			prepareB := *suite.parsePacketToPacketDataPrepare(suite.chainB.App.AppCodec(), p0).(*types.PacketDataPrepare)
			prepareC := *suite.parsePacketToPacketDataPrepare(suite.chainC.App.AppCodec(), p1).(*types.PacketDataPrepare)

			_, prepareAckB, err := kB.ReceivePacketPrepare(
				suite.chainB.GetContext(), p0.GetDestPort(), p0.GetDestChannel(), prepareB,
			)
			suite.Require().NoError(err)
			suite.Require().Equal(atomictypes.PREPARE_RESULT_OK, prepareAckB.Result)
			suite.chainB.NextBlock()
			{
				ctxs, found := kB.GetContractTransactionState(suite.chainB.GetContext(), txID, 0)
				suite.Require().True(found)
				suite.Require().Equal(atomictypes.CONTRACT_TRANSACTION_STATUS_PREPARE, ctxs.Status)
				suite.Require().Equal(atomictypes.PREPARE_RESULT_OK, ctxs.PrepareResult)
			}

			_, prepareAckC, err := kC.ReceivePacketPrepare(
				suite.chainC.GetContext(), p1.GetDestPort(), p1.GetDestChannel(), prepareC,
			)
			suite.Require().NoError(err)
			suite.Require().Equal(atomictypes.PREPARE_RESULT_OK, prepareAckC.Result)
			suite.chainC.NextBlock()
			{
				ctxs, found := kC.GetContractTransactionState(suite.chainC.GetContext(), txID, 1)
				suite.Require().True(found)
				suite.Require().Equal(atomictypes.CONTRACT_TRANSACTION_STATUS_PREPARE, ctxs.Status)
				suite.Require().Equal(atomictypes.PREPARE_RESULT_OK, ctxs.PrepareResult)
			}

			// check if ReceiveCallAcknowledgement call is expected
			ps.Clear()
			_, err = kA.HandlePacketAcknowledgementPrepare(
				suite.chainA.GetContext(),
				p0.GetSourcePort(), p0.GetSourceChannel(),
				*prepareAckB, txID, 0, ps,
			)
			suite.Require().NoError(err)
			suite.Require().Equal(0, len(ps.Packets()))
			suite.chainA.NextBlock()

			ps.Clear()
			_, err = kA.HandlePacketAcknowledgementPrepare(
				suite.chainA.GetContext(),
				p1.GetSourcePort(), p1.GetSourceChannel(),
				*prepareAckC, txID, 1, ps,
			)
			suite.Require().NoError(err)
			suite.Require().Equal(2, len(ps.Packets()))
			suite.chainA.NextBlock()

			// check if each ReceivePacketCommit calls are expected

			p0, p1 = ps.Packets()[0], ps.Packets()[1]
			commitB := *suite.parsePacketToPacketDataPrepare(suite.chainB.App.AppCodec(), p0).(*types.PacketDataCommit)
			commitC := *suite.parsePacketToPacketDataPrepare(suite.chainC.App.AppCodec(), p1).(*types.PacketDataCommit)

			_, commitAckB, err := kB.ReceivePacketCommit(
				suite.chainB.GetContext(),
				p0.GetDestPort(), p0.GetDestChannel(),
				commitB,
			)
			suite.Require().NoError(err)
			suite.Require().Equal(types.COMMIT_STATUS_OK, commitAckB.Status)
			suite.chainB.NextBlock()

			_, commitAckC, err := kC.ReceivePacketCommit(
				suite.chainC.GetContext(),
				p1.GetDestPort(), p1.GetDestChannel(),
				commitC,
			)
			suite.Require().NoError(err)
			suite.Require().Equal(types.COMMIT_STATUS_OK, commitAckC.Status)
			suite.chainC.NextBlock()
		})
	}
}

func (suite *KeeperTestSuite) parsePacketToPacketDataPrepare(cdc codec.Marshaler, p packets.OutgoingPacket) packets.PacketDataPayload {
	var pd packets.PacketData
	suite.Require().NoError(packets.UnmarshalJSONPacketData(p.GetData(), &pd))
	var payload packets.PacketDataPayload
	utils.MustUnmarshalJSONAny(cdc, &payload, pd.GetPayload())
	return payload
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
