package keeper_test

import (
	"testing"

	"github.com/datachainlab/cross/example/simapp"
	"github.com/datachainlab/cross/x/ibc/contract"
	"github.com/datachainlab/cross/x/ibc/cross"
	"github.com/datachainlab/cross/x/ibc/cross/types"
	lock "github.com/datachainlab/cross/x/ibc/store/lock"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tpctypes "github.com/datachainlab/cross/x/ibc/cross/types/tpc"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

func TestTPCKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(TPCKeeperTestSuite))
}

type TPCKeeperTestSuite struct {
	KeeperTestSuite

	initiator sdk.AccAddress
	signer1   sdk.AccAddress
	signer2   sdk.AccAddress
	signer3   sdk.AccAddress

	app0 *appContext
	app1 *appContext
	app2 *appContext

	chd1 cross.ContractHandler
	chd2 cross.ContractHandler

	ch0to1 cross.ChannelInfo
	ch1to0 cross.ChannelInfo
	ch0to2 cross.ChannelInfo
	ch2to0 cross.ChannelInfo
}

func (suite *TPCKeeperTestSuite) SetupTest() {
	suite.setup(types.ChannelInfoResolver{})
}

func (suite *TPCKeeperTestSuite) setup(channelResolver types.ChannelResolver) {
	suite.initiator = sdk.AccAddress("initiator")
	suite.signer1 = sdk.AccAddress("signer1")

	suite.app0 = suite.createAppWithHeader(abci.Header{ChainID: "app0"}, simapp.DefaultContractHandlerProvider, func() types.ChannelResolver { return channelResolver }) // coordinator node

	suite.app1 = suite.createAppWithHeader(abci.Header{ChainID: "app1"}, func(k contract.Keeper, r types.ChannelResolver) cross.ContractHandler {
		return suite.createContractHandler(k, "c1", r)
	}, func() types.ChannelResolver { return channelResolver })
	suite.chd1 = suite.createContractHandler(contract.NewKeeper(suite.app1.cdc, suite.app1.app.GetKey(cross.StoreKey)), "c1", channelResolver)

	suite.app2 = suite.createAppWithHeader(abci.Header{ChainID: "app2"}, func(k contract.Keeper, r types.ChannelResolver) cross.ContractHandler {
		return suite.createContractHandler(k, "c2", r)
	}, func() types.ChannelResolver { return channelResolver })
	suite.chd2 = suite.createContractHandler(contract.NewKeeper(suite.app2.cdc, suite.app2.app.GetKey(cross.StoreKey)), "c2", channelResolver)

	suite.signer2 = sdk.AccAddress("signer2")
	suite.signer3 = sdk.AccAddress("signer3")

	suite.ch0to1 = cross.NewChannelInfo("cross", "testchannelone")  // app0 -> app1
	suite.ch1to0 = cross.NewChannelInfo("cross", "testchannelzero") // app1 -> app0
	suite.ch0to2 = cross.NewChannelInfo("cross", "testchanneltwo")  // app0 -> app2
	suite.ch2to0 = cross.NewChannelInfo("cross", "testchannelzero") // app2 -> app0
}

func (suite *TPCKeeperTestSuite) TestInitiateMsg() {
	ci1 := contract.NewContractCallInfo("c1", "issue", [][]byte{[]byte("tone"), []byte("80")})
	ci2 := contract.NewContractCallInfo("c2", "issue", [][]byte{[]byte("ttwo"), []byte("60")})

	var nonce uint64 = 1
	var tss = []cross.ContractTransaction{
		cross.NewContractTransaction(
			suite.ch0to1,
			[]sdk.AccAddress{suite.signer1},
			ci1.Bytes(),
			cross.NewStateConstraint(
				cross.ExactMatchStateConstraint,
				[]cross.OP{lock.WriteOP{K: suite.signer1, V: marshalCoins(sdk.Coins{sdk.NewInt64Coin("tone", 80)})}},
			),
			nil,
			nil,
		),
		cross.NewContractTransaction(
			suite.ch0to2,
			[]sdk.AccAddress{suite.signer2},
			ci2.Bytes(),
			cross.NewStateConstraint(
				cross.ExactMatchStateConstraint,
				[]cross.OP{lock.WriteOP{K: suite.signer2, V: marshalCoins(sdk.Coins{sdk.NewInt64Coin("ttwo", 60)})}},
			),
			nil,
			nil,
		),
	}

	{
		msg := cross.NewMsgInitiate(
			suite.initiator,
			suite.app0.chainID,
			tss,
			5,
			nonce,
			cross.COMMIT_PROTOCOL_TPC,
		)
		_, err := suite.app0.app.CrossKeeper.TPCKeeper().MulticastPreparePacket(
			suite.app0.ctx,
			types.NewSimplePacketSender(suite.app0.app.IBCKeeper.ChannelKeeper),
			suite.initiator,
			msg,
			msg.ContractTransactions,
		)
		suite.Error(err) // channel does not exist
	}

	// Try to open a channel and connection between app0 and app1, app2

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

	suite.openChannels(
		suite.app2.chainID,
		suite.app0.chainID+suite.app2.chainID,
		suite.ch0to2,
		suite.app0,

		suite.app0.chainID,
		suite.app2.chainID+suite.app1.chainID,
		suite.ch2to0,
		suite.app2,
	)

	// ensure that current block height is correct
	suite.EqualValues(3, suite.app0.ctx.BlockHeight())

	{
		msg := cross.NewMsgInitiate(
			suite.initiator,
			suite.app0.chainID,
			tss,
			3,
			nonce,
			cross.COMMIT_PROTOCOL_TPC,
		)
		_, err := suite.app0.app.CrossKeeper.TPCKeeper().MulticastPreparePacket(
			suite.app0.ctx,
			types.NewSimplePacketSender(suite.app0.app.IBCKeeper.ChannelKeeper),
			suite.initiator,
			msg,
			msg.ContractTransactions,
		)
		suite.Error(err) // timeout error
	}

	{
		msg := cross.NewMsgInitiate(
			suite.initiator,
			"dummy", // invalid chainID
			tss,
			4,
			nonce,
			cross.COMMIT_PROTOCOL_TPC,
		)
		_, err := suite.app0.app.CrossKeeper.TPCKeeper().MulticastPreparePacket(
			suite.app0.ctx,
			types.NewSimplePacketSender(suite.app0.app.IBCKeeper.ChannelKeeper),
			suite.initiator,
			msg,
			msg.ContractTransactions,
		)
		suite.Error(err) // occur an error due to invalid chainID
	}

	{
		msg := cross.NewMsgInitiate(
			suite.initiator,
			suite.app0.chainID,
			tss,
			4,
			nonce,
			cross.COMMIT_PROTOCOL_TPC,
		)
		_, err := suite.app0.app.CrossKeeper.TPCKeeper().MulticastPreparePacket(
			suite.app0.ctx,
			types.NewSimplePacketSender(suite.app0.app.IBCKeeper.ChannelKeeper),
			suite.initiator,
			msg,
			msg.ContractTransactions,
		)
		suite.NoError(err) // successfully executed
	}
}

