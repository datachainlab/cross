package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"

	samplemodtypes "github.com/datachainlab/cross/simapp/samplemod/types"
	accounttypes "github.com/datachainlab/cross/x/account/types"
	crosshost "github.com/datachainlab/cross/x/core/host"
	initiatortypes "github.com/datachainlab/cross/x/core/initiator/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
	ibctesting "github.com/datachainlab/cross/x/ibc/testing"
	"github.com/datachainlab/cross/x/packets"
	"github.com/datachainlab/cross/x/utils"
	"github.com/stretchr/testify/suite"
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

func (suite *KeeperTestSuite) TestInitiateTx() {
	// setup

	_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, ibctesting.Tendermint)
	suite.chainB.CreatePortCapability(crosshost.PortID)

	channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, crosshost.PortID, crosshost.PortID, channeltypes.UNORDERED)

	chAB := xcctypes.ChannelInfo{Port: channelA.PortID, Channel: channelA.ID}
	xccB, err := xcctypes.PackCrossChainChannel(&chAB)

	chBA := xcctypes.ChannelInfo{Port: channelB.PortID, Channel: channelB.ID}

	xccSelf, err := xcctypes.PackCrossChainChannel(suite.chainA.App.XCCResolver.GetSelfCrossChainChannel(suite.chainA.GetContext()))
	suite.Require().NoError(err)

	txs := []initiatortypes.ContractTransaction{
		{
			CrossChainChannel: xccSelf,
			Signers: []accounttypes.AccountID{
				accounttypes.AccountID(suite.chainA.SenderAccount.GetAddress()),
			},
			CallInfo: samplemodtypes.NewContractCallRequest("nop").ContractCallInfo(suite.chainA.App.AppCodec()),
		},
		{
			CrossChainChannel: xccB,
			Signers: []accounttypes.AccountID{
				accounttypes.AccountID(suite.chainB.SenderAccount.GetAddress()),
			},
			CallInfo: samplemodtypes.NewContractCallRequest("nop").ContractCallInfo(suite.chainB.App.AppCodec()),
		},
	}

	// InitiateTx on chainA

	msg0 := &initiatortypes.MsgInitiateTx{
		Sender:               suite.chainA.SenderAccount.GetAddress().Bytes(),
		ChainId:              suite.chainA.ChainID,
		Nonce:                0,
		CommitProtocol:       txtypes.COMMIT_PROTOCOL_SIMPLE,
		ContractTransactions: txs,
		Signers: []accounttypes.AccountID{
			suite.chainA.SenderAccount.GetAddress().Bytes(),
		},
		TimeoutHeight: clienttypes.NewHeight(0, uint64(suite.chainA.CurrentHeader.Height)+100),
	}
	suite.Require().NoError(msg0.ValidateBasic())

	res0, err := suite.chainA.App.CrossKeeper.InitiateTx(
		sdk.WrapSDKContext(suite.chainA.GetContext()),
		msg0,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(res0.Status, initiatortypes.INITIATE_TX_STATUS_PENDING)
	suite.chainA.NextBlock()

	// IBCSignTx on chainB
	ps := ibctesting.NewCapturePacketSender(
		packets.NewBasicPacketSender(suite.chainB.App.IBCKeeper.ChannelKeeper),
	)
	err = suite.chainB.App.CrossKeeper.InitiatorKeeper().SendIBCSignTx(
		suite.chainB.GetContext(),
		ps,
		&chBA,
		res0.TxID,
		[]accounttypes.AccountID{suite.chainB.SenderAccount.GetAddress().Bytes()},
		clienttypes.NewHeight(0, uint64(suite.chainB.CurrentHeader.Height)+100),
		0,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(ps.Packets()))
	suite.chainB.NextBlock()

	p0 := ps.Packets()[0]
	var pd0 packets.PacketData
	suite.Require().NoError(packets.UnmarshalJSONPacketData(p0.GetData(), &pd0))
	var payload0 packets.PacketDataPayload
	utils.MustUnmarshalJSONAny(suite.chainB.App.AppCodec(), &payload0, pd0.GetPayload())
	signData := payload0.(*initiatortypes.PacketDataIBCSignTx)

	// ReceiveIBCSignTx on chainA

	completed, err := suite.chainA.App.CrossKeeper.InitiatorKeeper().ReceiveIBCSignTx(
		suite.chainA.GetContext(),
		p0.GetDestPort(), p0.GetDestChannel(),
		*signData,
	)
	suite.Require().NoError(err)
	suite.Require().True(completed)
	suite.chainA.NextBlock()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
