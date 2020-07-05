package keeper_test

import (
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/datachainlab/cross/example/simapp"
	"github.com/datachainlab/cross/x/ibc/cross"
	"github.com/datachainlab/cross/x/ibc/cross/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
)

type KeeperTestSuite struct {
	suite.Suite
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

func (suite *KeeperTestSuite) SetupSuite() {}

type appContext struct {
	chainID string
	cdc     *codec.Codec
	ctx     sdk.Context
	app     *simapp.SimApp
	valSet  *tmtypes.ValidatorSet
	signers []tmtypes.PrivValidator

	// src => dst
	channels map[types.ChannelInfo]types.ChannelInfo
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
	srcConnectionID string, srcapp *appContext, srcc types.ChannelInfo,
	dstConnectionID string, dstapp *appContext, dstc types.ChannelInfo,
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
	srcc types.ChannelInfo, // src's channel with dstapp
	srcapp *appContext,

	dstClientID string, // clientID of srcapp
	dstConnectionID string, // id of the connection with srcapp
	dstc types.ChannelInfo, // dst's channel with srcapp
	dstapp *appContext,
) {
	suite.createClients(srcClientID, srcapp, dstClientID, dstapp)
	suite.createConnections(srcClientID, srcConnectionID, srcapp, dstClientID, dstConnectionID, dstapp)
	suite.createChannels(srcConnectionID, srcapp, srcc, dstConnectionID, dstapp, dstc)
}

func (suite *KeeperTestSuite) createApp(chainID string) *appContext {
	return suite.createAppWithHeader(abci.Header{ChainID: chainID}, simapp.DefaultContractHandlerProvider)
}

func (suite *KeeperTestSuite) createAppWithHeader(header abci.Header, contractHandlerProvider simapp.ContractHandlerProvider) *appContext {
	isCheckTx := false
	app := simapp.SetupWithContractHandlerProvider(isCheckTx, contractHandlerProvider, simapp.DefaultAnteHandlerProvider)
	ctx := app.BaseApp.NewContext(isCheckTx, header)
	ctx = ctx.WithLogger(log.NewTMLogger(os.Stdout))
	if testing.Verbose() {
		ctx = ctx.WithLogger(
			log.NewFilter(
				ctx.Logger(),
				log.AllowDebugWith("module", "cross/cross"),
			),
		)
	} else {
		ctx = ctx.WithLogger(
			log.NewFilter(
				ctx.Logger(),
				log.AllowErrorWith("module", "cross/cross"),
			),
		)
	}
	privVal := tmtypes.NewMockPV()
	pub, err := privVal.GetPubKey()
	if err != nil {
		panic(err)
	}
	validator := tmtypes.NewValidator(pub, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})
	signers := []tmtypes.PrivValidator{privVal}

	actx := &appContext{
		chainID:  header.GetChainID(),
		cdc:      app.Codec(),
		ctx:      ctx,
		app:      app,
		valSet:   valSet,
		signers:  signers,
		channels: make(map[types.ChannelInfo]types.ChannelInfo),
	}

	updateApp(actx, int(header.Height))

	return actx
}

func updateApp(actx *appContext, n int) {
	for i := 0; i < n; i++ {
		actx.app.Commit()
		actx.app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: actx.ctx.ChainID(), Height: actx.app.LastBlockHeight() + 1}})
		actx.ctx = actx.ctx.WithBlockHeader(abci.Header{ChainID: actx.ctx.ChainID()})
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