func (suite *TPCKeeperTestSuite) openAllChannels() {
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

	suite.openChannels(
		suite.app2.chainID,
		suite.app0.chainID+suite.app2.chainID,
		suite.ch0to2,
		suite.app0,

		suite.app0.chainID,
		suite.app2.chainID+suite.app1.chainID,
		suite.ch2to0,
		suite.app2,
	)
}

func makeTransactionInfo(tx cross.ContractTransaction, links ...cross.Object) cross.ContractTransactionInfo {
	return cross.ContractTransactionInfo{Transaction: tx, LinkObjects: links}
}

func (suite *TPCKeeperTestSuite) TestRelay() {
	ci1 := contract.NewContractCallInfo("c1", "issue", [][]byte{[]byte("tone"), []byte("80")})
	// app2 has multiple contract calls
	ci2 := contract.NewContractCallInfo("c2", "issue", [][]byte{[]byte("ttwo"), []byte("60")})
	ci3 := contract.NewContractCallInfo("c2", "issue", [][]byte{[]byte("tthree"), []byte("40")})

	var err error
	var nonce uint64 = 1
	var tss = []cross.ContractTransaction{
		cross.NewContractTransaction(
			suite.ch0to1,
			[]sdk.AccAddress{suite.signer1},
			ci1.Bytes(),
			cross.NewStateConstraint(
				cross.ExactMatchStateConstraint,
				[]cross.OP{
					lock.ReadOP{K: suite.signer1, V: nil},
					lock.WriteOP{K: suite.signer1, V: marshalCoins(sdk.Coins{sdk.NewInt64Coin("tone", 80)})},
				},
			),
			cross.NewReturnValue(marshalCoin(sdk.NewInt64Coin("tone", 80))),
			nil,
		),
		cross.NewContractTransaction(
			suite.ch0to2,
			[]sdk.AccAddress{suite.signer2},
			ci2.Bytes(),
			cross.NewStateConstraint(
				cross.ExactMatchStateConstraint,
				[]cross.OP{
					lock.ReadOP{K: suite.signer2, V: nil},
					lock.WriteOP{K: suite.signer2, V: marshalCoins(sdk.Coins{sdk.NewInt64Coin("ttwo", 60)})},
				},
			),
			cross.NewReturnValue(marshalCoin(sdk.NewInt64Coin("ttwo", 60))),
			nil,
		),
		cross.NewContractTransaction(
			suite.ch0to2,
			[]sdk.AccAddress{suite.signer3},
			ci3.Bytes(),
			cross.NewStateConstraint(
				cross.ExactMatchStateConstraint,
				[]cross.OP{
					lock.ReadOP{K: suite.signer3, V: nil},
					lock.WriteOP{K: suite.signer3, V: marshalCoins(sdk.Coins{sdk.NewInt64Coin("tthree", 40)})},
				},
			),
			cross.NewReturnValue(marshalCoin(sdk.NewInt64Coin("tthree", 40))),
			nil,
		),
	}

	msg := cross.NewMsgInitiate(
		suite.initiator,
		suite.app0.chainID,
		tss,
		256,
		nonce,
		cross.COMMIT_PROTOCOL_TPC,
	)
	_, err = suite.app0.app.CrossKeeper.TPCKeeper().MulticastPreparePacket(
		suite.app0.ctx,
		types.NewSimplePacketSender(suite.app0.app.IBCKeeper.ChannelKeeper),
		suite.initiator,
		msg,
		msg.ContractTransactions,
	)
	suite.Error(err) // channel does not exist

	// Try to open a channel and connection between app0 and app1, app2
	suite.openAllChannels()

	txID, err := suite.app0.app.CrossKeeper.TPCKeeper().MulticastPreparePacket(
		suite.app0.ctx,
		types.NewSimplePacketSender(suite.app0.app.IBCKeeper.ChannelKeeper),
		suite.initiator,
		msg,
		msg.ContractTransactions,
	)
	suite.NoError(err) // successfully executed

	ci, found := suite.app0.app.CrossKeeper.TPCKeeper().GetCoordinator(suite.app0.ctx, txID)
	if suite.True(found) {
		suite.Equal(ci.Status, cross.CO_STATUS_INIT)
	}

	nextSeqSend := uint64(1)
	packetCommitment := suite.app0.app.IBCKeeper.ChannelKeeper.GetPacketCommitment(suite.app0.ctx, suite.ch0to1.Port, suite.ch0to1.Channel, nextSeqSend)
	suite.NotNil(packetCommitment)
	packetCommitment = suite.app0.app.IBCKeeper.ChannelKeeper.GetPacketCommitment(suite.app0.ctx, suite.ch0to2.Port, suite.ch0to2.Channel, nextSeqSend)
	suite.NotNil(packetCommitment)
	packetCommitment = suite.app0.app.IBCKeeper.ChannelKeeper.GetPacketCommitment(suite.app0.ctx, suite.ch0to2.Port, suite.ch0to2.Channel, nextSeqSend+1)
	suite.NotNil(packetCommitment)

	suite.testPreparePacket(suite.app1, suite.ch1to0, suite.ch0to1, txID, 0, suite.chd1, makeTransactionInfo(tss[0]), nextSeqSend, types.PREPARE_RESULT_OK)
	suite.testPreparePacket(suite.app2, suite.ch2to0, suite.ch0to2, txID, 1, suite.chd2, makeTransactionInfo(tss[1]), nextSeqSend, types.PREPARE_RESULT_OK)
	suite.testPreparePacket(suite.app2, suite.ch2to0, suite.ch0to2, txID, 2, suite.chd2, makeTransactionInfo(tss[2]), nextSeqSend+1, types.PREPARE_RESULT_OK)

	// Tests for Confirm step

	nextSeqSend += 1

	// ensure that coordinator decides 'abort'
	{
		capp, _ := suite.app0.Cache()
		canMulticast, isCommitable, err := suite.testConfirmPrepareResult(&capp, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_FAILED), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
		suite.NoError(err)
		suite.True(canMulticast)
		suite.False(isCommitable)
	}
	// ensure that coordinator decides 'abort'
	{
		capp, _ := suite.app0.Cache()
		canMulticast, isCommitable, err := suite.testConfirmPrepareResult(&capp, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_OK), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)
		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_FAILED), txID, 1, suite.ch2to0, suite.ch0to2, nextSeqSend)
		suite.NoError(err)
		suite.True(canMulticast)
		suite.False(isCommitable)
	}
	// ensure that coordinator decides 'abort' (ordered sequence number)
	{
		capp, _ := suite.app0.Cache()
		canMulticast, isCommitable, err := suite.testConfirmPrepareResult(&capp, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_OK), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)
		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_OK), txID, 1, suite.ch2to0, suite.ch0to2, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)
		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_FAILED), txID, 2, suite.ch2to0, suite.ch0to2, nextSeqSend)
		suite.NoError(err)
		suite.True(canMulticast)
		suite.False(isCommitable)
	}
	// ensure that coordinator decides 'abort' (unordered sequence number)
	{
		capp, _ := suite.app0.Cache()
		canMulticast, isCommitable, err := suite.testConfirmPrepareResult(&capp, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_OK), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)
		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_OK), txID, 2, suite.ch2to0, suite.ch0to2, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)
		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_FAILED), txID, 1, suite.ch2to0, suite.ch0to2, nextSeqSend)
		suite.NoError(err)
		suite.True(canMulticast)
		suite.False(isCommitable)
	}
	// ensure that contractTransaction ID conflict occurs
	{
		capp, _ := suite.app0.Cache()
		canMulticast, isCommitable, err := suite.testConfirmPrepareResult(&capp, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_OK), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)
		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_OK), txID, 0, suite.ch2to0, suite.ch0to2, nextSeqSend)
		suite.Error(err)
	}
	// invalid transactionID
	{
		capp, _ := suite.app0.Cache()
		var invalidTxID types.TxID
		copy(invalidTxID[:], tmhash.Sum(txID[:]))
		_, _, err := suite.testConfirmPrepareResult(&capp, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_OK), invalidTxID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
		suite.Error(err)
	}
	// invalid transactionIndex
	{
		capp, _ := suite.app0.Cache()
		_, _, err := suite.testConfirmPrepareResult(&capp, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_OK), txID, 3, suite.ch1to0, suite.ch0to1, nextSeqSend)
		suite.Error(err)
	}
	// ensure that coordinator doesn't execute to multicast more than once
	{
		capp, _ := suite.app0.Cache()
		canMulticast, isCommitable, err := suite.testConfirmPrepareResult(&capp, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_OK), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)
		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_FAILED), txID, 1, suite.ch2to0, suite.ch0to2, nextSeqSend)
		suite.NoError(err)
		suite.True(canMulticast)
		suite.False(isCommitable)
		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_OK), txID, 2, suite.ch2to0, suite.ch0to2, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)
	}
	// ensure that coordinator doesn't receive a result of same contract call
	{
		capp, _ := suite.app0.Cache()
		canMulticast, isCommitable, err := suite.testConfirmPrepareResult(&capp, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_OK), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)
		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_OK), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
		suite.Error(err)
	}
	// ensure that coordinator decides 'commit' (unordered sequence number)
	{
		capp, _ := suite.app0.Cache()
		canMulticast, isCommitable, err := suite.testConfirmPrepareResult(&capp, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_OK), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)

		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_OK), txID, 2, suite.ch2to0, suite.ch0to2, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)

		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_OK), txID, 1, suite.ch2to0, suite.ch0to2, nextSeqSend)
		suite.NoError(err)
		suite.True(canMulticast)
		suite.True(isCommitable)
	}
	// ensure that coordinator decides 'commit' (ordered sequence number)
	{
		capp, writer := suite.app0.Cache()
		canMulticast, isCommitable, err := suite.testConfirmPrepareResult(&capp, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_OK), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)

		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_OK), txID, 1, suite.ch2to0, suite.ch0to2, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)

		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_OK), txID, 2, suite.ch2to0, suite.ch0to2, nextSeqSend)
		suite.NoError(err)
		suite.True(canMulticast)
		suite.True(isCommitable)

		writer()
	}

	// ensure that each participants execute to commit or abort
	{
		// In a1, execute to abort
		{
			capp, _ := suite.app1.Cache()
			suite.testAbortPacket(&capp, suite.chd1, suite.ch0to1, suite.ch1to0, tpctypes.NewPacketDataCommit(txID, 0, false), suite.signer1)
		}

		// In a2-0, execute to abort
		{
			capp, _ := suite.app2.Cache()
			suite.testAbortPacket(&capp, suite.chd2, suite.ch0to2, suite.ch2to0, tpctypes.NewPacketDataCommit(txID, 1, false), suite.signer2)
		}

		// In a2-1, execute to abort
		{
			capp, _ := suite.app2.Cache()
			suite.testAbortPacket(&capp, suite.chd2, suite.ch0to2, suite.ch2to0, tpctypes.NewPacketDataCommit(txID, 2, false), suite.signer3)
		}

		// In a1, execute to commit
		{
			capp, writer := suite.app1.Cache()
			suite.testCommitPacket(&capp, suite.chd1, suite.ch0to1, suite.ch1to0, tpctypes.NewPacketDataCommit(txID, 0, true), suite.signer1, func(res cross.ContractHandlerResult) {
				coin := sdk.NewCoin("tone", sdk.NewInt(80))
				expectedEvent := sdk.NewEvent("issue", sdk.NewAttribute("coin", coin.String()))
				suite.Equal(expectedEvent, res.GetEvents()[0])
				bz, err := coin.Marshal()
				if err != nil {
					suite.FailNow(err.Error())
				}
				suite.Equal(bz, res.GetData())
			})
			writer()
		}

		// In a2-0, execute to commit
		{
			capp, _ := suite.app2.Cache()
			suite.testCommitPacket(&capp, suite.chd2, suite.ch0to2, suite.ch2to0, tpctypes.NewPacketDataCommit(txID, 1, true), suite.signer2, func(res cross.ContractHandlerResult) {
				coin := sdk.NewCoin("ttwo", sdk.NewInt(60))
				expectedEvent := sdk.NewEvent("issue", sdk.NewAttribute("coin", coin.String()))
				suite.Equal(expectedEvent, res.GetEvents()[0])
				bz, err := coin.Marshal()
				if err != nil {
					suite.FailNow(err.Error())
				}
				suite.Equal(bz, res.GetData())
			})
		}

		// In a2-1, execute to commit
		{
			capp, writer := suite.app2.Cache()
			suite.testCommitPacket(&capp, suite.chd2, suite.ch0to2, suite.ch2to0, tpctypes.NewPacketDataCommit(txID, 2, true), suite.signer3, func(res cross.ContractHandlerResult) {
				coin := sdk.NewCoin("tthree", sdk.NewInt(40))
				expectedEvent := sdk.NewEvent("issue", sdk.NewAttribute("coin", coin.String()))
				suite.Equal(expectedEvent, res.GetEvents()[0])
				bz, err := coin.Marshal()
				if err != nil {
					suite.FailNow(err.Error())
				}
				suite.Equal(bz, res.GetData())
			})
			writer()
		}
	}
}

