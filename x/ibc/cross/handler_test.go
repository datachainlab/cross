// This test suite is based on [ibc/transfer](https://github.com/cosmos/cosmos-sdk/blob/4d5c2d1f9e24f20f740d42c642f9fb5378e31f9e/x/ibc/20-transfer/handler_test.go)
package cross_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/datachainlab/cross/example/simapp"
	"github.com/datachainlab/cross/x/ibc/contract"
	"github.com/datachainlab/cross/x/ibc/cross"
	lock "github.com/datachainlab/cross/x/ibc/store/lock"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

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

type HandlerTestSuite struct {
	suite.Suite

	cdc *codec.Codec
	ctx sdk.Context
	app *simapp.SimApp
}

func init() {
	lock.RegisterCodec(cross.ModuleCdc)
}

func (suite *HandlerTestSuite) SetupTest() {
	isCheckTx := false
	app := simapp.Setup(isCheckTx)

	suite.cdc = app.Codec()
	suite.ctx = app.BaseApp.NewContext(isCheckTx, abci.Header{})
	suite.app = app

	suite.createClient()
	suite.createConnection(connectionexported.OPEN)
}

const (
	trustingPeriod time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod      time.Duration = time.Hour * 24 * 7 * 3
)

func (suite *HandlerTestSuite) createClient() {
	suite.app.Commit()

	h := abci.Header{Height: suite.app.LastBlockHeight() + 1}
	suite.app.BeginBlock(abci.RequestBeginBlock{Header: h})
	suite.ctx = suite.app.BaseApp.NewContext(false, abci.Header{})

	privVal := tmtypes.NewMockPV()
	validator := tmtypes.NewValidator(privVal.GetPubKey(), 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})
	signers := []tmtypes.PrivValidator{privVal}
	now := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)

	header := tendermint.CreateTestHeader(testChainID, 1, now, valSet, valSet, signers)
	consensusState := header.ConsensusState()

	// create client
	clientState, err := tendermint.Initialize(testClient, trustingPeriod, ubdPeriod, header)
	if err != nil {
		panic(err)
	}
	_, err = suite.app.IBCKeeper.ClientKeeper.CreateClient(suite.ctx, clientState, consensusState)
	suite.NoError(err)
}

func (suite *HandlerTestSuite) updateClient() {
	// always commit and begin a new block on updateClient
	suite.app.Commit()
	commitID := suite.app.LastCommitID()

	suite.app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: suite.app.LastBlockHeight() + 1}})
	suite.ctx = suite.app.BaseApp.NewContext(false, abci.Header{})

	state := tendermint.ConsensusState{
		Root: commitment.NewMerkleRoot(commitID.Hash),
	}

	suite.app.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.ctx, testClient, 1, state)
}

func (suite *HandlerTestSuite) createConnection(state connectionexported.State) {
	connection := connection.ConnectionEnd{
		State:    state,
		ClientID: testClient,
		Counterparty: connection.Counterparty{
			ClientID:     testClient,
			ConnectionID: testConnection,
			Prefix:       suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(),
		},
		Versions: connection.GetCompatibleVersions(),
	}

	suite.app.IBCKeeper.ConnectionKeeper.SetConnection(suite.ctx, testConnection, connection)
}

func (suite *HandlerTestSuite) createChannel(
	portID, channnelID, connnnectionID, counterpartyPortID, counterpartyChannelID string, state channelexported.State) {
	ch := channel.Channel{
		State:    state,
		Ordering: testChannelOrder,
		Counterparty: channel.Counterparty{
			PortID:    counterpartyPortID,
			ChannelID: counterpartyChannelID,
		},
		ConnectionHops: []string{connnnectionID},
		Version:        testChannelVersion,
	}

	suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, portID, channnelID, ch)
}

func (suite *HandlerTestSuite) queryProof(key []byte) (proof commitmentexported.Proof, height int64) {
	res := suite.app.Query(abci.RequestQuery{
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

func (suite *HandlerTestSuite) TestHandleCrossc() {
	stk := sdk.NewKVStoreKey("main")
	contractHandler := contract.NewContractHandler(contract.NewKeeper(suite.app.Codec(), stk), func(kvs sdk.KVStore) cross.State {
		return lock.NewStore(kvs)
	})
	handler := cross.NewHandler(suite.app.CrossKeeper, contractHandler)
	coordinator := sdk.AccAddress("coordinator")

	signer0 := sdk.AccAddress("signerzero")
	src0 := cross.NewChannelInfo("testportzero", "testchannelzero")
	ci0 := contract.NewContractCallInfo("c0", "issue", [][]byte{[]byte("100")})
	dst0 := cross.NewChannelInfo("dstportzero", "dstchannelzero")

	signer1 := sdk.AccAddress("signerfirst")
	src1 := cross.NewChannelInfo("testportone", "testchannelone")
	ci1 := contract.NewContractCallInfo("c1", "issue", [][]byte{[]byte("100")})
	dst1 := cross.NewChannelInfo("dstportone", "dstchannelone")

	var nonce uint64 = 1
	var tss = []cross.ContractTransaction{
		cross.NewContractTransaction(
			src0,
			[]sdk.AccAddress{signer0},
			ci0.Bytes(),
			[]cross.OP{lock.Read{}, lock.Write{}},
		),
		cross.NewContractTransaction(
			src1,
			[]sdk.AccAddress{signer1},
			ci1.Bytes(),
			[]cross.OP{lock.Read{}, lock.Write{}},
		),
	}

	msg := cross.NewMsgInitiate(coordinator, "", tss, 256, nonce)
	res, err := handler(suite.ctx, msg)
	suite.Require().Error(err)
	suite.Require().Nil(res, "%+v", res) // channel does not exist

	suite.createChannel(src0.Port, src0.Channel, testConnection, dst0.Port, dst0.Channel, channelexported.OPEN)
	suite.createChannel(src1.Port, src1.Channel, testConnection, dst1.Port, dst1.Channel, channelexported.OPEN)

	res, err = handler(suite.ctx, msg)
	suite.Require().Error(err)
	suite.Require().Nil(res, "%+v", res) // next send sequence not found

	nextSeqSend := uint64(1)
	suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, src0.Port, src0.Channel, nextSeqSend)
	suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, src1.Port, src1.Channel, nextSeqSend)
	res, err = handler(suite.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res, "%+v", res) // successfully executed
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
