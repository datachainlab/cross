package keeper_test

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/datachainlab/cross/x/ibc/contract"
	"github.com/datachainlab/cross/x/ibc/cross"
	lock "github.com/datachainlab/cross/x/ibc/store/lock"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// define constants used for testing
const (
	testClientType     = clientexported.Tendermint
	testChannelOrder   = channelexported.UNORDERED
	testChannelVersion = "1.0"
)

func (suite *KeeperTestSuite) createClient(actx *appContext, clientID string) {
	actx.app.Commit()
	commitID := actx.app.LastCommitID()

	actx.app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: actx.app.LastBlockHeight() + 1}})
	actx.ctx = actx.app.BaseApp.NewContext(false, abci.Header{})

	consensusState := tendermint.ConsensusState{
		Root:             commitment.NewRoot(commitID.Hash),
		ValidatorSetHash: actx.valSet.Hash(),
	}

	_, err := actx.app.IBCKeeper.ClientKeeper.CreateClient(actx.ctx, clientID, testClientType, consensusState)
	suite.NoError(err)
}

func (suite *KeeperTestSuite) updateClient(actx *appContext, clientID string) {
	// always commit and begin a new block on updateClient
	actx.app.Commit()
	commitID := actx.app.LastCommitID()

	actx.app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: actx.app.LastBlockHeight() + 1}})
	actx.ctx = actx.app.BaseApp.NewContext(false, abci.Header{})

	state := tendermint.ConsensusState{
		Root: commitment.NewRoot(commitID.Hash),
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

func (suite *KeeperTestSuite) createContractHandler(stk sdk.StoreKey, cid string) cross.ContractHandler {
	contractHandler := contract.NewContractHandler(contract.NewKeeper(stk), func(kvs sdk.KVStore) cross.State {
		return lock.NewStore(kvs)
	})
	c := contract.NewContract([]contract.Method{
		{
			Name: "issue",
			F: func(ctx contract.Context, store cross.Store) error {
				coin, err := parseCoin(ctx, 0, 1)
				if err != nil {
					return err
				}
				balance := getBalanceOf(store, ctx.Signers()[0])
				balance = balance.Add(coin)
				setBalance(store, ctx.Signers()[0], balance)
				return nil
			},
		},
		{
			Name: "test-balance",
			F: func(ctx contract.Context, store cross.Store) error {
				coin, err := parseCoin(ctx, 0, 1)
				if err != nil {
					return err
				}
				balance := getBalanceOf(store, ctx.Signers()[0])
				if !balance.AmountOf(coin.Denom).Equal(coin.Amount) {
					return errors.New("amount is unexpected")
				}
				return nil
			},
		},
		{
			Name: "test-not-issued",
			F: func(ctx contract.Context, store cross.Store) error {
				balance := getBalanceOf(store, ctx.Signers()[0])
				if len(balance) == 0 {
					return nil
				} else {
					return errors.New("maybe coin is already issued")
				}
			},
		},
	})
	contractHandler.AddRoute(cid, c)
	return contractHandler
}

func (suite *KeeperTestSuite) TestSendInitiate() {
	lock.RegisterCodec(cross.ModuleCdc)

	initiator := sdk.AccAddress("initiator")

	app0 := suite.createApp("app0") // coordinator node

	app1 := suite.createApp("app1")
	signer1 := sdk.AccAddress("signer1")
	ci1 := contract.NewContractInfo("c1", "issue", [][]byte{[]byte("tone"), []byte("80")})
	chd1 := suite.createContractHandler(app1.app.GetKey(cross.StoreKey), "c1")

	app2 := suite.createApp("app2")
	signer2 := sdk.AccAddress("signer2")
	ci2 := contract.NewContractInfo("c2", "issue", [][]byte{[]byte("ttwo"), []byte("60")})
	chd2 := suite.createContractHandler(app2.app.GetKey(cross.StoreKey), "c2")

	ch0to1 := cross.NewChannelInfo("testportzeroone", "testchannelzeroone") // app0 -> app1
	ch1to0 := cross.NewChannelInfo("testportonezero", "testchannelonezero") // app1 -> app0
	ch0to2 := cross.NewChannelInfo("testportzerotwo", "testchannelzerotwo") // app0 -> app2
	ch2to0 := cross.NewChannelInfo("testporttwozero", "testchanneltwozero") // app2 -> app0

	var err error
	var nonce uint64 = 1
	var tss = []cross.ContractTransaction{
		cross.NewContractTransaction(
			ch0to1,
			[]sdk.AccAddress{signer1},
			ci1.Bytes(),
			[]cross.OP{lock.Write{K: signer1, V: marshalCoin(sdk.Coins{sdk.NewInt64Coin("tone", 80)})}},
		),
		cross.NewContractTransaction(
			ch0to2,
			[]sdk.AccAddress{signer2},
			ci2.Bytes(),
			[]cross.OP{lock.Write{K: signer2, V: marshalCoin(sdk.Coins{sdk.NewInt64Coin("ttwo", 60)})}},
		),
	}

	msg := cross.NewMsgInitiate(
		initiator,
		tss,
		nonce,
	)
	txID := msg.GetTxID()

	err = app0.app.CrossKeeper.MulticastPreparePacket(
		app0.ctx,
		initiator,
		msg,
		msg.ContractTransactions,
	)
	suite.Error(err) // channel does not exist

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

	err = app0.app.CrossKeeper.MulticastPreparePacket(
		app0.ctx,
		initiator,
		msg,
		msg.ContractTransactions,
	)
	suite.NoError(err) // successfully executed

	ci, found := app0.app.CrossKeeper.GetCoordinator(app0.ctx, msg.GetTxID())
	if suite.True(found) {
		suite.Equal(ci.Status, cross.CO_STATUS_INIT)
	}

	nextSeqSend := uint64(1)
	packetCommitment := app0.app.IBCKeeper.ChannelKeeper.GetPacketCommitment(app0.ctx, ch0to1.Port, ch0to1.Channel, nextSeqSend)
	suite.NotNil(packetCommitment)
	packetCommitment = app0.app.IBCKeeper.ChannelKeeper.GetPacketCommitment(app0.ctx, ch0to2.Port, ch0to2.Channel, nextSeqSend)
	suite.NotNil(packetCommitment)

	suite.testPreparePacket(app1, ch1to0, ch0to1, txID, 0, chd1, tss[0], nextSeqSend)
	suite.testPreparePacket(app2, ch2to0, ch0to2, txID, 1, chd2, tss[1], nextSeqSend)

	// Tests for Confirm step

	nextSeqSend += 1
	srcs := [2]cross.ChannelInfo{
		ch0to1,
		ch0to2,
	}
	dsts := [2]cross.ChannelInfo{
		ch1to0,
		ch2to0,
	}

	var makePrepareDataPacket = func(sender sdk.AccAddress, txID []byte, transID int, status uint8) channel.MsgPacket {
		return channel.MsgPacket{Packet: channel.Packet{Data: cross.NewPacketDataPrepareResult(sender, txID, transID, status)}}
	}

	relayer := sdk.AccAddress("relayer1")
	// ensure that coordinator decides 'abort'
	{
		pps := []cross.PreparePacket{}
		p1 := cross.NewPreparePacket(makePrepareDataPacket(relayer, txID, 0, cross.PREPARE_STATUS_OK), ch0to1)
		p2 := cross.NewPreparePacket(makePrepareDataPacket(relayer, txID, 1, cross.PREPARE_STATUS_FAILED), ch0to2)
		pps = append(pps, p1, p2)

		capp, _ := app0.Cache()
		suite.testConfirmMsg(&capp, pps, srcs, dsts, txID, nextSeqSend, false)
	}
	// ensure that coordinator decides 'abort'
	{
		pps := []cross.PreparePacket{}
		p1 := cross.NewPreparePacket(makePrepareDataPacket(relayer, txID, 0, cross.PREPARE_STATUS_FAILED), ch0to1)
		p2 := cross.NewPreparePacket(makePrepareDataPacket(relayer, txID, 1, cross.PREPARE_STATUS_FAILED), ch0to2)
		pps = append(pps, p1, p2)

		capp, _ := app0.Cache()
		suite.testConfirmMsg(&capp, pps, srcs, dsts, txID, nextSeqSend, false)
	}
	// ensure that coordinator decides 'abort'
	{
		pps := []cross.PreparePacket{}
		p1 := cross.NewPreparePacket(makePrepareDataPacket(relayer, txID, 0, cross.PREPARE_STATUS_FAILED), ch0to1)
		p2 := cross.NewPreparePacket(makePrepareDataPacket(relayer, txID, 1, cross.PREPARE_STATUS_OK), ch0to2)
		pps = append(pps, p1, p2)

		capp, _ := app0.Cache()
		suite.testConfirmMsg(&capp, pps, srcs, dsts, txID, nextSeqSend, false)
	}
	// ensure that coordinator decides nothing
	{
		pps := []cross.PreparePacket{}
		p1 := cross.NewPreparePacket(makePrepareDataPacket(relayer, txID, 0, cross.PREPARE_STATUS_OK), ch0to1)
		pps = append(pps, p1)

		capp, _ := app0.Cache()
		suite.testConfirmMsg(&capp, pps, srcs, dsts, txID, nextSeqSend, true)
	}
	// ensure that transactionID conflict occurs
	{
		pps := []cross.PreparePacket{}
		p1 := cross.NewPreparePacket(makePrepareDataPacket(relayer, txID, 0, cross.PREPARE_STATUS_OK), ch0to1)
		p2 := cross.NewPreparePacket(makePrepareDataPacket(relayer, txID, 0, cross.PREPARE_STATUS_OK), ch0to2)
		pps = append(pps, p1, p2)

		capp, _ := app0.Cache()
		suite.testConfirmMsg(&capp, pps, srcs, dsts, txID, nextSeqSend, true)
	}
	// invalid transactionID
	{
		pps := []cross.PreparePacket{}
		p1 := cross.NewPreparePacket(makePrepareDataPacket(relayer, txID, 0, cross.PREPARE_STATUS_OK), ch0to1)
		p2 := cross.NewPreparePacket(makePrepareDataPacket(relayer, txID, 2, cross.PREPARE_STATUS_OK), ch0to2)
		pps = append(pps, p1, p2)

		capp, _ := app0.Cache()
		suite.testConfirmMsg(&capp, pps, srcs, dsts, txID, nextSeqSend, true)
	}
	// includes all expected packets, but also include invalid transactionID
	{
		pps := []cross.PreparePacket{}
		p1 := cross.NewPreparePacket(makePrepareDataPacket(relayer, txID, 0, cross.PREPARE_STATUS_OK), ch0to1)
		p2 := cross.NewPreparePacket(makePrepareDataPacket(relayer, txID, 1, cross.PREPARE_STATUS_OK), ch0to2)
		p3 := cross.NewPreparePacket(makePrepareDataPacket(relayer, txID, 2, cross.PREPARE_STATUS_OK), ch0to2)
		pps = append(pps, p1, p2, p3)

		capp, _ := app0.Cache()
		suite.testConfirmMsg(&capp, pps, srcs, dsts, txID, nextSeqSend, true)
	}
	// ensure that coordinator decides 'commit'
	{
		pps := []cross.PreparePacket{}
		p1 := cross.NewPreparePacket(makePrepareDataPacket(relayer, txID, 0, cross.PREPARE_STATUS_OK), ch0to1)
		p2 := cross.NewPreparePacket(makePrepareDataPacket(relayer, txID, 1, cross.PREPARE_STATUS_OK), ch0to2)
		pps = append(pps, p1, p2)

		capp, writer := app0.Cache()
		suite.testConfirmMsg(&capp, pps, srcs, dsts, txID, nextSeqSend, false)
		ci, found := capp.app.CrossKeeper.GetCoordinator(capp.ctx, msg.GetTxID())
		if suite.True(found) {
			suite.Equal(ci.Status, cross.CO_STATUS_COMMIT)
		}
		writer()
	}

	// ensure that each corhorts commit or abort
	{
		relayer := sdk.AccAddress("relayer2")
		// In a1, execute to commit
		{
			capp, _ := app1.Cache()
			suite.testCommitPacket(&capp, chd1, ch0to1, ch1to0, cross.NewPacketDataCommit(relayer, txID, true), signer1)
		}

		// In a2, execute to commit
		{
			capp, _ := app2.Cache()
			suite.testCommitPacket(&capp, chd2, ch0to2, ch2to0, cross.NewPacketDataCommit(relayer, txID, true), signer2)
		}

		// In a1, execute to abort
		{
			capp, _ := app1.Cache()
			suite.testAbortPacket(&capp, chd1, ch0to1, ch1to0, cross.NewPacketDataCommit(relayer, txID, false), signer1)
		}

		// In a2, execute to abort
		{
			capp, _ := app2.Cache()
			suite.testAbortPacket(&capp, chd2, ch0to2, ch2to0, cross.NewPacketDataCommit(relayer, txID, false), signer2)
		}
	}
}

func (suite *KeeperTestSuite) testCommitPacket(actx *appContext, contractHandler cross.ContractHandler, src, dst cross.ChannelInfo, packet cross.PacketDataCommit, txSigner sdk.AccAddress) {
	err := actx.app.CrossKeeper.ReceiveCommitPacket(actx.ctx, contractHandler, src.Port, src.Channel, dst.Port, dst.Channel, packet)
	if !suite.NoError(err) {
		return
	}
	tx, found := actx.app.CrossKeeper.GetTx(actx.ctx, packet.TxID)
	if !suite.True(found) {
		return
	}
	suite.Equal(cross.TX_STATUS_COMMIT, tx.Status)
	// ensure that the state is expected
	_, err = contractHandler.GetState(actx.ctx, tx.Contract)
	if !suite.NoError(err) {
		return
	}
	ci, err := contract.DecodeContractSignature(tx.Contract)
	if !suite.NoError(err) {
		return
	}
	contractInfo := contract.NewContractInfo(ci.ID, "test-balance", [][]byte{
		ci.Args[0],
		ci.Args[1],
	})
	bz, err := contract.EncodeContractSignature(contractInfo)
	if !suite.NoError(err) {
		return
	}
	ctx := cross.WithSigners(actx.ctx, []sdk.AccAddress{txSigner})
	_, err = contractHandler.Handle(ctx, bz)
	suite.NoError(err)
}

func (suite *KeeperTestSuite) testAbortPacket(actx *appContext, contractHandler cross.ContractHandler, src, dst cross.ChannelInfo, packet cross.PacketDataCommit, txSigner sdk.AccAddress) {
	err := actx.app.CrossKeeper.ReceiveCommitPacket(actx.ctx, contractHandler, src.Port, src.Channel, dst.Port, dst.Channel, packet)
	if !suite.NoError(err) {
		return
	}
	tx, found := actx.app.CrossKeeper.GetTx(actx.ctx, packet.TxID)
	if !suite.True(found) {
		return
	}
	suite.Equal(cross.TX_STATUS_ABORT, tx.Status)

	ci, err := contract.DecodeContractSignature(tx.Contract)
	if !suite.NoError(err) {
		return
	}
	contractInfo := contract.NewContractInfo(ci.ID, "test-not-issued", [][]byte{})
	bz, err := contract.EncodeContractSignature(contractInfo)
	if !suite.NoError(err) {
		return
	}
	ctx := cross.WithSigners(actx.ctx, []sdk.AccAddress{txSigner})
	_, err = contractHandler.Handle(ctx, bz)
	suite.NoError(err)
}

func (suite *KeeperTestSuite) testConfirmMsg(actx *appContext, pps []cross.PreparePacket, srcs, dsts [2]cross.ChannelInfo, txID []byte, nextseq uint64, hasMulticastError bool) {
	relayer := sdk.AccAddress("relayer2")
	msgConfirm := cross.NewMsgConfirm(txID, pps, relayer)
	isCommit := msgConfirm.IsCommittable()
	err := actx.app.CrossKeeper.MulticastCommitPacket(actx.ctx, txID, pps, relayer, isCommit)
	if hasMulticastError {
		suite.Error(err, err)
		return
	} else {
		suite.NoError(err, err)
	}
	for i, src := range srcs {
		dst := dsts[i]

		newNextSeqSend, found := actx.app.IBCKeeper.ChannelKeeper.GetNextSequenceSend(actx.ctx, src.Port, src.Channel)
		suite.True(found)
		suite.Equal(nextseq+1, newNextSeqSend)

		packetCommitment := actx.app.IBCKeeper.ChannelKeeper.GetPacketCommitment(actx.ctx, src.Port, src.Channel, nextseq)
		suite.NotNil(packetCommitment)

		// ensure that commit packet exists in store
		expectedPacket1 := actx.app.CrossKeeper.CreateCommitPacket(
			actx.ctx,
			nextseq,
			src.Port,
			src.Channel,
			dst.Port,
			dst.Channel,
			relayer,
			txID,
			isCommit,
		)
		suite.Equal(
			packetCommitment,
			channeltypes.CommitPacket(
				expectedPacket1.Data,
			),
		)
	}
}

func (suite *KeeperTestSuite) testPreparePacket(actx *appContext, src, dst cross.ChannelInfo, txID []byte, tsID int, contractHandler cross.ContractHandler, ts cross.ContractTransaction, nextseq uint64) {
	var err error

	relayer := sdk.AccAddress("relayer1")
	packetData := cross.NewPacketDataPrepare(relayer, txID, tsID, ts)
	ctx, writer := actx.ctx.CacheContext()
	ctx = cross.WithSigners(ctx, ts.Signers)
	err = actx.app.CrossKeeper.PrepareTransaction(
		ctx,
		contractHandler,
		dst.Port,
		dst.Channel,
		src.Port,
		src.Channel,
		packetData,
		relayer,
	)
	suite.NoError(err)
	tx, ok := actx.app.CrossKeeper.GetTx(ctx, txID)
	if suite.True(ok) {
		suite.Equal(cross.TX_STATUS_PREPARE, tx.Status)
	}
	newNextSeqSend, found := actx.app.IBCKeeper.ChannelKeeper.GetNextSequenceSend(ctx, src.Port, src.Channel)
	suite.True(found)
	suite.Equal(nextseq+1, newNextSeqSend)

	packetCommitment := actx.app.IBCKeeper.ChannelKeeper.GetPacketCommitment(ctx, src.Port, src.Channel, nextseq)
	suite.NotNil(packetCommitment)

	suite.Equal(
		packetCommitment,
		channeltypes.CommitPacket(
			actx.app.CrossKeeper.CreatePrepareResultPacket(
				nextseq,
				src.Port,
				src.Channel,
				dst.Port,
				dst.Channel,
				relayer,
				txID,
				tsID,
				cross.PREPARE_STATUS_OK,
			).Data,
		),
	)
	writer()
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