func (suite *TPCKeeperTestSuite) TestAbort1() {
	ci1 := contract.NewContractCallInfo("c1", "issue", [][]byte{[]byte("tone"), []byte("80")})
	ci2 := contract.NewContractCallInfo("c2", "must-fail", nil)

	var err error
	var nonce uint64 = 1
	var tss = []cross.ContractTransaction{
		cross.NewContractTransaction(
			suite.ch0to1,
			[]sdk.AccAddress{suite.signer1},
			ci1.Bytes(),
			cross.NewStateConstraint(
				cross.ExactMatchStateConstraint,
				[]cross.OP{
					lock.ReadOP{K: suite.signer1, V: nil},
					lock.WriteOP{K: suite.signer1, V: marshalCoins(sdk.Coins{sdk.NewInt64Coin("tone", 80)})},
				},
			),
			nil,
			nil,
		),
		cross.NewContractTransaction(
			suite.ch0to2,
			[]sdk.AccAddress{suite.signer2},
			ci2.Bytes(),
			cross.NewStateConstraint(
				cross.ExactMatchStateConstraint,
				[]cross.OP{},
			),
			nil,
			nil,
		),
	}

	suite.openAllChannels()
	msg := cross.NewMsgInitiate(
		suite.initiator,
		suite.app0.chainID,
		tss,
		256,
		nonce,
		cross.COMMIT_PROTOCOL_TPC,
	)
	txID, err := suite.app0.app.CrossKeeper.TPCKeeper().MulticastPreparePacket(
		suite.app0.ctx,
		types.NewSimplePacketSender(suite.app0.app.IBCKeeper.ChannelKeeper),
		suite.initiator,
		msg,
		msg.ContractTransactions,
	)
	suite.NoError(err)

	var nextSeqSend uint64 = 1
	suite.testPreparePacket(suite.app1, suite.ch1to0, suite.ch0to1, txID, 0, suite.chd1, makeTransactionInfo(tss[0]), nextSeqSend, types.PREPARE_RESULT_OK)
	suite.testPreparePacket(suite.app2, suite.ch2to0, suite.ch0to2, txID, 1, suite.chd2, makeTransactionInfo(tss[1]), nextSeqSend, types.PREPARE_RESULT_FAILED)

	nextSeqSend += 1

	canMulticast, isCommitable, err := suite.testConfirmPrepareResult(suite.app0, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_OK), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
	suite.NoError(err)
	suite.False(canMulticast)
	suite.False(isCommitable)

	canMulticast, isCommitable, err = suite.testConfirmPrepareResult(suite.app0, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_FAILED), txID, 1, suite.ch2to0, suite.ch0to2, nextSeqSend)
	suite.NoError(err)
	suite.True(canMulticast)
	suite.False(isCommitable)

	// In a1, execute to abort
	suite.testAbortPacket(suite.app1, suite.chd1, suite.ch0to1, suite.ch1to0, tpctypes.NewPacketDataCommit(txID, 0, false), suite.signer1)

	// In a2, execute to abort
	suite.testAbortPacket(suite.app2, suite.chd2, suite.ch0to2, suite.ch2to0, tpctypes.NewPacketDataCommit(txID, 1, false), suite.signer2)
}

