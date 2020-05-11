package keeper_test

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/datachainlab/cross/x/ibc/contract"
	"github.com/datachainlab/cross/x/ibc/cross"
	"github.com/datachainlab/cross/x/ibc/cross/types"
	lock "github.com/datachainlab/cross/x/ibc/store/lock"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

// define constants used for testing
const (
	testClientType     = clientexported.Tendermint
	testChannelOrder   = channelexported.UNORDERED
	testChannelVersion = "1.0"
)

const (
	trustingPeriod time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod      time.Duration = time.Hour * 24 * 7 * 3
	maxClockDrift  time.Duration = time.Second * 10
)

func (suite *KeeperTestSuite) createClient(actx *appContext, clientID string) {
	actx.app.Commit()

	h := abci.Header{ChainID: actx.ctx.ChainID(), Height: actx.app.LastBlockHeight() + 1}
	actx.app.BeginBlock(abci.RequestBeginBlock{Header: h})
	actx.ctx = actx.ctx.WithBlockHeader(h)
	now := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)

	header := tendermint.CreateTestHeader(actx.chainID, 1, now, actx.valSet, actx.signers)
	consensusState := header.ConsensusState()

	clientState, err := tendermint.Initialize(clientID, trustingPeriod, ubdPeriod, maxClockDrift, header)
	if err != nil {
		panic(err)
	}

	_, err = actx.app.IBCKeeper.ClientKeeper.CreateClient(actx.ctx, clientState, consensusState)
	suite.NoError(err)
}

func (suite *KeeperTestSuite) updateClient(actx *appContext, clientID string) {
	// always commit and begin a new block on updateClient
	actx.app.Commit()
	commitID := actx.app.LastCommitID()

	h := abci.Header{ChainID: actx.ctx.ChainID(), Height: actx.app.LastBlockHeight() + 1}
	actx.app.BeginBlock(abci.RequestBeginBlock{Header: h})
	actx.ctx = actx.ctx.WithBlockHeader(h)

	state := tendermint.ConsensusState{
		Root: commitment.NewMerkleRoot(commitID.Hash),
	}

	actx.app.IBCKeeper.ClientKeeper.SetClientConsensusState(actx.ctx, clientID, 1, state)
}

func (suite *KeeperTestSuite) createConnection(actx *appContext, clientID, connectionID, counterpartyClientID, counterpartyConnectionID string, state connectionexported.State) {
	connection := connection.ConnectionEnd{
		State:    state,
		ClientID: clientID,
		Counterparty: connection.Counterparty{
			ClientID:     counterpartyClientID,
			ConnectionID: counterpartyConnectionID,
			Prefix:       actx.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(),
		},
		Versions: connection.GetCompatibleVersions(),
	}

	actx.app.IBCKeeper.ConnectionKeeper.SetConnection(actx.ctx, connectionID, connection)
}

func (suite *KeeperTestSuite) createChannel(actx *appContext, portID string, chanID string, connID string, counterpartyPort string, counterpartyChan string, state channelexported.State) {
	ch := channel.Channel{
		State:    state,
		Ordering: testChannelOrder,
		Counterparty: channel.Counterparty{
			PortID:    counterpartyPort,
			ChannelID: counterpartyChan,
		},
		ConnectionHops: []string{connID},
		Version:        testChannelVersion,
	}

	actx.app.IBCKeeper.ChannelKeeper.SetChannel(actx.ctx, portID, chanID, ch)
	capName := ibctypes.ChannelCapabilityPath(portID, chanID)
	cap, err := actx.app.ScopedIBCKeeper.NewCapability(actx.ctx, capName)
	if err != nil {
		suite.FailNow(err.Error())
	}
	if err := actx.app.CrossKeeper.ClaimCapability(actx.ctx, cap, capName); err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *KeeperTestSuite) queryProof(actx *appContext, key []byte) (proof commitmentexported.Proof, height int64) {
	res := actx.app.Query(abci.RequestQuery{
		Path:  fmt.Sprintf("store/%s/key", ibctypes.StoreKey),
		Data:  key,
		Prove: true,
	})

	height = res.Height
	proof = commitment.MerkleProof{
		Proof: res.Proof,
	}

	return
}

