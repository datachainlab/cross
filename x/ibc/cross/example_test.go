package cross_test

import (
	"fmt"
	"testing"
	"time"

	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec"
	codecstd "github.com/cosmos/cosmos-sdk/codec/std"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	ibcante "github.com/cosmos/cosmos-sdk/x/ibc/ante"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
	"github.com/datachainlab/cross/example/simapp"
	simappcontract "github.com/datachainlab/cross/example/simapp/contract"
	"github.com/datachainlab/cross/x/ibc/contract"
	"github.com/datachainlab/cross/x/ibc/cross"
	"github.com/datachainlab/cross/x/ibc/store/lock"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

type ExampleTestSuite struct {
	suite.Suite
	accountSeqs map[string]uint64 // chainID/AccAddress => seq
}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(ExampleTestSuite))
}

func (suite *ExampleTestSuite) SetupSuite() {
	suite.accountSeqs = make(map[string]uint64)
}

func createMnemonics(kb keyring.Keyring, names ...string) ([]keyring.Info, error) {
	var infos []keyring.Info
	for _, name := range names {
		info, _, err := kb.NewMnemonic(
			name, keyring.English, clientkeys.DefaultKeyPass, hd.Secp256k1,
		)
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}
	return infos, nil
}

func getAnteHandler(app *simapp.SimApp) sdk.AnteHandler {
	ak := app.AccountKeeper
	sigGasConsumer := ante.DefaultSigVerificationGasConsumer
	ibcKeeper := app.IBCKeeper
	return sdk.ChainAnteDecorators(
		// ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		ante.NewMempoolFeeDecorator(),
		ante.NewValidateBasicDecorator(),
		ante.NewValidateMemoDecorator(ak),
		ante.NewConsumeGasForTxSizeDecorator(ak),
		ante.NewSetPubKeyDecorator(ak), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(ak),
		ante.NewDeductFeeDecorator(ak, app.BankKeeper),
		ante.NewSigGasConsumeDecorator(ak, sigGasConsumer),
		ante.NewSigVerificationDecorator(ak),
		ante.NewIncrementSequenceDecorator(ak),
		ibcante.NewProofVerificationDecorator(ibcKeeper.ClientKeeper, ibcKeeper.ChannelKeeper), // innermost AnteDecorator
	)
}