func (suite *TPCKeeperTestSuite) TestAbort2() {
	ci1 := contract.NewContractCallInfo("c1", "must-fail", nil)
	ci2 := contract.NewContractCallInfo("c2", "must-fail", nil)

	var err error
	var nonce uint64 = 1
	var tss = []cross.ContractTransaction{
		cross.NewContractTransaction(
			suite.ch0to1,
			[]sdk.AccAddress{suite.signer1},
			ci1.Bytes(),
			cross.NewStateConstraint(
				cross.ExactMatchStateConstraint,
				[]cross.OP{},
			),
			nil,
			nil,
		),
		cross.NewContractTransaction(
			suite.ch0to2,
			[]sdk.AccAddress{suite.signer2},
			ci2.Bytes(),
			cross.NewStateConstraint(
				cross.ExactMatchStateConstraint,
				[]cross.OP{},
			),
			nil,
			nil,
		),
	}

	suite.openAllChannels()
	msg := cross.NewMsgInitiate(
		suite.initiator,
		suite.app0.chainID,
		tss,
		256,
		nonce,
		cross.COMMIT_PROTOCOL_TPC,
	)
	txID, err := suite.app0.app.CrossKeeper.TPCKeeper().MulticastPreparePacket(
		suite.app0.ctx,
		types.NewSimplePacketSender(suite.app0.app.IBCKeeper.ChannelKeeper),
		suite.initiator,
		msg,
		msg.ContractTransactions,
	)
	suite.NoError(err)

	var nextSeqSend uint64 = 1
	suite.testPreparePacket(suite.app1, suite.ch1to0, suite.ch0to1, txID, 0, suite.chd1, makeTransactionInfo(tss[0]), nextSeqSend, types.PREPARE_RESULT_FAILED)
	suite.testPreparePacket(suite.app2, suite.ch2to0, suite.ch0to2, txID, 1, suite.chd2, makeTransactionInfo(tss[1]), nextSeqSend, types.PREPARE_RESULT_FAILED)

	nextSeqSend += 1

	canMulticast, isCommitable, err := suite.testConfirmPrepareResult(suite.app0, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_FAILED), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
	suite.NoError(err)
	suite.True(canMulticast)
	suite.False(isCommitable)

	canMulticast, isCommitable, err = suite.testConfirmPrepareResult(suite.app0, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_FAILED), txID, 1, suite.ch2to0, suite.ch0to2, nextSeqSend)
	suite.NoError(err)
	suite.False(canMulticast)
	suite.False(isCommitable)

	// In a1, execute to abort
	suite.testAbortPacket(suite.app1, suite.chd1, suite.ch0to1, suite.ch1to0, tpctypes.NewPacketDataCommit(txID, 0, false), suite.signer1)

	// In a2, execute to abort
	suite.testAbortPacket(suite.app2, suite.chd2, suite.ch0to2, suite.ch2to0, tpctypes.NewPacketDataCommit(txID, 1, false), suite.signer2)
}