func (suite *KeeperTestSuite) createContractHandler(cdc *codec.Codec, stk sdk.StoreKey, cid string) cross.ContractHandler {
	contractHandler := contract.NewContractHandler(contract.NewKeeper(cdc, stk), func(kvs sdk.KVStore, tp cross.StateConditionType) cross.State {
		return lock.NewStore(kvs, cross.ExactStateCondition)
	})
	c := contract.NewContract([]contract.Method{
		{
			Name: "issue",
			F: func(ctx contract.Context, store cross.Store) ([]byte, error) {
				coin, err := parseCoin(ctx, 0, 1)
				if err != nil {
					return nil, err
				}
				balance := getBalanceOf(store, ctx.Signers()[0])
				balance = balance.Add(coin)
				setBalance(store, ctx.Signers()[0], balance)
				ctx.EventManager().EmitEvent(
					sdk.NewEvent("issue", sdk.NewAttribute("coin", coin.String())),
				)
				return coin.Marshal()
			},
		},
		{
			Name: "must-fail",
			F: func(ctx contract.Context, store cross.Store) ([]byte, error) {
				return nil, errors.New("failed to call contract")
			},
		},
		{
			Name: "test-balance",
			F: func(ctx contract.Context, store cross.Store) ([]byte, error) {
				coin, err := parseCoin(ctx, 0, 1)
				if err != nil {
					return nil, err
				}
				balance := getBalanceOf(store, ctx.Signers()[0])
				if !balance.AmountOf(coin.Denom).Equal(coin.Amount) {
					return nil, errors.New("amount is unexpected")
				}
				return nil, nil
			},
		},
		{
			Name: "test-not-issued",
			F: func(ctx contract.Context, store cross.Store) ([]byte, error) {
				balance := getBalanceOf(store, ctx.Signers()[0])
				if len(balance) == 0 {
					return nil, nil
				} else {
					return nil, errors.New("maybe coin is already issued")
				}
			},
		},
	})
	contractHandler.AddRoute(cid, c)
	return contractHandler
}

