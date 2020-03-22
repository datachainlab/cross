package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/datachainlab/cross/example/simapp"
	"github.com/datachainlab/cross/x/ibc/cross"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

type KeeperTestSuite struct {
	suite.Suite
}

func (suite *KeeperTestSuite) SetupSuite() {}

type appContext struct {
	chainID string
	cdc     *codec.Codec
	ctx     sdk.Context
	app     *simapp.SimApp
	valSet  *tmtypes.ValidatorSet
	signers []tmtypes.PrivValidator

	// src => dst
	channels map[cross.ChannelInfo]cross.ChannelInfo
}

func (a appContext) Cache() (appContext, func()) {
	ctx, writer := a.ctx.CacheContext()
	a.ctx = ctx
	return a, writer
}

func (suite *KeeperTestSuite) createClients(
	srcClientID string, // clientID of dstapp
	srcapp *appContext,
	dstClientID string, // clientID of srcapp
	dstapp *appContext,
) {
	suite.createClient(srcapp, srcClientID)
	suite.createClient(dstapp, dstClientID)
}

func (suite *KeeperTestSuite) createConnections(
	srcClientID string,
	srcConnectionID string,
	srcapp *appContext,

	dstClientID string,
	dstConnectionID string,
	dstapp *appContext,
) {
	suite.createConnection(srcapp, srcClientID, srcConnectionID, dstClientID, dstConnectionID, connectionexported.OPEN)
	suite.createConnection(dstapp, dstClientID, dstConnectionID, srcClientID, srcConnectionID, connectionexported.OPEN)
}

func (suite *KeeperTestSuite) createChannels(
	srcConnectionID string, srcapp *appContext, srcc cross.ChannelInfo,
	dstConnectionID string, dstapp *appContext, dstc cross.ChannelInfo,
) {
	suite.createChannel(srcapp, srcc.Port, srcc.Channel, srcConnectionID, dstc.Port, dstc.Channel, channelexported.OPEN)
	suite.createChannel(dstapp, dstc.Port, dstc.Channel, dstConnectionID, srcc.Port, srcc.Channel, channelexported.OPEN)

	nextSeqSend := uint64(1)
	srcapp.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(srcapp.ctx, srcc.Port, srcc.Channel, nextSeqSend)
	dstapp.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(dstapp.ctx, dstc.Port, dstc.Channel, nextSeqSend)

	srcapp.channels[srcc] = dstc
	dstapp.channels[dstc] = srcc
}

func (suite *KeeperTestSuite) openChannels(
	srcClientID string, // clientID of dstapp
	srcConnectionID string, // id of the connection with dstapp
	srcc cross.ChannelInfo, // src's channel with dstapp
	srcapp *appContext,

	dstClientID string, // clientID of srcapp
	dstConnectionID string, // id of the connection with srcapp
	dstc cross.ChannelInfo, // dst's channel with srcapp
	dstapp *appContext,
) {
	suite.createClients(srcClientID, srcapp, dstClientID, dstapp)
	suite.createConnections(srcClientID, srcConnectionID, srcapp, dstClientID, dstConnectionID, dstapp)
	suite.createChannels(srcConnectionID, srcapp, srcc, dstConnectionID, dstapp, dstc)
}

func (suite *KeeperTestSuite) createApp(chainID string) *appContext {
	return suite.createAppWithHeader(abci.Header{ChainID: chainID})
}

func (suite *KeeperTestSuite) createAppWithHeader(header abci.Header) *appContext {
	isCheckTx := false
	app := simapp.Setup(isCheckTx)
	ctx := app.BaseApp.NewContext(isCheckTx, header)
	privVal := tmtypes.NewMockPV()
	validator := tmtypes.NewValidator(privVal.GetPubKey(), 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})
	signers := []tmtypes.PrivValidator{privVal}

	actx := &appContext{
		chainID:  header.GetChainID(),
		cdc:      app.Codec(),
		ctx:      ctx,
		app:      app,
		valSet:   valSet,
		signers:  signers,
		channels: make(map[cross.ChannelInfo]cross.ChannelInfo),
	}

	updateApp(actx, int(header.Height))

	return actx
}

func updateApp(actx *appContext, n int) {
	for i := 0; i < n; i++ {
		actx.app.Commit()
		actx.app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: actx.ctx.ChainID(), Height: actx.app.LastBlockHeight() + 1}})
		actx.ctx = actx.app.BaseApp.NewContext(false, abci.Header{ChainID: actx.ctx.ChainID()})
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
