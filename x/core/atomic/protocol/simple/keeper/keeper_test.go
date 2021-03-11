package keeper_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	"github.com/stretchr/testify/suite"

	samplemodtypes "github.com/datachainlab/cross/simapp/samplemod/types"
	accounttypes "github.com/datachainlab/cross/x/core/account/types"
	"github.com/datachainlab/cross/x/core/atomic/protocol/simple/keeper"
	"github.com/datachainlab/cross/x/core/atomic/protocol/simple/types"
	atomictypes "github.com/datachainlab/cross/x/core/atomic/types"
	contracttypes "github.com/datachainlab/cross/x/core/contract/types"
	initiatortypes "github.com/datachainlab/cross/x/core/initiator/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
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
	channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, crosstypes.PortID, crosstypes.PortID, channeltypes.UNORDERED)

	chAB := xcctypes.ChannelInfo{Port: channelA.PortID, Channel: channelA.ID}
	xccB, err := xcctypes.PackCrossChainChannel(&chAB)
	suite.Require().NoError(err)
	xccSelf, err := xcctypes.PackCrossChainChannel(
		suite.chainA.App.XCCResolver.GetSelfCrossChainChannel(suite.chainA.GetContext()),
	)
	suite.Require().NoError(err)

	var cases = []struct {
		name                                 string
		txs                                  [2]initiatortypes.ContractTransaction
		hasErrorSendCall                     bool
		participantCommitStatus              types.CommitStatus
		participantContractTransactionStatus atomictypes.ContractTransactionStatus
		participantPrepareResult             atomictypes.PrepareResult
		concurrentAccessCheck                bool
		initiatorCommittable                 bool
		expectedResult                       [2][]byte
	}{
		{
			"case0",
			[2]initiatortypes.ContractTransaction{
				{
					CrossChainChannel: xccSelf,
					Signers: []accounttypes.AccountID{
						accounttypes.AccountID(suite.chainA.SenderAccount.GetAddress()),
					},
					CallInfo: samplemodtypes.NewContractCallRequest("counter").ContractCallInfo(suite.chainA.App.AppCodec()),
				},
				{
					CrossChainChannel: xccB,
					Signers: []accounttypes.AccountID{
						accounttypes.AccountID(suite.chainB.SenderAccount.GetAddress()),
					},
					CallInfo: samplemodtypes.NewContractCallRequest("counter").ContractCallInfo(suite.chainB.App.AppCodec()),
				},
			},
			false,
			types.COMMIT_STATUS_OK,
			atomictypes.CONTRACT_TRANSACTION_STATUS_COMMIT,
			atomictypes.PREPARE_RESULT_OK,
			true,
			true,
			[2][]byte{sdk.Uint64ToBigEndian(1), sdk.Uint64ToBigEndian(1)},
		},
		{
			"case1",
			[2]initiatortypes.ContractTransaction{
				{
					CrossChainChannel: xccSelf,
					Signers: []accounttypes.AccountID{
						accounttypes.AccountID(suite.chainA.SenderAccount.GetAddress()),
					},
					CallInfo: samplemodtypes.NewContractCallRequest("counter").ContractCallInfo(suite.chainA.App.AppCodec()),
				},
				{
					CrossChainChannel: xccB,
					Signers: []accounttypes.AccountID{
						accounttypes.AccountID(suite.chainB.SenderAccount.GetAddress()),
					},
					CallInfo: samplemodtypes.NewContractCallRequest("fail").ContractCallInfo(suite.chainB.App.AppCodec()),
				},
			},
			false,
			types.COMMIT_STATUS_FAILED,
			atomictypes.CONTRACT_TRANSACTION_STATUS_ABORT,
			atomictypes.PREPARE_RESULT_FAILED,
			true,
			false,
			[2][]byte{nil, nil},
		},
		{
			"case2",
			[2]initiatortypes.ContractTransaction{
				{
					CrossChainChannel: xccSelf,
					Signers: []accounttypes.AccountID{
						accounttypes.AccountID(suite.chainA.SenderAccount.GetAddress()),
					},
					CallInfo: samplemodtypes.NewContractCallRequest("fail").ContractCallInfo(suite.chainA.App.AppCodec()),
				},
				{
					CrossChainChannel: xccB,
					Signers: []accounttypes.AccountID{
						accounttypes.AccountID(suite.chainB.SenderAccount.GetAddress()),
					},
					CallInfo: samplemodtypes.NewContractCallRequest("counter").ContractCallInfo(suite.chainB.App.AppCodec()),
				},
			},
			true,
			// the following parameters are ignored
			0, 0, 0, false, false, [2][]byte{nil, nil},
		},
		{
			"case3",
			[2]initiatortypes.ContractTransaction{
				{
					CrossChainChannel: xccSelf,
					Signers: []accounttypes.AccountID{
						accounttypes.AccountID(suite.chainA.SenderAccount.GetAddress()),
					},
					CallInfo: samplemodtypes.NewContractCallRequest("fail").ContractCallInfo(suite.chainA.App.AppCodec()),
				},
				{
					CrossChainChannel: xccB,
					Signers: []accounttypes.AccountID{
						accounttypes.AccountID(suite.chainB.SenderAccount.GetAddress()),
					},
					CallInfo: samplemodtypes.NewContractCallRequest("fail").ContractCallInfo(suite.chainB.App.AppCodec()),
				},
			},
			true,
			// the following parameters are ignored
			0, 0, 0, false, false, [2][]byte{nil, nil},
		},
		{
			"case4",
			[2]initiatortypes.ContractTransaction{
				{
					CrossChainChannel: xccSelf,
					Signers: []accounttypes.AccountID{
						accounttypes.AccountID(suite.chainA.SenderAccount.GetAddress()),
					},
					CallInfo: samplemodtypes.NewContractCallRequest("counter").ContractCallInfo(suite.chainA.App.AppCodec()),
				},
				{
					CrossChainChannel: xccB,
					Signers: []accounttypes.AccountID{
						accounttypes.AccountID(suite.chainB.SenderAccount.GetAddress()),
					},
					CallInfo: samplemodtypes.NewContractCallRequest("counter").ContractCallInfo(suite.chainB.App.AppCodec()),
				},
			},
			false,
			types.COMMIT_STATUS_OK,
			atomictypes.CONTRACT_TRANSACTION_STATUS_COMMIT,
			atomictypes.PREPARE_RESULT_OK,
			true,
			true,
			[2][]byte{sdk.Uint64ToBigEndian(2), sdk.Uint64ToBigEndian(2)},
		},
		{
			"case5",
			[2]initiatortypes.ContractTransaction{
				{
					CrossChainChannel: xccSelf,
					Signers: []accounttypes.AccountID{
						accounttypes.AccountID(suite.chainA.SenderAccount.GetAddress()),
					},
					CallInfo: samplemodtypes.NewContractCallRequest(
						"external-call",
						hex.EncodeToString(suite.chainB.SenderAccount.GetAddress()),
						channelA.ID,
					).ContractCallInfo(suite.chainA.App.AppCodec()),
					Links: []initiatortypes.Link{{SrcIndex: 1}},
				},
				{
					CrossChainChannel: xccB,
					Signers: []accounttypes.AccountID{
						accounttypes.AccountID(suite.chainB.SenderAccount.GetAddress()),
					},
					CallInfo:    samplemodtypes.NewContractCallRequest("counter").ContractCallInfo(suite.chainB.App.AppCodec()),
					ReturnValue: txtypes.NewReturnValue(sdk.Uint64ToBigEndian(3)),
				},
			},
			false,
			types.COMMIT_STATUS_OK,
			atomictypes.CONTRACT_TRANSACTION_STATUS_COMMIT,
			atomictypes.PREPARE_RESULT_OK,
			false,
			true,
			[2][]byte{sdk.Uint64ToBigEndian(3), sdk.Uint64ToBigEndian(3)},
		},
		{
			"case6",
			[2]initiatortypes.ContractTransaction{
				{
					CrossChainChannel: xccSelf,
					Signers: []accounttypes.AccountID{
						accounttypes.AccountID(suite.chainA.SenderAccount.GetAddress()),
					},
					CallInfo:    samplemodtypes.NewContractCallRequest("counter").ContractCallInfo(suite.chainA.App.AppCodec()),
					ReturnValue: txtypes.NewReturnValue(sdk.Uint64ToBigEndian(3)),
				},
				{
					CrossChainChannel: xccB,
					Signers: []accounttypes.AccountID{
						accounttypes.AccountID(suite.chainB.SenderAccount.GetAddress()),
					},
					CallInfo: samplemodtypes.NewContractCallRequest(
						"external-call",
						hex.EncodeToString(suite.chainA.SenderAccount.GetAddress()),
						channelB.ID,
					).ContractCallInfo(suite.chainB.App.AppCodec()),
					Links: []initiatortypes.Link{{SrcIndex: 0}},
				},
			},
			false,
			types.COMMIT_STATUS_OK,
			atomictypes.CONTRACT_TRANSACTION_STATUS_COMMIT,
			atomictypes.PREPARE_RESULT_OK,
			false,
			true,
			[2][]byte{sdk.Uint64ToBigEndian(3), sdk.Uint64ToBigEndian(3)},
		},
	}

	for i, c := range cases {
		suite.Run(c.name, func() {
			txs, err := suite.chainA.App.CrossKeeper.InitiatorKeeper().ResolveTransactions(
				suite.chainA.GetContext(),
				c.txs[:],
			)
			suite.Require().NoError(err)

			// check if SendCall call is expected

			txID := []byte(fmt.Sprintf("txid-%v", i))
			kA := suite.chainA.App.AtomicKeeper.SimpleKeeper()

			ps := ibctesting.NewCapturePacketSender(
				packets.NewBasicPacketSender(suite.chainA.App.IBCKeeper.ChannelKeeper),
			)
			suite.Require().NoError(
				kA.SendCall(
					suite.chainA.GetContext(),
					ps,
					txID,
					txs,
					clienttypes.NewHeight(0, uint64(suite.chainA.CurrentHeader.Height)+100),
					0,
				),
			)

			// check if coordinator state is expected
			cs, found := kA.GetCoordinatorState(suite.chainA.GetContext(), txID)
			suite.Require().True(found)
			if c.hasErrorSendCall {
				suite.Require().Equal(atomictypes.COORDINATOR_PHASE_COMMIT, cs.Phase)
				suite.Require().Equal(atomictypes.COORDINATOR_DECISION_ABORT, cs.Decision)
				suite.Require().Equal(0, len(ps.Packets()))
				return
			} else {
				suite.Require().Equal(atomictypes.COORDINATOR_PHASE_PREPARE, cs.Phase)
				suite.Require().Equal(atomictypes.COORDINATOR_DECISION_UNKNOWN, cs.Decision)
				suite.Require().Equal(1, len(ps.Packets()))
			}

			suite.chainA.NextBlock()

			if c.concurrentAccessCheck {
				// check if concurrent access is failed
				suite.Require().Panics(func() {
					ctx, _ := suite.chainA.GetContext().CacheContext()
					ctx = contracttypes.SetupContractContext(ctx, c.txs[0].Signers, contracttypes.ContractRuntimeInfo{CommitMode: contracttypes.BasicMode})
					_, err = suite.chainA.App.SamplemodKeeper.HandleCounter(
						ctx,
						samplemodtypes.NewContractCallRequest("counter"),
					)
				})
			}

			// check if coordinator state is expected

			cs, found = suite.chainA.App.AtomicKeeper.SimpleKeeper().GetCoordinatorState(suite.chainA.GetContext(), txID)
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

			kB := suite.chainB.App.AtomicKeeper.SimpleKeeper()
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
					ctx = contracttypes.SetupContractContext(ctx, c.txs[1].Signers, contracttypes.ContractRuntimeInfo{CommitMode: contracttypes.BasicMode})
					_, err = suite.chainB.App.SamplemodKeeper.HandleCounter(
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

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