func (suite *KeeperTestSuite) TestInitiateMsg() {
	initiator := sdk.AccAddress("initiator")
	app0 := suite.createApp("app0") // coordinator node
	app1 := suite.createApp("app1")
	signer1 := sdk.AccAddress("signer1")
	ci1 := contract.NewContractCallInfo("c1", "issue", [][]byte{[]byte("tone"), []byte("80")})

	app2 := suite.createApp("app2")
	signer2 := sdk.AccAddress("signer2")
	ci2 := contract.NewContractCallInfo("c2", "issue", [][]byte{[]byte("ttwo"), []byte("60")})

	ch0to1 := cross.NewChannelInfo("testportzeroone", "testchannelzeroone") // app0 -> app1
	ch1to0 := cross.NewChannelInfo("testportonezero", "testchannelonezero") // app1 -> app0
	ch0to2 := cross.NewChannelInfo("testportzerotwo", "testchannelzerotwo") // app0 -> app2
	ch2to0 := cross.NewChannelInfo("testporttwozero", "testchanneltwozero") // app2 -> app0

	var nonce uint64 = 1
	var tss = []cross.ContractTransaction{
		cross.NewContractTransaction(
			ch0to1,
			[]sdk.AccAddress{signer1},
			ci1.Bytes(),
			cross.NewStateCondition(
				cross.ExactStateCondition,
				[]cross.OP{lock.WriteOP{K: signer1, V: marshalCoin(sdk.Coins{sdk.NewInt64Coin("tone", 80)})}},
			),
		),
		cross.NewContractTransaction(
			ch0to2,
			[]sdk.AccAddress{signer2},
			ci2.Bytes(),
			cross.NewStateCondition(
				cross.ExactStateCondition,
				[]cross.OP{lock.WriteOP{K: signer2, V: marshalCoin(sdk.Coins{sdk.NewInt64Coin("ttwo", 60)})}},
			),
		),
	}

	{
		msg := cross.NewMsgInitiate(
			initiator,
			app0.chainID,
			tss,
			5,
			nonce,
		)
		_, err := app0.app.CrossKeeper.MulticastPreparePacket(
			app0.ctx,
			initiator,
			msg,
			msg.ContractTransactions,
		)
		suite.Error(err) // channel does not exist
	}

	// Try to open a channel and connection between app0 and app1, app2

	suite.openChannels(
		app1.chainID,
		app0.chainID+app1.chainID,
		ch0to1,
		app0,

		app0.chainID,
		app1.chainID+app0.chainID,
		ch1to0,
		app1,
	)

	suite.openChannels(
		app2.chainID,
		app0.chainID+app2.chainID,
		ch0to2,
		app0,

		app0.chainID,
		app2.chainID+app1.chainID,
		ch2to0,
		app2,
	)

	// ensure that current block height is correct
	suite.EqualValues(3, app0.ctx.BlockHeight())

	{
		msg := cross.NewMsgInitiate(
			initiator,
			app0.chainID,
			tss,
			3,
			nonce,
		)
		_, err := app0.app.CrossKeeper.MulticastPreparePacket(
			app0.ctx,
			initiator,
			msg,
			msg.ContractTransactions,
		)
		suite.Error(err) // timeout error
	}

	{
		msg := cross.NewMsgInitiate(
			initiator,
			"dummy", // invalid chainID
			tss,
			4,
			nonce,
		)
		_, err := app0.app.CrossKeeper.MulticastPreparePacket(
			app0.ctx,
			initiator,
			msg,
			msg.ContractTransactions,
		)
		suite.Error(err) // occur an error due to invalid chainID
	}

	{
		msg := cross.NewMsgInitiate(
			initiator,
			app0.chainID,
			tss,
			4,
			nonce,
		)
		_, err := app0.app.CrossKeeper.MulticastPreparePacket(
			app0.ctx,
			initiator,
			msg,
			msg.ContractTransactions,
		)
		suite.NoError(err) // successfully executed
	}
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.initiator = sdk.AccAddress("initiator")
	suite.signer1 = sdk.AccAddress("signer1")

	suite.app0 = suite.createApp("app0") // coordinator node
	suite.app1 = suite.createApp("app1")

	suite.chd1 = suite.createContractHandler(suite.app1.cdc, suite.app1.app.GetKey(cross.StoreKey), "c1")

	suite.app2 = suite.createApp("app2")
	suite.signer2 = sdk.AccAddress("signer2")
	suite.signer3 = sdk.AccAddress("signer3")

	suite.chd2 = suite.createContractHandler(suite.app2.cdc, suite.app2.app.GetKey(cross.StoreKey), "c2")

	suite.ch0to1 = cross.NewChannelInfo("testportzeroone", "testchannelzeroone") // app0 -> app1
	suite.ch1to0 = cross.NewChannelInfo("testportonezero", "testchannelonezero") // app1 -> app0
	suite.ch0to2 = cross.NewChannelInfo("testportzerotwo", "testchannelzerotwo") // app0 -> app2
	suite.ch2to0 = cross.NewChannelInfo("testporttwozero", "testchanneltwozero") // app2 -> app0
}