func (suite *TPCKeeperTestSuite) TestAbort3() {
	ci1 := contract.NewContractCallInfo("c1", "issue", [][]byte{[]byte("tone"), []byte("80")})
	ci2 := contract.NewContractCallInfo("c2", "issue", [][]byte{[]byte("ttwo"), []byte("60")})

	var err error
	var nonce uint64 = 1
	var tss = []cross.ContractTransaction{
		cross.NewContractTransaction(
			suite.ch0to1,
			[]sdk.AccAddress{suite.signer1},
			ci1.Bytes(),
			cross.NewStateConstraint(
				cross.ExactMatchStateConstraint,
				[]cross.OP{
					lock.ReadOP{K: suite.signer1, V: nil},
					lock.WriteOP{K: suite.signer1, V: marshalCoins(sdk.Coins{sdk.NewInt64Coin("tone", 80)})},
				},
			),
			cross.NewReturnValue(marshalCoin(sdk.NewInt64Coin("tone", 80))),
			nil,
		),
		cross.NewContractTransaction(
			suite.ch0to2,
			[]sdk.AccAddress{suite.signer2},
			ci2.Bytes(),
			cross.NewStateConstraint(
				cross.ExactMatchStateConstraint,
				[]cross.OP{
					lock.ReadOP{K: suite.signer2, V: nil},
					lock.WriteOP{K: suite.signer2, V: marshalCoins(sdk.Coins{sdk.NewInt64Coin("ttwo", 100)})}, // invalid OP
				},
			),
			cross.NewReturnValue(marshalCoin(sdk.NewInt64Coin("ttwo", 60))),
			nil,
		),
	}

	suite.openAllChannels()
	msg := cross.NewMsgInitiate(
		suite.initiator,
		suite.app0.chainID,
		tss,
		256,
		nonce,
		cross.COMMIT_PROTOCOL_TPC,
	)
	txID, err := suite.app0.app.CrossKeeper.TPCKeeper().MulticastPreparePacket(
		suite.app0.ctx,
		types.NewSimplePacketSender(suite.app0.app.IBCKeeper.ChannelKeeper),
		suite.initiator,
		msg,
		msg.ContractTransactions,
	)
	suite.NoError(err)

	var nextSeqSend uint64 = 1
	suite.testPreparePacket(suite.app1, suite.ch1to0, suite.ch0to1, txID, 0, suite.chd1, makeTransactionInfo(tss[0]), nextSeqSend, types.PREPARE_RESULT_OK)
	suite.testPreparePacket(suite.app2, suite.ch2to0, suite.ch0to2, txID, 1, suite.chd2, makeTransactionInfo(tss[1]), nextSeqSend, types.PREPARE_RESULT_FAILED)

	nextSeqSend += 1

	canMulticast, isCommitable, err := suite.testConfirmPrepareResult(suite.app0, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_OK), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
	suite.NoError(err)
	suite.False(canMulticast)
	suite.False(isCommitable)

	canMulticast, isCommitable, err = suite.testConfirmPrepareResult(suite.app0, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_FAILED), txID, 1, suite.ch2to0, suite.ch0to2, nextSeqSend)
	suite.NoError(err)
	suite.True(canMulticast)
	suite.False(isCommitable)

	// In a1, execute to abort
	suite.testAbortPacket(suite.app1, suite.chd1, suite.ch0to1, suite.ch1to0, tpctypes.NewPacketDataCommit(txID, 0, false), suite.signer1)

	// In a2, execute to abort
	suite.testAbortPacket(suite.app2, suite.chd2, suite.ch0to2, suite.ch2to0, tpctypes.NewPacketDataCommit(txID, 1, false), suite.signer2)
}

