package keeper_test

import (
	"fmt"

	"github.com/bluele/crossccc/x/ibc/contract"
	"github.com/bluele/crossccc/x/ibc/crossccc"
	"github.com/bluele/crossccc/x/ibc/store/lock"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// define constants used for testing
const (
	testChainID    = "test-chain-id"
	testClient     = "test-client"
	testClientType = clientexported.Tendermint

	testConnection     = "testconnection"
	testChannelOrder   = channelexported.UNORDERED
	testChannelVersion = "1.0"
)

func (suite *KeeperTestSuite) createClient(actx *appContext) {
	actx.app.Commit()
	commitID := actx.app.LastCommitID()

	actx.app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: actx.app.LastBlockHeight() + 1}})
	actx.ctx = actx.app.BaseApp.NewContext(false, abci.Header{})

	consensusState := tendermint.ConsensusState{
		Root:             commitment.NewRoot(commitID.Hash),
		ValidatorSetHash: actx.valSet.Hash(),
	}

	_, err := actx.app.IBCKeeper.ClientKeeper.CreateClient(actx.ctx, testClient, testClientType, consensusState)
	suite.NoError(err)
}

func (suite *KeeperTestSuite) updateClient(actx *appContext) {
	// always commit and begin a new block on updateClient
	actx.app.Commit()
	commitID := actx.app.LastCommitID()

	actx.app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: actx.app.LastBlockHeight() + 1}})
	actx.ctx = actx.app.BaseApp.NewContext(false, abci.Header{})

	state := tendermint.ConsensusState{
		Root: commitment.NewRoot(commitID.Hash),
	}

	actx.app.IBCKeeper.ClientKeeper.SetClientConsensusState(actx.ctx, testClient, 1, state)
}

func (suite *KeeperTestSuite) createConnection(actx *appContext, state connectionexported.State) {
	connection := connection.ConnectionEnd{
		State:    state,
		ClientID: testClient,
		Counterparty: connection.Counterparty{
			ClientID:     testClient,
			ConnectionID: testConnection,
			Prefix:       actx.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(),
		},
		Versions: connection.GetCompatibleVersions(),
	}

	actx.app.IBCKeeper.ConnectionKeeper.SetConnection(actx.ctx, testConnection, connection)
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
}

func (suite *KeeperTestSuite) queryProof(actx *appContext, key []byte) (proof commitment.Proof, height int64) {
	res := actx.app.Query(abci.RequestQuery{
		Path:  fmt.Sprintf("store/%s/key", ibctypes.StoreKey),
		Data:  key,
		Prove: true,
	})

	height = res.Height
	proof = commitment.Proof{
		Proof: res.Proof,
	}

	return
}

func (suite *KeeperTestSuite) TestSendInitiate() {
	lock.RegisterCodec(crossccc.ModuleCdc)

	coordinator := sdk.AccAddress("coordinator")

	signer0 := sdk.AccAddress("signerzero")
	src0 := crossccc.NewChannelInfo("testportzero", "testchannelzero")
	ci0 := contract.NewContractInfo("c0", "issue", [][]byte{[]byte("100")})
	dst0 := crossccc.NewChannelInfo("dstportzero", "dstchannelzero")

	signer1 := sdk.AccAddress("signerfirst")
	src1 := crossccc.NewChannelInfo("testportone", "testchannelone")
	ci1 := contract.NewContractInfo("c1", "issue", [][]byte{[]byte("100")})
	dst1 := crossccc.NewChannelInfo("dstportone", "dstchannelone")

	var err error
	var nonce uint64 = 1
	var tss = []crossccc.StateTransition{
		crossccc.NewStateTransition(
			src0,
			signer0,
			ci0.Bytes(),
			[]crossccc.OP{lock.Read{}, lock.Write{}},
		),
		crossccc.NewStateTransition(
			src1,
			signer1,
			ci1.Bytes(),
			[]crossccc.OP{lock.Read{}, lock.Write{}},
		),
	}

	msg := crossccc.NewMsgInitiate(
		coordinator,
		tss,
		nonce,
	)
	actx0 := suite.createApp()
	err = actx0.app.CrosscccKeeper.MulticastInitiatePacket(
		actx0.ctx,
		coordinator,
		msg,
		msg.StateTransitions,
	)
	suite.Error(err) // channel does not exist

	suite.createChannel(actx0, src0.Port, src0.Channel, testConnection, dst0.Port, dst0.Channel, channelexported.OPEN)
	suite.createChannel(actx0, src1.Port, src1.Channel, testConnection, dst1.Port, dst1.Channel, channelexported.OPEN)
	nextSeqSend := uint64(1)
	actx0.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(actx0.ctx, src0.Port, src0.Channel, nextSeqSend)
	actx0.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(actx0.ctx, src1.Port, src1.Channel, nextSeqSend)

	err = actx0.app.CrosscccKeeper.MulticastInitiatePacket(
		actx0.ctx,
		coordinator,
		msg,
		msg.StateTransitions,
	)
	suite.NoError(err) // successfully executed

	ci, found := actx0.app.CrosscccKeeper.GetCoordinator(actx0.ctx, msg.GetTxID())
	if suite.True(found) {
		suite.Equal(ci.Status, crossccc.CO_STATUS_INIT)
	}

	packetCommitment := actx0.app.IBCKeeper.ChannelKeeper.GetPacketCommitment(actx0.ctx, src0.Port, src0.Channel, nextSeqSend)
	suite.NotNil(packetCommitment)
	packetCommitment = actx0.app.IBCKeeper.ChannelKeeper.GetPacketCommitment(actx0.ctx, src1.Port, src1.Channel, nextSeqSend)
	suite.NotNil(packetCommitment)
}