func (suite *ExampleTestSuite) TestTrainAndHotelProblem() {
	kb := keyring.NewInMemory()
	signer0, signer1, signer2 := "signer0", "signer1", "signer2"
	relayer0 := "relayer0"
	infos, err := createMnemonics(kb, signer0, signer1, signer2, relayer0)
	if err != nil {
		suite.FailNow(err.Error())
	}
	signer0Info, signer1Info, signer2Info, relayer0Info := infos[0], infos[1], infos[2], infos[3]

	signer0Acc := authtypes.NewBaseAccountWithAddress(signer0Info.GetAddress())
	signer1Acc := authtypes.NewBaseAccountWithAddress(signer1Info.GetAddress())
	signer2Acc := authtypes.NewBaseAccountWithAddress(signer2Info.GetAddress())
	relayer0Acc := authtypes.NewBaseAccountWithAddress(relayer0Info.GetAddress())

	cdc := codecstd.MakeCodec(simapp.ModuleBasics)
	cdc.Seal()

	txBuilder := authtypes.NewTxBuilder(
		authclient.GetTxEncoder(cdc),
		0,
		0,
		0,
		0,
		false,
		"",
		"",
		sdk.NewCoins(),
		sdk.NewDecCoins(),
	).WithKeybase(kb)

	app0 := suite.createApp("app0", simapp.DefaultContractHandlerProvider, getAnteHandler, []authexported.GenesisAccount{signer0Acc, signer1Acc, signer2Acc, relayer0Acc}) // coordinator node
	app1 := suite.createApp("app1", simappcontract.TrainReservationContractHandler, getAnteHandler, []authexported.GenesisAccount{signer0Acc, signer1Acc, signer2Acc, relayer0Acc})
	app2 := suite.createApp("app2", simappcontract.HotelReservationContractHandler, getAnteHandler, []authexported.GenesisAccount{signer0Acc, signer1Acc, signer2Acc, relayer0Acc})

	ch0to1 := cross.NewChannelInfo(cross.RouterKey, "testchannelzeroone") // app0 -> app1
	ch1to0 := cross.NewChannelInfo(cross.RouterKey, "testchannelonezero") // app1 -> app
	ch0to2 := cross.NewChannelInfo(cross.RouterKey, "testchannelzerotwo") // app0 -> app2
	ch2to0 := cross.NewChannelInfo(cross.RouterKey, "testchanneltwozero") // app2 -> app0

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

	trainCall := contract.NewContractCallInfo(simappcontract.TrainContractID, simappcontract.ReserveFnName, [][]byte{contract.ToBytes(int32(1))})
	hotelCall := contract.NewContractCallInfo(simappcontract.HotelContractID, simappcontract.ReserveFnName, [][]byte{contract.ToBytes(int32(8))})

	var tss = []cross.ContractTransaction{
		cross.NewContractTransaction(
			ch0to1,
			[]sdk.AccAddress{signer1Info.GetAddress()},
			trainCall.Bytes(),
			[]cross.OP{lock.Write{K: simappcontract.MakeSeatKey(1), V: signer1Info.GetAddress()}},
		),
		cross.NewContractTransaction(
			ch0to2,
			[]sdk.AccAddress{signer2Info.GetAddress()},
			hotelCall.Bytes(),
			[]cross.OP{lock.Write{K: simappcontract.MakeRoomKey(8), V: signer2Info.GetAddress()}},
		),
	}
	var txID cross.TxID
	{
		var nonce uint64 = 1
		msg := cross.NewMsgInitiate(
			signer0Info.GetAddress(),
			app0.chainID,
			tss,
			256,
			nonce,
		)
		suite.NoError(msg.ValidateBasic())
		txID = cross.MakeTxID(app0.ctx, msg)
		stdTx := authtypes.NewStdTx([]sdk.Msg{msg}, authtypes.StdFee{}, nil, "")
		for i, signer := range []string{signer0, signer1, signer2} {
			stdTx, err = txBuilder.WithChainID(app0.chainID).WithAccountNumber(uint64(i)).SignStdTx(signer, clientkeys.DefaultKeyPass, stdTx, true)
			if err != nil {
				suite.FailNow(err.Error())
			}
		}
		txBytes, err := txBuilder.WithChainID(app0.chainID).TxEncoder()(stdTx)
		if err != nil {
			suite.FailNow(err.Error())
		}

		res := app0.app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
		if !suite.True(res.IsOK()) {
			suite.FailNow(res.String())
			return
		}

		suite.nextBlock(app0)
	}

	{ // update client
		apps := []*appContext{app1, app2}
		for _, app := range apps {
			suite.updateClient(app0, app.chainID, app)
			suite.updateClient(app, app0.chainID, app0)
		}
	}

	var packetSeq uint64 = 1

	// doPrepare

	{ // execute Train contract on app1
		data := cross.NewPacketDataPrepare(
			signer0Info.GetAddress(),
			txID,
			0,
			tss[0],
		)
		packet := channeltypes.NewPacket(
			data.GetBytes(),
			packetSeq, ch0to1.Port, ch0to1.Channel, ch1to0.Port, ch1to0.Channel, data.GetTimeoutHeight())
		suite.buildMsgAndDoRelay(packet, app0, app1, txID, relayer0Info, txBuilder, packetSeq)
	}

	{ // execute Hotel contract on app2
		data := cross.NewPacketDataPrepare(
			signer0Info.GetAddress(),
			txID,
			1,
			tss[1],
		)
		packet := channeltypes.NewPacket(
			data.GetBytes(),
			packetSeq, ch0to2.Port, ch0to2.Channel, ch2to0.Port, ch2to0.Channel, data.GetTimeoutHeight())
		suite.buildMsgAndDoRelay(packet, app0, app2, txID, relayer0Info, txBuilder, packetSeq)
	}

	// doConfirm

	{ // app0 receives PacketPrepareResult from app1
		data := cross.NewPacketDataPrepareResult(
			txID,
			0,
			cross.PREPARE_STATUS_OK,
		)
		packet := channeltypes.NewPacket(
			data.GetBytes(),
			packetSeq, ch1to0.Port, ch1to0.Channel, ch0to1.Port, ch0to1.Channel, data.GetTimeoutHeight())
		suite.buildMsgAndDoRelay(packet, app1, app0, txID, relayer0Info, txBuilder, packetSeq)
	}

	{ // app0 receives PacketPrepareResult from app2
		data := cross.NewPacketDataPrepareResult(
			txID,
			1,
			cross.PREPARE_STATUS_OK,
		)
		packet := channeltypes.NewPacket(
			data.GetBytes(),
			packetSeq, ch2to0.Port, ch2to0.Channel, ch0to2.Port, ch0to2.Channel, data.GetTimeoutHeight())
		suite.buildMsgAndDoRelay(packet, app2, app0, txID, relayer0Info, txBuilder, packetSeq)

		ci, ok := app0.app.CrossKeeper.GetCoordinator(app0.ctx, txID)
		suite.True(ok)
		suite.Equal(cross.CO_DECISION_COMMIT, ci.Decision)

		suite.updateClient(app1, app0.chainID, app0)
	}

	packetSeq++

	// doCommit
	var (
		commitPacketTx0 channeltypes.Packet
		commitPacketTx1 channeltypes.Packet
	)

	{ // execute to commit on app1
		data := cross.NewPacketDataCommit(
			txID,
			0,
			true,
		)
		commitPacketTx0 = channeltypes.NewPacket(
			data.GetBytes(),
			packetSeq, ch0to1.Port, ch0to1.Channel, ch1to0.Port, ch1to0.Channel, data.GetTimeoutHeight())
		suite.buildMsgAndDoRelay(commitPacketTx0, app0, app1, txID, relayer0Info, txBuilder, packetSeq)
	}
	{ // execute to commit on app2
		data := cross.NewPacketDataCommit(
			txID,
			1,
			true,
		)
		commitPacketTx1 = channeltypes.NewPacket(
			data.GetBytes(),
			packetSeq, ch0to2.Port, ch0to2.Channel, ch2to0.Port, ch2to0.Channel, data.GetTimeoutHeight())
		suite.buildMsgAndDoRelay(commitPacketTx1, app0, app2, txID, relayer0Info, txBuilder, packetSeq)
	}

	// Receive an Ack packet

	{ // app1
		data := cross.NewPacketDataAckCommit(txID, 0)
		packet := channeltypes.NewPacket(
			data.GetBytes(),
			packetSeq, ch1to0.Port, ch1to0.Channel, ch0to1.Port, ch0to1.Channel, data.GetTimeoutHeight())
		suite.buildMsgAndDoRelay(packet, app1, app0, txID, relayer0Info, txBuilder, packetSeq)
		ci, ok := app0.app.CrossKeeper.GetCoordinator(app0.ctx, txID)
		suite.True(ok)
		suite.False(ci.IsReceivedALLAcks())
	}
	{ // app2
		data := cross.NewPacketDataAckCommit(txID, 1)
		packet := channeltypes.NewPacket(
			data.GetBytes(),
			packetSeq, ch2to0.Port, ch2to0.Channel, ch0to2.Port, ch0to2.Channel, data.GetTimeoutHeight())
		suite.buildMsgAndDoRelay(packet, app2, app0, txID, relayer0Info, txBuilder, packetSeq)

		ci, ok := app0.app.CrossKeeper.GetCoordinator(app0.ctx, txID)
		suite.True(ok)
		suite.True(ci.IsReceivedALLAcks())
	}
}