func (suite *TPCKeeperTestSuite) TestStateConstraint() {
	ci1 := contract.NewContractCallInfo("c1", "issue", [][]byte{[]byte("tone"), []byte("80")})
	// app2 has multiple contract calls
	ci2 := contract.NewContractCallInfo("c2", "issue", [][]byte{[]byte("ttwo"), []byte("60")})
	ci3 := contract.NewContractCallInfo("c2", "issue", [][]byte{[]byte("tthree"), []byte("40")})

	var err error
	var nonce uint64 = 1
	var tss = []cross.ContractTransaction{
		cross.NewContractTransaction(
			suite.ch0to1,
			[]sdk.AccAddress{suite.signer1},
			ci1.Bytes(),
			cross.NewStateConstraint(
				cross.PreStateConstraint,
				[]cross.OP{
					lock.ReadOP{K: suite.signer1, V: nil},
				},
			),
			nil,
			nil,
		),
		cross.NewContractTransaction(
			suite.ch0to2,
			[]sdk.AccAddress{suite.signer2},
			ci2.Bytes(),
			cross.NewStateConstraint(
				cross.PostStateConstraint,
				[]cross.OP{
					lock.WriteOP{K: suite.signer2, V: marshalCoins(sdk.Coins{sdk.NewInt64Coin("ttwo", 60)})},
				},
			),
			nil,
			nil,
		),
		cross.NewContractTransaction(
			suite.ch0to2,
			[]sdk.AccAddress{suite.signer3},
			ci3.Bytes(),
			cross.NewStateConstraint(
				cross.NoStateConstraint,
				[]cross.OP{},
			),
			nil,
			nil,
		),
	}

	suite.openAllChannels()
	msg := cross.NewMsgInitiate(
		suite.initiator,
		suite.app0.chainID,
		tss,
		256,
		nonce,
		cross.COMMIT_PROTOCOL_TPC,
	)
	txID, err := suite.app0.app.CrossKeeper.TPCKeeper().MulticastPreparePacket(
		suite.app0.ctx,
		types.NewSimplePacketSender(suite.app0.app.IBCKeeper.ChannelKeeper),
		suite.initiator,
		msg,
		msg.ContractTransactions,
	)
	suite.NoError(err)

	var nextSeqSend uint64 = 1
	suite.testPreparePacket(suite.app1, suite.ch1to0, suite.ch0to1, txID, 0, suite.chd1, makeTransactionInfo(tss[0]), nextSeqSend, types.PREPARE_RESULT_OK)
	suite.testPreparePacket(suite.app2, suite.ch2to0, suite.ch0to2, txID, 1, suite.chd2, makeTransactionInfo(tss[1]), nextSeqSend, types.PREPARE_RESULT_OK)
	suite.testPreparePacket(suite.app2, suite.ch2to0, suite.ch0to2, txID, 2, suite.chd2, makeTransactionInfo(tss[2]), nextSeqSend+1, types.PREPARE_RESULT_OK)

	nextSeqSend += 1

	canMulticast, isCommitable, err := suite.testConfirmPrepareResult(suite.app0, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_OK), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
	suite.NoError(err)
	suite.False(canMulticast)
	suite.False(isCommitable)

	canMulticast, isCommitable, err = suite.testConfirmPrepareResult(suite.app0, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_OK), txID, 1, suite.ch2to0, suite.ch0to2, nextSeqSend)
	suite.NoError(err)
	suite.False(canMulticast)
	suite.False(isCommitable)

	canMulticast, isCommitable, err = suite.testConfirmPrepareResult(suite.app0, tpctypes.NewPacketPrepareAcknowledgement(types.PREPARE_RESULT_OK), txID, 2, suite.ch2to0, suite.ch0to2, nextSeqSend)
	suite.NoError(err)
	suite.True(canMulticast)
	suite.True(isCommitable)

	// In a1, execute to commit
	{
		capp, writer := suite.app1.Cache()
		suite.testCommitPacket(&capp, suite.chd1, suite.ch0to1, suite.ch1to0, tpctypes.NewPacketDataCommit(txID, 0, true), suite.signer1, func(res cross.ContractHandlerResult) {
			coin := sdk.NewCoin("tone", sdk.NewInt(80))
			expectedEvent := sdk.NewEvent("issue", sdk.NewAttribute("coin", coin.String()))
			suite.Equal(expectedEvent, res.GetEvents()[0])
			bz, err := coin.Marshal()
			if err != nil {
				suite.FailNow(err.Error())
			}
			suite.Equal(bz, res.GetData())
		})
		writer()
	}

	// In a2-0, execute to commit
	{
		capp, _ := suite.app2.Cache()
		suite.testCommitPacket(&capp, suite.chd2, suite.ch0to2, suite.ch2to0, tpctypes.NewPacketDataCommit(txID, 1, true), suite.signer2, func(res cross.ContractHandlerResult) {
			coin := sdk.NewCoin("ttwo", sdk.NewInt(60))
			expectedEvent := sdk.NewEvent("issue", sdk.NewAttribute("coin", coin.String()))
			suite.Equal(expectedEvent, res.GetEvents()[0])
			bz, err := coin.Marshal()
			if err != nil {
				suite.FailNow(err.Error())
			}
			suite.Equal(bz, res.GetData())
		})
	}

	// In a2-1, execute to commit
	{
		capp, writer := suite.app2.Cache()
		suite.testCommitPacket(&capp, suite.chd2, suite.ch0to2, suite.ch2to0, tpctypes.NewPacketDataCommit(txID, 2, true), suite.signer3, func(res cross.ContractHandlerResult) {
			coin := sdk.NewCoin("tthree", sdk.NewInt(40))
			expectedEvent := sdk.NewEvent("issue", sdk.NewAttribute("coin", coin.String()))
			suite.Equal(expectedEvent, res.GetEvents()[0])
			bz, err := coin.Marshal()
			if err != nil {
				suite.FailNow(err.Error())
			}
			suite.Equal(bz, res.GetData())
		})
		writer()
	}
}