func (suite *KeeperTestSuite) openAllChannels() {
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

func (suite *KeeperTestSuite) TestRelay() {
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
			cross.NewStateCondition(
				cross.ExactStateCondition,
				[]cross.OP{
					lock.ReadValueOP{K: suite.signer1, V: nil},
					lock.WriteOP{K: suite.signer1, V: marshalCoin(sdk.Coins{sdk.NewInt64Coin("tone", 80)})},
				},
			),
		),
		cross.NewContractTransaction(
			suite.ch0to2,
			[]sdk.AccAddress{suite.signer2},
			ci2.Bytes(),
			cross.NewStateCondition(
				cross.ExactStateCondition,
				[]cross.OP{
					lock.ReadValueOP{K: suite.signer2, V: nil},
					lock.WriteOP{K: suite.signer2, V: marshalCoin(sdk.Coins{sdk.NewInt64Coin("ttwo", 60)})},
				},
			),
		),
		cross.NewContractTransaction(
			suite.ch0to2,
			[]sdk.AccAddress{suite.signer3},
			ci3.Bytes(),
			cross.NewStateCondition(
				cross.ExactStateCondition,
				[]cross.OP{
					lock.ReadValueOP{K: suite.signer3, V: nil},
					lock.WriteOP{K: suite.signer3, V: marshalCoin(sdk.Coins{sdk.NewInt64Coin("tthree", 40)})},
				},
			),
		),
	}

	msg := cross.NewMsgInitiate(
		suite.initiator,
		suite.app0.chainID,
		tss,
		256,
		nonce,
	)
	_, err = suite.app0.app.CrossKeeper.MulticastPreparePacket(
		suite.app0.ctx,
		suite.initiator,
		msg,
		msg.ContractTransactions,
	)
	suite.Error(err) // channel does not exist

	// Try to open a channel and connection between app0 and app1, app2
	suite.openAllChannels()

	txID, err := suite.app0.app.CrossKeeper.MulticastPreparePacket(
		suite.app0.ctx,
		suite.initiator,
		msg,
		msg.ContractTransactions,
	)
	suite.NoError(err) // successfully executed

	ci, found := suite.app0.app.CrossKeeper.GetCoordinator(suite.app0.ctx, txID)
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

	suite.testPreparePacket(suite.app1, suite.ch1to0, suite.ch0to1, txID, 0, suite.chd1, tss[0], nextSeqSend, cross.PREPARE_RESULT_OK)
	suite.testPreparePacket(suite.app2, suite.ch2to0, suite.ch0to2, txID, 1, suite.chd2, tss[1], nextSeqSend, cross.PREPARE_RESULT_OK)
	suite.testPreparePacket(suite.app2, suite.ch2to0, suite.ch0to2, txID, 2, suite.chd2, tss[2], nextSeqSend+1, cross.PREPARE_RESULT_OK)

	// Tests for Confirm step

	nextSeqSend += 1

	// ensure that coordinator decides 'abort'
	{
		capp, _ := suite.app0.Cache()
		canMulticast, isCommitable, err := suite.testConfirmPrepareResult(&capp, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_FAILED), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
		suite.NoError(err)
		suite.True(canMulticast)
		suite.False(isCommitable)
	}
	// ensure that coordinator decides 'abort'
	{
		capp, _ := suite.app0.Cache()
		canMulticast, isCommitable, err := suite.testConfirmPrepareResult(&capp, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_OK), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)
		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_FAILED), txID, 1, suite.ch2to0, suite.ch0to2, nextSeqSend)
		suite.NoError(err)
		suite.True(canMulticast)
		suite.False(isCommitable)
	}
	// ensure that coordinator decides 'abort' (ordered sequence number)
	{
		capp, _ := suite.app0.Cache()
		canMulticast, isCommitable, err := suite.testConfirmPrepareResult(&capp, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_OK), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)
		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_OK), txID, 1, suite.ch2to0, suite.ch0to2, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)
		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_FAILED), txID, 2, suite.ch2to0, suite.ch0to2, nextSeqSend)
		suite.NoError(err)
		suite.True(canMulticast)
		suite.False(isCommitable)
	}
	// ensure that coordinator decides 'abort' (unordered sequence number)
	{
		capp, _ := suite.app0.Cache()
		canMulticast, isCommitable, err := suite.testConfirmPrepareResult(&capp, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_OK), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)
		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_OK), txID, 2, suite.ch2to0, suite.ch0to2, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)
		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_FAILED), txID, 1, suite.ch2to0, suite.ch0to2, nextSeqSend)
		suite.NoError(err)
		suite.True(canMulticast)
		suite.False(isCommitable)
	}
	// ensure that contractTransaction ID conflict occurs
	{
		capp, _ := suite.app0.Cache()
		canMulticast, isCommitable, err := suite.testConfirmPrepareResult(&capp, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_OK), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)
		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_OK), txID, 0, suite.ch2to0, suite.ch0to2, nextSeqSend)
		suite.Error(err)
	}
	// invalid transactionID
	{
		capp, _ := suite.app0.Cache()
		var invalidTxID types.TxID
		copy(invalidTxID[:], tmhash.Sum(txID[:]))
		_, _, err := suite.testConfirmPrepareResult(&capp, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_OK), invalidTxID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
		suite.Error(err)
	}
	// invalid transactionIndex
	{
		capp, _ := suite.app0.Cache()
		_, _, err := suite.testConfirmPrepareResult(&capp, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_OK), txID, 3, suite.ch1to0, suite.ch0to1, nextSeqSend)
		suite.Error(err)
	}
	// ensure that coordinator doesn't execute to multicast more than once
	{
		capp, _ := suite.app0.Cache()
		canMulticast, isCommitable, err := suite.testConfirmPrepareResult(&capp, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_OK), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)
		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_FAILED), txID, 1, suite.ch2to0, suite.ch0to2, nextSeqSend)
		suite.NoError(err)
		suite.True(canMulticast)
		suite.False(isCommitable)
		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_OK), txID, 2, suite.ch2to0, suite.ch0to2, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)
	}
	// ensure that coordinator doesn't receive a result of same contract call
	{
		capp, _ := suite.app0.Cache()
		canMulticast, isCommitable, err := suite.testConfirmPrepareResult(&capp, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_OK), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)
		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_OK), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
		suite.Error(err)
	}
	// ensure that coordinator decides 'commit' (unordered sequence number)
	{
		capp, _ := suite.app0.Cache()
		canMulticast, isCommitable, err := suite.testConfirmPrepareResult(&capp, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_OK), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)

		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_OK), txID, 2, suite.ch2to0, suite.ch0to2, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)

		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_OK), txID, 1, suite.ch2to0, suite.ch0to2, nextSeqSend)
		suite.NoError(err)
		suite.True(canMulticast)
		suite.True(isCommitable)
	}
	// ensure that coordinator decides 'commit' (ordered sequence number)
	{
		capp, writer := suite.app0.Cache()
		canMulticast, isCommitable, err := suite.testConfirmPrepareResult(&capp, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_OK), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)

		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_OK), txID, 1, suite.ch2to0, suite.ch0to2, nextSeqSend)
		suite.NoError(err)
		suite.False(canMulticast)
		suite.False(isCommitable)

		canMulticast, isCommitable, err = suite.testConfirmPrepareResult(&capp, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_OK), txID, 2, suite.ch2to0, suite.ch0to2, nextSeqSend)
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
			suite.testAbortPacket(&capp, suite.chd1, suite.ch0to1, suite.ch1to0, cross.NewPacketDataCommit(txID, 0, false), suite.signer1)
		}

		// In a2-0, execute to abort
		{
			capp, _ := suite.app2.Cache()
			suite.testAbortPacket(&capp, suite.chd2, suite.ch0to2, suite.ch2to0, cross.NewPacketDataCommit(txID, 1, false), suite.signer2)
		}

		// In a2-1, execute to abort
		{
			capp, _ := suite.app2.Cache()
			suite.testAbortPacket(&capp, suite.chd2, suite.ch0to2, suite.ch2to0, cross.NewPacketDataCommit(txID, 2, false), suite.signer3)
		}

		// In a1, execute to commit
		{
			capp, writer := suite.app1.Cache()
			suite.testCommitPacket(&capp, suite.chd1, suite.ch0to1, suite.ch1to0, cross.NewPacketDataCommit(txID, 0, true), suite.signer1, func(res cross.ContractHandlerResult) {
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
			suite.testCommitPacket(&capp, suite.chd2, suite.ch0to2, suite.ch2to0, cross.NewPacketDataCommit(txID, 1, true), suite.signer2, func(res cross.ContractHandlerResult) {
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
			suite.testCommitPacket(&capp, suite.chd2, suite.ch0to2, suite.ch2to0, cross.NewPacketDataCommit(txID, 2, true), suite.signer3, func(res cross.ContractHandlerResult) {
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

func (suite *KeeperTestSuite) TestAbort1() {
	ci1 := contract.NewContractCallInfo("c1", "issue", [][]byte{[]byte("tone"), []byte("80")})
	ci2 := contract.NewContractCallInfo("c2", "must-fail", nil)

	var err error
	var nonce uint64 = 1
	var tss = []cross.ContractTransaction{
		cross.NewContractTransaction(
			suite.ch0to1,
			[]sdk.AccAddress{suite.signer1},
			ci1.Bytes(),
			cross.NewStateCondition(
				cross.ExactStateCondition,
				[]cross.OP{
					lock.ReadValueOP{K: suite.signer1, V: nil},
					lock.WriteOP{K: suite.signer1, V: marshalCoin(sdk.Coins{sdk.NewInt64Coin("tone", 80)})},
				},
			),
		),
		cross.NewContractTransaction(
			suite.ch0to2,
			[]sdk.AccAddress{suite.signer2},
			ci2.Bytes(),
			cross.NewStateCondition(
				cross.ExactStateCondition,
				[]cross.OP{},
			),
		),
	}

	suite.openAllChannels()
	msg := cross.NewMsgInitiate(
		suite.initiator,
		suite.app0.chainID,
		tss,
		256,
		nonce,
	)
	txID, err := suite.app0.app.CrossKeeper.MulticastPreparePacket(
		suite.app0.ctx,
		suite.initiator,
		msg,
		msg.ContractTransactions,
	)
	suite.NoError(err)

	var nextSeqSend uint64 = 1
	suite.testPreparePacket(suite.app1, suite.ch1to0, suite.ch0to1, txID, 0, suite.chd1, tss[0], nextSeqSend, cross.PREPARE_RESULT_OK)
	suite.testPreparePacket(suite.app2, suite.ch2to0, suite.ch0to2, txID, 1, suite.chd2, tss[1], nextSeqSend, cross.PREPARE_RESULT_FAILED)

	nextSeqSend += 1

	canMulticast, isCommitable, err := suite.testConfirmPrepareResult(suite.app0, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_OK), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
	suite.NoError(err)
	suite.False(canMulticast)
	suite.False(isCommitable)

	canMulticast, isCommitable, err = suite.testConfirmPrepareResult(suite.app0, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_FAILED), txID, 1, suite.ch2to0, suite.ch0to2, nextSeqSend)
	suite.NoError(err)
	suite.True(canMulticast)
	suite.False(isCommitable)

	// In a1, execute to abort
	suite.testAbortPacket(suite.app1, suite.chd1, suite.ch0to1, suite.ch1to0, cross.NewPacketDataCommit(txID, 0, false), suite.signer1)

	// In a2, execute to abort
	suite.testAbortPacket(suite.app2, suite.chd2, suite.ch0to2, suite.ch2to0, cross.NewPacketDataCommit(txID, 1, false), suite.signer2)
}

func (suite *KeeperTestSuite) TestAbort2() {
	ci1 := contract.NewContractCallInfo("c1", "must-fail", nil)
	ci2 := contract.NewContractCallInfo("c2", "must-fail", nil)

	var err error
	var nonce uint64 = 1
	var tss = []cross.ContractTransaction{
		cross.NewContractTransaction(
			suite.ch0to1,
			[]sdk.AccAddress{suite.signer1},
			ci1.Bytes(),
			cross.NewStateCondition(
				cross.ExactStateCondition,
				[]cross.OP{},
			),
		),
		cross.NewContractTransaction(
			suite.ch0to2,
			[]sdk.AccAddress{suite.signer2},
			ci2.Bytes(),
			cross.NewStateCondition(
				cross.ExactStateCondition,
				[]cross.OP{},
			),
		),
	}

	suite.openAllChannels()
	msg := cross.NewMsgInitiate(
		suite.initiator,
		suite.app0.chainID,
		tss,
		256,
		nonce,
	)
	txID, err := suite.app0.app.CrossKeeper.MulticastPreparePacket(
		suite.app0.ctx,
		suite.initiator,
		msg,
		msg.ContractTransactions,
	)
	suite.NoError(err)

	var nextSeqSend uint64 = 1
	suite.testPreparePacket(suite.app1, suite.ch1to0, suite.ch0to1, txID, 0, suite.chd1, tss[0], nextSeqSend, cross.PREPARE_RESULT_FAILED)
	suite.testPreparePacket(suite.app2, suite.ch2to0, suite.ch0to2, txID, 1, suite.chd2, tss[1], nextSeqSend, cross.PREPARE_RESULT_FAILED)

	nextSeqSend += 1

	canMulticast, isCommitable, err := suite.testConfirmPrepareResult(suite.app0, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_FAILED), txID, 0, suite.ch1to0, suite.ch0to1, nextSeqSend)
	suite.NoError(err)
	suite.True(canMulticast)
	suite.False(isCommitable)

	canMulticast, isCommitable, err = suite.testConfirmPrepareResult(suite.app0, cross.NewPacketPrepareAcknowledgement(cross.PREPARE_RESULT_FAILED), txID, 1, suite.ch2to0, suite.ch0to2, nextSeqSend)
	suite.NoError(err)
	suite.False(canMulticast)
	suite.False(isCommitable)

	// In a1, execute to abort
	suite.testAbortPacket(suite.app1, suite.chd1, suite.ch0to1, suite.ch1to0, cross.NewPacketDataCommit(txID, 0, false), suite.signer1)

	// In a2, execute to abort
	suite.testAbortPacket(suite.app2, suite.chd2, suite.ch0to2, suite.ch2to0, cross.NewPacketDataCommit(txID, 1, false), suite.signer2)
}

func (suite *KeeperTestSuite) testPreparePacket(actx *appContext, src, dst cross.ChannelInfo, txID types.TxID, txIndex types.TxIndex, contractHandler cross.ContractHandler, ts cross.ContractTransaction, nextseq uint64, expectedPrepareResult uint8) {
	relayer := sdk.AccAddress("relayer1")
	packetData := cross.NewPacketDataPrepare(relayer, txID, txIndex, ts)
	ctx, writer := actx.ctx.CacheContext()
	ctx = cross.WithSigners(ctx, ts.Signers)
	result, err := actx.app.CrossKeeper.PrepareTransaction(
		ctx,
		contractHandler,
		dst.Port,
		dst.Channel,
		src.Port,
		src.Channel,
		packetData,
	)
	suite.NoError(err)
	suite.Equal(expectedPrepareResult, result)
	tx, ok := actx.app.CrossKeeper.GetTx(ctx, txID, txIndex)
	if suite.True(ok) {
		suite.Equal(cross.TX_STATUS_PREPARE, tx.Status)
	}
	suite.Equal(expectedPrepareResult, tx.PrepareResult)

	writer()
}

func (suite *KeeperTestSuite) testConfirmPrepareResult(actx *appContext, ack cross.PacketPrepareAcknowledgement, txID cross.TxID, txIndex cross.TxIndex, src, dst cross.ChannelInfo, nextseq uint64) (bool, bool, error) {
	canMulticast, isCommitable, err := actx.app.CrossKeeper.ReceivePrepareAcknowledgement(actx.ctx, dst.Port, dst.Channel, ack, txID, txIndex)
	if err != nil {
		return false, false, err
	}
	if canMulticast {
		return canMulticast, isCommitable, actx.app.CrossKeeper.MulticastCommitPacket(actx.ctx, txID, isCommitable)
	} else {
		return canMulticast, isCommitable, nil
	}
}

func (suite *KeeperTestSuite) testAbortPacket(actx *appContext, contractHandler cross.ContractHandler, src, dst cross.ChannelInfo, packet cross.PacketDataCommit, txSigner sdk.AccAddress) {
	_, err := actx.app.CrossKeeper.ReceiveCommitPacket(actx.ctx, contractHandler, src.Port, src.Channel, dst.Port, dst.Channel, packet)
	if !suite.NoError(err) {
		return
	}
	tx, found := actx.app.CrossKeeper.GetTx(actx.ctx, packet.TxID, packet.TxIndex)
	if !suite.True(found) {
		return
	}
	suite.Equal(cross.TX_STATUS_ABORT, tx.Status)
	// ensure that the state is expected
	_, err = contractHandler.GetState(actx.ctx, cross.NoStateCondition, tx.ContractCallInfo)
	if !suite.NoError(err) {
		return
	}
	ci, err := contract.DecodeContractSignature(tx.ContractCallInfo)
	if !suite.NoError(err) {
		return
	}
	contractInfo := contract.NewContractCallInfo(ci.ID, "test-not-issued", [][]byte{})
	bz, err := contract.EncodeContractSignature(contractInfo)
	if !suite.NoError(err) {
		return
	}
	actx2, _ := actx.Cache()
	ctx := cross.WithSigners(actx2.ctx, []sdk.AccAddress{txSigner})
	_, _, err = contractHandler.Handle(ctx, cross.ExactStateCondition, bz)
	suite.NoError(err)
}

func (suite *KeeperTestSuite) testCommitPacket(actx *appContext, contractHandler cross.ContractHandler, src, dst cross.ChannelInfo, packet cross.PacketDataCommit, txSigner sdk.AccAddress, checkResult func(cross.ContractHandlerResult)) {
	res, err := actx.app.CrossKeeper.ReceiveCommitPacket(actx.ctx, contractHandler, src.Port, src.Channel, dst.Port, dst.Channel, packet)
	if !suite.NoError(err) {
		return
	}
	checkResult(res)
	tx, found := actx.app.CrossKeeper.GetTx(actx.ctx, packet.TxID, packet.TxIndex)
	if !suite.True(found) {
		return
	}
	suite.Equal(cross.TX_STATUS_COMMIT, tx.Status)
	// ensure that the state is expected
	_, err = contractHandler.GetState(actx.ctx, cross.ExactStateCondition, tx.ContractCallInfo)
	if !suite.NoError(err) {
		return
	}
	ci, err := contract.DecodeContractSignature(tx.ContractCallInfo)
	if !suite.NoError(err) {
		return
	}
	contractInfo := contract.NewContractCallInfo(ci.ID, "test-balance", [][]byte{
		ci.Args[0],
		ci.Args[1],
	})
	bz, err := contract.EncodeContractSignature(contractInfo)
	if !suite.NoError(err) {
		return
	}
	ctx := cross.WithSigners(actx.ctx, []sdk.AccAddress{txSigner})
	_, _, err = contractHandler.Handle(ctx, cross.ExactStateCondition, bz)
	suite.NoError(err)
}

func parseCoin(ctx contract.Context, denomIdx, amountIdx int) (sdk.Coin, error) {
	denom := string(ctx.Args()[denomIdx])
	amount, err := strconv.Atoi(string(ctx.Args()[amountIdx]))
	if err != nil {
		return sdk.Coin{}, err
	}
	if amount < 0 {
		return sdk.Coin{}, fmt.Errorf("amount must be positive number")
	}
	coin := sdk.NewInt64Coin(denom, int64(amount))
	return coin, nil
}

func marshalCoin(coins sdk.Coins) []byte {
	return testcdc.MustMarshalBinaryLengthPrefixed(coins)
}

func unmarshalCoin(bz []byte) sdk.Coins {
	var coins sdk.Coins
	testcdc.MustUnmarshalBinaryLengthPrefixed(bz, &coins)
	return coins
}

func getBalanceOf(store cross.Store, address sdk.AccAddress) sdk.Coins {
	bz := store.Get(address)
	if bz == nil {
		return sdk.NewCoins()
	}
	return unmarshalCoin(bz)
}

func setBalance(store cross.Store, address sdk.AccAddress, balance sdk.Coins) {
	bz := marshalCoin(balance)
	store.Set(address, bz)
}

var testcdc *codec.Codec

func init() {
	testcdc = codec.New()

	testcdc.RegisterConcrete(sdk.Coin{}, "sdk/Coin", nil)
	testcdc.RegisterConcrete(sdk.Coins{}, "sdk/Coins", nil)
}
