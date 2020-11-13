package keeper_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
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
		hasErrorSendCall                     bool
		participantCommitStatus              types.CommitStatus
		participantContractTransactionStatus atomictypes.ContractTransactionStatus
		participantPrepareResult             atomictypes.PrepareResult
		initiatorCommittable                 bool
		expectedResult                       [2][]byte
	}{
		{
			[2]crosstypes.ContractCallInfo{
				samplemodtypes.NewContractCallRequest("counter").ContractCallInfo(suite.chainA.App.AppCodec()),
				samplemodtypes.NewContractCallRequest("counter").ContractCallInfo(suite.chainB.App.AppCodec()),
			},
			false,
			types.COMMIT_STATUS_OK,
			atomictypes.CONTRACT_TRANSACTION_STATUS_COMMIT,
			atomictypes.PREPARE_RESULT_OK,
			true,
			[2][]byte{sdk.Uint64ToBigEndian(1), sdk.Uint64ToBigEndian(1)},
		},
		{
			[2]crosstypes.ContractCallInfo{
				samplemodtypes.NewContractCallRequest("counter").ContractCallInfo(suite.chainA.App.AppCodec()),
				samplemodtypes.NewContractCallRequest("fail").ContractCallInfo(suite.chainB.App.AppCodec()),
			},
			false,
			types.COMMIT_STATUS_FAILED,
			atomictypes.CONTRACT_TRANSACTION_STATUS_ABORT,
			atomictypes.PREPARE_RESULT_FAILED,
			false,
			[2][]byte{nil, nil},
		},
		{
			[2]crosstypes.ContractCallInfo{
				samplemodtypes.NewContractCallRequest("fail").ContractCallInfo(suite.chainA.App.AppCodec()),
				samplemodtypes.NewContractCallRequest("counter").ContractCallInfo(suite.chainB.App.AppCodec()),
			},
			true,
			// the following parameters are ignored
			0, 0, 0, false, [2][]byte{nil, nil},
		},
		{
			[2]crosstypes.ContractCallInfo{
				samplemodtypes.NewContractCallRequest("fail").ContractCallInfo(suite.chainA.App.AppCodec()),
				samplemodtypes.NewContractCallRequest("fail").ContractCallInfo(suite.chainB.App.AppCodec()),
			},
			true,
			// the following parameters are ignored
			0, 0, 0, false, [2][]byte{nil, nil},
		},
		{
			[2]crosstypes.ContractCallInfo{
				samplemodtypes.NewContractCallRequest("counter").ContractCallInfo(suite.chainA.App.AppCodec()),
				samplemodtypes.NewContractCallRequest("counter").ContractCallInfo(suite.chainB.App.AppCodec()),
			},
			false,
			types.COMMIT_STATUS_OK,
			atomictypes.CONTRACT_TRANSACTION_STATUS_COMMIT,
			atomictypes.PREPARE_RESULT_OK,
			true,
			[2][]byte{sdk.Uint64ToBigEndian(2), sdk.Uint64ToBigEndian(2)},
		},
	}

	for i, c := range cases {
		suite.Run(fmt.Sprint(i), func() {
			// check if SendCall call is expected

			txID := []byte(fmt.Sprintf("txid-%v", i))
			chAB := crosstypes.ChannelInfo{Port: channelA.PortID, Channel: channelA.ID}
			cidB, err := crosstypes.PackChainID(&chAB)
			suite.Require().NoError(err)
			cidOurs, err := crosstypes.PackChainID(crosstypes.GetOurChainID())
			suite.Require().NoError(err)

			kA := suite.chainA.App.CrossKeeper.SimpleKeeper()
			txs := []crosstypes.ContractTransaction{
				{
					ChainId: cidOurs,
					Signers: []crosstypes.AccountID{
						crosstypes.AccountID(suite.chainA.SenderAccount.GetAddress()),
					},
					CallInfo: c.callinfos[0],
				},
				{
					ChainId: cidB,
					Signers: []crosstypes.AccountID{
						crosstypes.AccountID(suite.chainB.SenderAccount.GetAddress()),
					},
					CallInfo: c.callinfos[1],
				},
			}
			// TODO add test for verifyTx

			ps := newCapturePacketSender(
				packets.NewBasicPacketSender(suite.chainA.App.IBCKeeper.ChannelKeeper),
			)
			err = kA.SendCall(
				suite.chainA.GetContext(),
				ps,
				txID,
				txs,
				clienttypes.NewHeight(0, uint64(suite.chainA.CurrentHeader.Height)+100),
				0,
			)

			// If the transaction initiator fails, this process is terminated at this point.
			if c.hasErrorSendCall {
				suite.Require().Error(err)

				_, found := kA.GetCoordinatorState(suite.chainA.GetContext(), txID)
				suite.Require().False(found)
				suite.Require().Equal(0, len(ps.Packets()))
				return
			} else {
				suite.Require().NoError(err)
			}

			suite.chainA.NextBlock()

			// check if concurrent access is failed
			suite.Require().Panics(func() {
				ctx, _ := suite.chainA.GetContext().CacheContext()
				_, _, err = suite.chainA.App.SamplemodKeeper.HandleCounter(
					ctx,
					samplemodtypes.NewContractCallRequest("counter"),
				)
			})

			// check if coordinator state is expected

			cs, found := suite.chainA.App.CrossKeeper.SimpleKeeper().GetCoordinatorState(suite.chainA.GetContext(), txID)
			suite.Require().True(found)
			suite.Require().Equal(atomictypes.COORDINATOR_PHASE_PREPARE, cs.Phase)
			suite.Require().Equal(atomictypes.COORDINATOR_DECISION_UNKNOWN, cs.Decision)

			// check if ReceiveCallPacket call is expected

			suite.Require().Equal(1, len(ps.Packets()))
			p0 := ps.Packets()[0]
			var pd0 packets.PacketData
			suite.Require().NoError(packets.UnmarshalJSONPacketData(p0.GetData(), &pd0))
			var payload0 packets.PacketDataPayload
			utils.MustUnmarshalJSONAny(suite.chainB.App.AppCodec(), &payload0, pd0.GetPayload())
			callData := payload0.(*types.PacketDataCall)

			kB := suite.chainB.App.CrossKeeper.SimpleKeeper()
			res, ack, err := kB.ReceiveCallPacket(suite.chainB.GetContext(), p0.GetDestPort(), p0.GetDestChannel(), *callData)
			suite.Require().NoError(err)
			suite.Require().Equal(c.participantCommitStatus, ack.Status)

			// If participant call is success, returns the result data of contract.
			if ack.Status == types.COMMIT_STATUS_OK {
				suite.Require().NotNil(res)
				suite.Require().Equal(res.Data, c.expectedResult[1])
				// check if concurrent access is success
				suite.Require().NotPanics(func() {
					ctx, _ := suite.chainB.GetContext().CacheContext()
					_, _, err = suite.chainB.App.SamplemodKeeper.HandleCounter(
						ctx,
						samplemodtypes.NewContractCallRequest("counter"),
					)
				})
			}
			suite.chainB.NextBlock()
			ctxs, found := kB.GetContractTransactionState(suite.chainB.GetContext(), txID, keeper.TxIndexParticipant)
			suite.Require().True(found)
			suite.Require().Equal(c.participantContractTransactionStatus, ctxs.Status)
			suite.Require().Equal(c.participantPrepareResult, ctxs.PrepareResult)

			// check if ReceiveCallAcknowledgement call is expected

			isCommittable, err := kA.ReceiveCallAcknowledgement(
				suite.chainA.GetContext(),
				channelA.PortID, channelA.ID,
				*ack, txID,
			)
			suite.Require().NoError(err)
			suite.Require().Equal(c.initiatorCommittable, isCommittable)

			// check if TryCommit call is expected

			res, err = kA.TryCommit(suite.chainA.GetContext(), txID, isCommittable)
			suite.Require().NoError(err)
			if c.initiatorCommittable {
				suite.Require().NotNil(res)
				suite.Require().Equal(c.expectedResult[0], res.Data)
			} else {
				suite.Require().Nil(res)
			}
			suite.chainA.NextBlock()
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