func (suite *ExampleTestSuite) buildMsgAndDoRelay(packet channeltypes.Packet, sender, receiver *appContext, txID cross.TxID, relayer keyring.Info, txBuilder authtypes.TxBuilder, seq uint64) {
	state, ok := receiver.app.IBCKeeper.ClientKeeper.GetClientState(receiver.ctx, sender.chainID)
	suite.True(ok)
	res := sender.app.Query(abci.RequestQuery{
		Path:   "store/ibc/key",
		Data:   ibctypes.KeyPacketCommitment(packet.GetSourcePort(), packet.GetSourceChannel(), seq),
		Height: int64(state.GetLatestHeight()),
		Prove:  true,
	})
	suite.True(res.IsOK())
	proof := commitment.MerkleProof{Proof: res.Proof}

	msg := channeltypes.NewMsgPacket(packet, proof, uint64(state.GetLatestHeight()), relayer.GetAddress())
	if err := suite.doRelay(msg, sender, receiver, relayer, txBuilder); err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *ExampleTestSuite) doRelay(msg sdk.Msg, sender, receiver *appContext, relayer keyring.Info, txBuilder authtypes.TxBuilder) error {
	var err error
	stdTx := authtypes.NewStdTx([]sdk.Msg{msg}, authtypes.StdFee{}, nil, "")
	stdTx, err = txBuilder.WithChainID(receiver.chainID).
		WithAccountNumber(3). // TODO should make it configurable?
		WithSequence(suite.getAndIncrAccountSeq(receiver.chainID, relayer.GetAddress())).
		SignStdTx(relayer.GetName(), clientkeys.DefaultKeyPass, stdTx, true)
	if err != nil {
		return err
	}
	txBytes, err := txBuilder.WithChainID(receiver.chainID).TxEncoder()(stdTx)
	if err != nil {
		return err
	}

	if res := receiver.app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes}); !res.IsOK() {
		return fmt.Errorf("failed to deliverTx: %v", res.String())
	}

	suite.nextBlock(receiver)
	suite.updateClient(sender, receiver.chainID, receiver)

	return nil
}