/*
TrustedChannelInfoResolver just returns a given ChannelInfo as is.
CAUTION: This assumes that the coordinator and participant have the same channel ID and port ID for a chain.
*/
type TrustedChannelInfoResolver struct {
	types.ChannelInfoResolver
}

// Capabilities implements ChannelResolver.Capabilities
func (r TrustedChannelInfoResolver) Capabilities() types.ChannelResolverCapabilities {
	return channelResolverCapabilities{crossChainCalls: true}
}

type channelResolverCapabilities struct {
	crossChainCalls bool
}

func (c channelResolverCapabilities) CrossChainCalls() bool {
	return c.crossChainCalls
}

func (suite *TPCKeeperTestSuite) TestCrossChainCall() {
	suite.setup(TrustedChannelInfoResolver{})

	// First, issue some token to signer1
	store, _, err := suite.chd2.Handle(
		cross.WithSigners(suite.app2.ctx, []sdk.AccAddress{suite.signer1}),
		contract.NewContractCallInfo("c2", "issue", [][]byte{[]byte("tone"), []byte("100")}).Bytes(),
		cross.ContractRuntimeInfo{StateConstraintType: cross.NoStateConstraint},
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	if err := store.CommitImmediately(); err != nil {
		suite.FailNow(err.Error())
	}

	ci1 := contract.NewContractCallInfo("c1", "peg-coin", [][]byte{[]byte("tone"), []byte("100")})
	ci2 := contract.NewContractCallInfo("c2", "lock-coin", [][]byte{[]byte("tone"), []byte("100")})

	var nonce uint64 = 1
	var tss = []cross.ContractTransaction{
		cross.NewContractTransaction(
			suite.ch0to1,
			[]sdk.AccAddress{suite.signer1},
			ci1.Bytes(),
			cross.NewStateConstraint(
				cross.NoStateConstraint,
				[]cross.OP{},
			),
			nil,
			[]cross.Link{
				cross.NewCallResultLink(1),
			},
		),
		cross.NewContractTransaction(
			suite.ch0to2,
			[]sdk.AccAddress{suite.signer1},
			ci2.Bytes(),
			cross.NewStateConstraint(
				cross.NoStateConstraint,
				[]cross.OP{},
			),
			cross.NewReturnValue(marshalCoins(sdk.NewCoins(sdk.NewInt64Coin("tone", 100)))),
			nil,
		),
	}

	suite.openAllChannels()
	msg := cross.NewMsgInitiate(
		suite.initiator,
		suite.app0.chainID,
		tss,
		256,
		nonce,
		cross.COMMIT_PROTOCOL_TPC,
	)
	txID, err := suite.app0.app.CrossKeeper.TPCKeeper().MulticastPreparePacket(
		suite.app0.ctx,
		types.NewSimplePacketSender(suite.app0.app.IBCKeeper.ChannelKeeper),
		suite.initiator,
		msg,
		msg.ContractTransactions,
	)
	suite.NoError(err)

	lkr, err := types.MakeLinker(tss)
	if err != nil {
		suite.FailNow(err.Error())
	}
	objs0, err := lkr.Resolve(tss[0].Links)
	if err != nil {
		suite.FailNow(err.Error())
	}
	objs1, err := lkr.Resolve(tss[1].Links)
	if err != nil {
		suite.FailNow(err.Error())
	}

	var nextSeqSend uint64 = 1
	suite.testPreparePacket(suite.app1, suite.ch1to0, suite.ch0to1, txID, 0, suite.chd1, makeTransactionInfo(tss[0], objs0...), nextSeqSend, types.PREPARE_RESULT_OK)
	suite.testPreparePacket(suite.app2, suite.ch2to0, suite.ch0to2, txID, 1, suite.chd2, makeTransactionInfo(tss[1], objs1...), nextSeqSend, types.PREPARE_RESULT_OK)
}

func (suite *TPCKeeperTestSuite) testPreparePacket(actx *appContext, src, dst cross.ChannelInfo, txID types.TxID, txIndex types.TxIndex, contractHandler cross.ContractHandler, txInfo cross.ContractTransactionInfo, nextseq uint64, expectedPrepareResult uint8) {
	relayer := sdk.AccAddress("relayer1")
	packetData := tpctypes.NewPacketDataPrepare(relayer, txID, txIndex, txInfo)
	ctx, writer := actx.ctx.CacheContext()
	ctx = cross.WithSigners(ctx, txInfo.Transaction.Signers)
	result, err := actx.app.CrossKeeper.TPCKeeper().Prepare(
		ctx,
		contractHandler,
		src.Port,
		src.Channel,
		packetData,
	)
	suite.NoError(err)
	suite.Equal(expectedPrepareResult, result)
	tx, ok := actx.app.CrossKeeper.TPCKeeper().GetTx(ctx, txID, txIndex)
	suite.Require().True(ok)
	suite.Equal(cross.TX_STATUS_PREPARE, tx.Status)
	suite.Equal(expectedPrepareResult, tx.PrepareResult)
	writer()
}

func (suite *TPCKeeperTestSuite) testConfirmPrepareResult(actx *appContext, ack tpctypes.PacketPrepareAcknowledgement, txID cross.TxID, txIndex cross.TxIndex, src, dst cross.ChannelInfo, nextseq uint64) (bool, bool, error) {
	canMulticast, isCommitable, err := actx.app.CrossKeeper.TPCKeeper().ReceivePrepareAcknowledgement(actx.ctx, dst.Port, dst.Channel, ack, txID, txIndex)
	if err != nil {
		return false, false, err
	}
	if canMulticast {
		return canMulticast, isCommitable, actx.app.CrossKeeper.TPCKeeper().MulticastCommitPacket(actx.ctx, types.NewSimplePacketSender(actx.app.IBCKeeper.ChannelKeeper), txID, isCommitable)
	} else {
		return canMulticast, isCommitable, nil
	}
}

func (suite *TPCKeeperTestSuite) testAbortPacket(actx *appContext, contractHandler cross.ContractHandler, src, dst cross.ChannelInfo, packet tpctypes.PacketDataCommit, txSigner sdk.AccAddress) {
	_, err := actx.app.CrossKeeper.TPCKeeper().ReceiveCommitPacket(actx.ctx, contractHandler, src.Port, src.Channel, dst.Port, dst.Channel, packet)
	if !suite.NoError(err) {
		return
	}
	tx, found := actx.app.CrossKeeper.TPCKeeper().GetTx(actx.ctx, packet.TxID, packet.TxIndex)
	if !suite.True(found) {
		return
	}
	suite.Equal(cross.TX_STATUS_ABORT, tx.Status)
	// ensure that the state is expected
	_, err = contractHandler.GetState(actx.ctx, tx.ContractCallInfo, types.ContractRuntimeInfo{StateConstraintType: cross.NoStateConstraint})
	if !suite.NoError(err) {
		return
	}
	ci, err := contract.DecodeContractCallInfo(tx.ContractCallInfo)
	if !suite.NoError(err) {
		return
	}
	contractInfo := contract.NewContractCallInfo(ci.ID, "test-not-issued", [][]byte{})
	bz, err := contract.EncodeContractCallInfo(contractInfo)
	if !suite.NoError(err) {
		return
	}
	actx2, _ := actx.Cache()
	ctx := cross.WithSigners(actx2.ctx, []sdk.AccAddress{txSigner})
	_, _, err = contractHandler.Handle(ctx, bz, types.ContractRuntimeInfo{StateConstraintType: cross.ExactMatchStateConstraint})
	suite.NoError(err)
}

func (suite *TPCKeeperTestSuite) testCommitPacket(actx *appContext, contractHandler cross.ContractHandler, src, dst cross.ChannelInfo, packet tpctypes.PacketDataCommit, txSigner sdk.AccAddress, checkResult func(cross.ContractHandlerResult)) {
	res, err := actx.app.CrossKeeper.TPCKeeper().ReceiveCommitPacket(actx.ctx, contractHandler, src.Port, src.Channel, dst.Port, dst.Channel, packet)
	if !suite.NoError(err) {
		return
	}
	checkResult(res)
	tx, found := actx.app.CrossKeeper.TPCKeeper().GetTx(actx.ctx, packet.TxID, packet.TxIndex)
	if !suite.True(found) {
		return
	}
	suite.Equal(cross.TX_STATUS_COMMIT, tx.Status)
	// ensure that the state is expected
	_, err = contractHandler.GetState(actx.ctx, tx.ContractCallInfo, types.ContractRuntimeInfo{StateConstraintType: cross.ExactMatchStateConstraint})
	if !suite.NoError(err) {
		return
	}
	ci, err := contract.DecodeContractCallInfo(tx.ContractCallInfo)
	if !suite.NoError(err) {
		return
	}
	contractInfo := contract.NewContractCallInfo(ci.ID, "test-balance", [][]byte{
		ci.Args[0],
		ci.Args[1],
	})
	bz, err := contract.EncodeContractCallInfo(contractInfo)
	if !suite.NoError(err) {
		return
	}
	ctx := cross.WithSigners(actx.ctx, []sdk.AccAddress{txSigner})
	_, _, err = contractHandler.Handle(ctx, bz, types.ContractRuntimeInfo{StateConstraintType: cross.ExactMatchStateConstraint})
	suite.NoError(err)
}
