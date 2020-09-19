package keeper_test

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

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
	"github.com/datachainlab/cross/example/simapp"
	"github.com/datachainlab/cross/x/ibc/contract"
	"github.com/datachainlab/cross/x/ibc/cross"
	"github.com/datachainlab/cross/x/ibc/cross/types"
	"github.com/datachainlab/cross/x/ibc/store/lock"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
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
	channels map[types.ChannelInfo]types.ChannelInfo
}

func (a appContext) Cache() (appContext, func()) {
	ctx, writer := a.ctx.CacheContext()
	a.ctx = ctx
	return a, writer
}

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

func (suite *KeeperTestSuite) createContractHandler(k contract.Keeper, cid string, channelResolver cross.ChannelResolver) cross.ContractHandler {
	contractHandler := contract.NewContractHandler(k, func(kvs sdk.KVStore, tp cross.StateConstraintType) cross.State {
		return lock.NewStore(kvs, tp)
	}, channelResolver)
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
		{
			Name: "peg-coin",
			F: func(ctx contract.Context, store cross.Store) ([]byte, error) {
				sender := ctx.Signers()[0]
				_, err := parseCoin(ctx, 0, 1)
				if err != nil {
					return nil, err
				}

				var args = [][]byte{
					ctx.Args()[0],
					ctx.Args()[1],
				}
				amount := unmarshalCoins(
					contract.CallExternalFunc(ctx, types.NewChannelInfo("cross", "testchanneltwo"), contract.NewContractCallInfo("c2", "lock-coin", args), []sdk.AccAddress{sender}),
				)
				setBalance(store, sender, amount)
				return marshalCoins(amount), nil
			},
		},
		{
			Name: "lock-coin",
			F: func(ctx contract.Context, store cross.Store) ([]byte, error) {
				coin, err := parseCoin(ctx, 0, 1)
				if err != nil {
					return nil, err
				}
				amount := sdk.NewCoins(coin)
				from := ctx.Signers()[0]
				fromBalance := getBalanceOf(store, from)
				fromBalance, isNega := fromBalance.SafeSub(amount)
				if isNega {
					return nil, fmt.Errorf("insuficient balance")
				}
				setBalance(store, from, fromBalance)
				return marshalCoins(amount), nil
			},
		},
	})
	contractHandler.AddRoute(cid, c)
	return contractHandler
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
	return suite.createAppWithHeader(abci.Header{ChainID: chainID}, simapp.DefaultContractHandlerProvider, simapp.DefaultChannelResolverProvider)
}

func (suite *KeeperTestSuite) createAppWithHeader(header abci.Header, contractHandlerProvider simapp.ContractHandlerProvider, channelResolverProvider simapp.ChannelResolverProvider) *appContext {
	isCheckTx := false
	app := simapp.SetupWithOptions(isCheckTx, contractHandlerProvider, channelResolverProvider, simapp.DefaultAnteHandlerProvider)
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