func (suite *ExampleTestSuite) getAccountSeq(chainID string, address sdk.AccAddress) uint64 {
	return suite.accountSeqs[chainID+"/"+address.String()]
}

func (suite *ExampleTestSuite) incrAccountSeq(chainID string, address sdk.AccAddress) {
	suite.accountSeqs[chainID+"/"+address.String()]++
}

func (suite *ExampleTestSuite) getAndIncrAccountSeq(chainID string, address sdk.AccAddress) uint64 {
	seq := suite.getAccountSeq(chainID, address)
	suite.incrAccountSeq(chainID, address)
	return seq
}

/**
FIXME: The following code comes from cross/internal/keeper_test.go. We need to consider refactoring.
**/

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

func (suite *ExampleTestSuite) createClient(actx *appContext, clientID string, dst *appContext) {
	dst.app.Commit()

	h := abci.Header{ChainID: dst.ctx.ChainID(), Height: dst.app.LastBlockHeight() + 1}
	dst.app.BeginBlock(abci.RequestBeginBlock{Header: h})
	dst.ctx = dst.app.BaseApp.NewContext(false, h)
	now := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)

	header := tendermint.CreateTestHeader(dst.chainID, dst.ctx.BlockHeight()-1, now, dst.valSet, dst.signers)
	consensusState := header.ConsensusState()
	clientState, err := tendermint.Initialize(clientID, trustingPeriod, ubdPeriod, header)
	if err != nil {
		panic(err)
	}

	_, err = actx.app.IBCKeeper.ClientKeeper.CreateClient(actx.ctx, clientState, consensusState)
	suite.NoError(err)
	suite.nextBlock(actx)
}

