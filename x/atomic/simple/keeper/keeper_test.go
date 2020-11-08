package keeper_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	"github.com/stretchr/testify/suite"

	samplemodtypes "github.com/datachainlab/cross/simapp/samplemod/types"
	atomictypes "github.com/datachainlab/cross/x/atomic/common/types"
	"github.com/datachainlab/cross/x/atomic/simple/keeper"
	"github.com/datachainlab/cross/x/atomic/simple/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
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

	queryClient transfertypes.QueryClient
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))

	queryHelper := baseapp.NewQueryServerTestHelper(suite.chainA.GetContext(), suite.chainA.App.InterfaceRegistry())
	transfertypes.RegisterQueryServer(queryHelper, suite.chainA.App.TransferKeeper)
	suite.queryClient = transfertypes.NewQueryClient(queryHelper)
}

func (suite *KeeperTestSuite) TestCall() {
	// setup

	_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, ibctesting.Tendermint)
	suite.chainB.CreatePortCapability(crosstypes.PortID)
	channelA, _ := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, crosstypes.PortID, crosstypes.PortID, channeltypes.UNORDERED)

	var cases = []struct {
		callinfos                            [2]crosstypes.ContractCallInfo
		participantCommitStatus              types.CommitStatus
		participantContractTransactionStatus atomictypes.ContractTransactionStatus
		participantPrepareResult             atomictypes.PrepareResult
		initiatorCommittable                 bool
	}{
		{
			[2]crosstypes.ContractCallInfo{
				samplemodtypes.NewContractCallRequest("nop").ContractCallInfo(suite.chainA.App.AppCodec()),
				samplemodtypes.NewContractCallRequest("nop").ContractCallInfo(suite.chainB.App.AppCodec()),
			},
			types.COMMIT_STATUS_OK,
			atomictypes.CONTRACT_TRANSACTION_STATUS_COMMIT,
			atomictypes.PREPARE_RESULT_OK,
			true,
		},
		{
			[2]crosstypes.ContractCallInfo{
				samplemodtypes.NewContractCallRequest("nop").ContractCallInfo(suite.chainA.App.AppCodec()),
				samplemodtypes.NewContractCallRequest("fail").ContractCallInfo(suite.chainB.App.AppCodec()),
			},
			types.COMMIT_STATUS_FAILED,
			atomictypes.CONTRACT_TRANSACTION_STATUS_ABORT,
			atomictypes.PREPARE_RESULT_FAILED,
			false,
		},
	}

	for i, c := range cases {
		suite.Run(fmt.Sprint(i), func() {
			// check if SendCall is successful
			txID := []byte(fmt.Sprintf("txid-%v", i))
			chAB := crosstypes.ChannelInfo{Port: channelA.PortID, Channel: channelA.ID}
			selfCh := crosstypes.ChannelInfo{}
			cidB, err := crosstypes.PackChainID(&chAB)
			suite.Require().NoError(err)
			selfCid, err := crosstypes.PackChainID(&selfCh)

			kA := suite.chainA.App.CrossKeeper.SimpleKeeper()
			txs := []crosstypes.ContractTransaction{
				{
					ChainId:  *selfCid,
					Signers:  []crosstypes.AccountAddress{suite.chainA.SenderAccount.GetAddress().Bytes()},
					CallInfo: c.callinfos[0],
				},
				{
					ChainId:  *cidB,
					Signers:  []crosstypes.AccountAddress{suite.chainB.SenderAccount.GetAddress().Bytes()},
					CallInfo: c.callinfos[1],
				},
			}
			ps := newCapturePacketSender(
				packets.NewBasicPacketSender(suite.chainA.App.IBCKeeper.ChannelKeeper),
			)
			suite.Require().NoError(
				kA.SendCall(
					suite.chainA.GetContext(),
					ps,
					txID,
					txs,
				),
			)
			suite.chainA.NextBlock()

			// check if coordinator state is expected

			cs, found := suite.chainA.App.CrossKeeper.SimpleKeeper().GetCoordinatorState(suite.chainA.GetContext(), txID)
			suite.Require().True(found)
			suite.Require().Equal(atomictypes.COORDINATOR_PHASE_PREPARE, cs.Phase)

			// check if ReceiveCallPacket is successful

			suite.Require().Equal(1, len(ps.Packets()))
			p0 := ps.Packets()[0]
			var pd0 packets.PacketData
			suite.Require().NoError(packets.UnmarshalJSONPacketData(p0.GetData(), &pd0))
			var payload0 packets.PacketDataPayload
			utils.MustUnmarshalJSONAny(suite.chainB.App.AppCodec(), &payload0, pd0.GetPayload())
			callData := payload0.(*types.PacketDataCall)

			kB := suite.chainB.App.CrossKeeper.SimpleKeeper()
			ack, err := kB.ReceiveCallPacket(suite.chainB.GetContext(), p0.GetDestPort(), p0.GetDestChannel(), *callData)
			suite.Require().NoError(err)
			suite.Equal(c.participantCommitStatus, ack.Status)
			ctxs, found := suite.chainB.App.CrossKeeper.SimpleKeeper().GetContractTransactionState(suite.chainB.GetContext(), txID, keeper.TxIndexParticipant)
			suite.Require().True(found)
			suite.Require().Equal(c.participantContractTransactionStatus, ctxs.Status)
			suite.Require().Equal(c.participantPrepareResult, ctxs.PrepareResult)

			// check if ReceiveCallAcknowledgement is successful

			isCommittable, err := suite.chainA.App.CrossKeeper.SimpleKeeper().ReceiveCallAcknowledgement(
				suite.chainA.GetContext(),
				channelA.PortID, channelA.ID,
				*ack, txID,
			)
			suite.Require().NoError(err)
			suite.Require().Equal(c.initiatorCommittable, isCommittable)

			// TODO try to commit
		})
	}

}

type capturePacketSender struct {
	inner   packets.PacketSender
	packets []packets.OutgoingPacket
}

var _ packets.PacketSender = (*capturePacketSender)(nil)

func newCapturePacketSender(ps packets.PacketSender) *capturePacketSender {
	return &capturePacketSender{inner: ps}
}

func (ps *capturePacketSender) SendPacket(
	ctx sdk.Context,
	channelCap *capabilitytypes.Capability,
	packet packets.OutgoingPacket,
) error {
	if err := ps.inner.SendPacket(ctx, channelCap, packet); err != nil {
		return err
	}
	ps.packets = append(ps.packets, packet)
	return nil
}

func (ps *capturePacketSender) Packets() []packets.OutgoingPacket {
	return ps.packets
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
