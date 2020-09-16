package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/ibc/contract"
	"github.com/datachainlab/cross/x/ibc/cross"
	"github.com/datachainlab/cross/x/ibc/cross/types"
	"github.com/datachainlab/cross/x/ibc/cross/types/simple"
	"github.com/datachainlab/cross/x/ibc/store/lock"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestSimpleKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(SimpleKeeperTestSuite))
}

type SimpleKeeperTestSuite struct {
	KeeperTestSuite

	app0 *appContext // coordinator
	app1 *appContext // participant

	initiator   sdk.AccAddress
	coordinator sdk.AccAddress
	relayer     sdk.AccAddress

	signer1 sdk.AccAddress
	signer2 sdk.AccAddress

	ch0to1 cross.ChannelInfo
	ch1to0 cross.ChannelInfo
}

func (suite *SimpleKeeperTestSuite) SetupTest() {
	suite.app0 = suite.createAppWithHeader(abci.Header{ChainID: "app0"}, func(k contract.Keeper) cross.ContractHandler {
		return suite.createContractHandler(k, "c1")
	})
	suite.app1 = suite.createAppWithHeader(abci.Header{ChainID: "app1"}, func(k contract.Keeper) cross.ContractHandler {
		return suite.createContractHandler(k, "c2")
	})

	suite.initiator = sdk.AccAddress("initiator")
	suite.coordinator = sdk.AccAddress("coordinator")
	suite.relayer = sdk.AccAddress("relayer")

	suite.signer1 = sdk.AccAddress("signer1")
	suite.signer2 = sdk.AccAddress("signer2")

	suite.ch0to1 = cross.NewChannelInfo("testportzeroone", "testchannelzeroone") // app0 -> app1
	suite.ch1to0 = cross.NewChannelInfo("testportonezero", "testchannelonezero") // app1 -> app0
}

func (suite *SimpleKeeperTestSuite) TestCall() {
	ci1 := contract.NewContractCallInfo("c1", "issue", [][]byte{[]byte("tone"), []byte("80")})
	ci2 := contract.NewContractCallInfo("c2", "issue", [][]byte{[]byte("ttwo"), []byte("60")})

	var nonce uint64 = 1
	var tss = []cross.ContractTransaction{
		cross.NewContractTransaction(
			cross.NewChannelInfo("", ""),
			[]sdk.AccAddress{suite.signer1},
			ci1.Bytes(),
			cross.NewStateConstraint(
				cross.PostStateConstraint,
				[]cross.OP{lock.WriteOP{K: suite.signer1, V: marshalCoins(sdk.Coins{sdk.NewInt64Coin("tone", 80)})}},
			),
			nil,
			nil,
		),
		cross.NewContractTransaction(
			suite.ch0to1,
			[]sdk.AccAddress{suite.signer2},
			ci2.Bytes(),
			cross.NewStateConstraint(
				cross.PostStateConstraint,
				[]cross.OP{lock.WriteOP{K: suite.signer2, V: marshalCoins(sdk.Coins{sdk.NewInt64Coin("ttwo", 60)})}},
			),
			nil,
			nil,
		),
	}

	// {
	// 	msg := cross.NewMsgInitiate(
	// 		suite.initiator,
	// 		suite.app0.chainID,
	// 		tss,
	// 		5,
	// 		nonce,
	// 		cross.COMMIT_PROTOCOL_SIMPLE,
	// 	)
	// 	_, err := suite.app0.app.CrossKeeper.SimpleKeeper().SendCall(
	// 		suite.app0.ctx,
	// 		suite.app0.app.ContractHandler,
	// 		msg,
	// 		msg.ContractTransactions,
	// 	)
	// 	suite.Error(err) // channel does not exist
	// }

	suite.openChannels(
		suite.app1.chainID,
		suite.app0.chainID+suite.app1.chainID,
		suite.ch0to1,
		suite.app0,

		suite.app0.chainID,
		suite.app1.chainID+suite.app0.chainID,
		suite.ch1to0,
		suite.app1,
	)

	// ensure that current block height is correct
	suite.EqualValues(2, suite.app0.ctx.BlockHeight())

	msg := cross.NewMsgInitiate(
		suite.initiator,
		suite.app0.chainID,
		tss,
		3,
		nonce,
		cross.COMMIT_PROTOCOL_TPC,
	)

	txID, err := suite.app0.app.CrossKeeper.SimpleKeeper().SendCall(
		suite.app0.ctx,
		types.NewSimplePacketSender(suite.app0.app.IBCKeeper.ChannelKeeper),
		suite.app0.app.ContractHandler,
		msg,
		msg.ContractTransactions,
	)
	suite.NoError(err) // successfully executed

	ci, found := suite.app0.app.CrossKeeper.SimpleKeeper().GetCoordinator(suite.app0.ctx, txID)
	if suite.True(found) {
		suite.Equal(ci.Status, cross.CO_STATUS_INIT)
	}

	nextSeqSend := uint64(1)
	packetCommitment := suite.app0.app.IBCKeeper.ChannelKeeper.GetPacketCommitment(suite.app0.ctx, suite.ch0to1.Port, suite.ch0to1.Channel, nextSeqSend)
	suite.NotNil(packetCommitment)

	nextSeqSend += 1
	lkr, err := types.MakeLinker(tss)
	if err != nil {
		suite.FailNow(err.Error())
	}
	objs1, err := lkr.Resolve(tss[1].Links)
	if err != nil {
		suite.FailNow(err.Error())
	}
	data := simple.NewPacketDataCall(suite.relayer, txID, types.NewContractTransactionInfo(tss[1], objs1))
	status, err := suite.app1.app.CrossKeeper.SimpleKeeper().ReceiveCallPacket(suite.app1.ctx, suite.app1.app.ContractHandler, suite.ch1to0.Port, suite.ch1to0.Channel, data)
	suite.NoError(err)
	suite.Equal(types.PREPARE_RESULT_OK, status)

	isCommittable, err := suite.app0.app.CrossKeeper.SimpleKeeper().ReceiveCallAcknowledgement(suite.app0.ctx, suite.ch0to1.Port, suite.ch0to1.Channel, simple.NewPacketCallAcknowledgement(status), txID)
	suite.NoError(err)
	suite.True(isCommittable)

	_, err = suite.app0.app.CrossKeeper.SimpleKeeper().TryCommit(suite.app0.ctx, suite.app0.app.ContractHandler, txID, isCommittable)
	suite.NoError(err)
}