func (suite *ExampleTestSuite) updateClient(actx *appContext, clientID string, dst *appContext) {
	// always commit and begin a new block on updateClient
	dst.app.Commit()
	commitID := dst.app.LastCommitID()

	now := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	height := dst.app.LastBlockHeight() + 1
	header := tendermint.CreateTestHeader(dst.chainID, height, now, dst.valSet, dst.signers)
	h := header.ToABCIHeader()
	dst.app.BeginBlock(abci.RequestBeginBlock{Header: h})
	dst.ctx = dst.app.BaseApp.NewContext(false, h)

	state := tendermint.ConsensusState{
		Root: commitment.NewMerkleRoot(commitID.Hash),
	}
	preheader := tendermint.CreateTestHeader(dst.chainID, height-1, now, dst.valSet, dst.signers)
	clientState, err := tendermint.Initialize(clientID, trustingPeriod, ubdPeriod, preheader)
	if err != nil {
		panic(err)
	}
	actx.app.IBCKeeper.ClientKeeper.SetClientState(actx.ctx, clientState)
	actx.app.IBCKeeper.ClientKeeper.SetClientConsensusState(actx.ctx, clientID, uint64(height-1), state)
	suite.nextBlock(actx)
}

func (suite *ExampleTestSuite) nextBlock(actx *appContext) int64 {
	actx.app.Commit()
	h := abci.Header{ChainID: actx.ctx.ChainID(), Height: actx.app.LastBlockHeight() + 1, Time: time.Now()}
	actx.app.BeginBlock(abci.RequestBeginBlock{Header: h})
	actx.ctx = actx.app.BaseApp.NewContext(false, h)
	return h.Height
}

func (suite *ExampleTestSuite) createConnection(actx *appContext, clientID, connectionID, counterpartyClientID, counterpartyConnectionID string, state connectionexported.State) {
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

func (suite *ExampleTestSuite) createChannel(actx *appContext, portID string, chanID string, connID string, counterpartyPort string, counterpartyChan string, state channelexported.State) {
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

func (suite *ExampleTestSuite) createClients(
	srcClientID string, // clientID of dstapp
	srcapp *appContext,
	dstClientID string, // clientID of srcapp
	dstapp *appContext,
) {
	suite.createClient(srcapp, srcClientID, dstapp)
	suite.createClient(dstapp, dstClientID, srcapp)

	srcapp.app.IBCKeeper.ClientKeeper.GetClientConsensusState(srcapp.ctx, srcClientID, 1)
	dstapp.app.IBCKeeper.ClientKeeper.GetClientConsensusState(dstapp.ctx, dstClientID, 1)
}

func (suite *ExampleTestSuite) createConnections(
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

func (suite *ExampleTestSuite) createChannels(
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

func (suite *ExampleTestSuite) openChannels(
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

func (suite *ExampleTestSuite) createApp(
	chainID string,
	contractHanderProvider simapp.ContractHandlerProvider,
	anteHandlerProvider simapp.AnteHandlerProvider,
	genAccs []authexported.GenesisAccount,
	balances ...bank.Balance,
) *appContext {
	return suite.createAppWithHeader(abci.Header{ChainID: chainID, Time: time.Now()}, contractHanderProvider, anteHandlerProvider, genAccs, balances...)
}

func (suite *ExampleTestSuite) createAppWithHeader(
	header abci.Header,
	contractHanderProvider simapp.ContractHandlerProvider,
	anteHandlerProvider simapp.AnteHandlerProvider,
	genAccs []authexported.GenesisAccount,
	balances ...bank.Balance,
) *appContext {
	isCheckTx := false
	app := simapp.SetupWithGenesisAccounts(header.ChainID, contractHanderProvider, anteHandlerProvider, genAccs, balances...)
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
		actx.ctx = actx.app.BaseApp.NewContext(false, abci.Header{ChainID: actx.ctx.ChainID(), Time: time.Now()})
	}
}
