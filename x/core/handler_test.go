package core_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"

	samplemodtypes "github.com/datachainlab/cross/simapp/samplemod/types"
	accounttypes "github.com/datachainlab/cross/x/account/types"
	crosshost "github.com/datachainlab/cross/x/core/host"
	initiatortypes "github.com/datachainlab/cross/x/core/initiator/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	ibctesting "github.com/datachainlab/cross/x/ibc/testing"
	xcctypes "github.com/datachainlab/cross/x/xcc/types"
)

type CrossTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain

	queryClient transfertypes.QueryClient
}

func (suite *CrossTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))

	queryHelper := baseapp.NewQueryServerTestHelper(suite.chainA.GetContext(), suite.chainA.App.InterfaceRegistry())
	transfertypes.RegisterQueryServer(queryHelper, suite.chainA.App.TransferKeeper)
	suite.queryClient = transfertypes.NewQueryClient(queryHelper)
}

func (suite *CrossTestSuite) TestHandleMsgInitiate() {
	// setup

	clientA, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, ibctesting.Tendermint)
	suite.chainB.CreatePortCapability(crosshost.PortID)
	channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, crosshost.PortID, crosshost.PortID, channeltypes.UNORDERED)

	chAB := xcctypes.ChannelInfo{Port: channelA.PortID, Channel: channelA.ID}
	xccB, err := xcctypes.PackCrossChainChannel(&chAB)
	suite.Require().NoError(err)
	chBA := xcctypes.ChannelInfo{Port: channelB.PortID, Channel: channelB.ID}
	xccA, err := xcctypes.PackCrossChainChannel(&chBA)
	suite.Require().NoError(err)

	xccSelf, err := xcctypes.PackCrossChainChannel(suite.chainA.App.XCCResolver.GetSelfCrossChainChannel(suite.chainA.GetContext()))
	suite.Require().NoError(err)

	var txID txtypes.TxID

	// Send a MsgInitiateTx to chainA
	{
		msg0 := initiatortypes.NewMsgInitiateTx(
			suite.chainA.SenderAccount.GetAddress().Bytes(),
			suite.chainA.ChainID,
			0,
			txtypes.COMMIT_PROTOCOL_SIMPLE,
			[]initiatortypes.ContractTransaction{
				{
					CrossChainChannel: xccSelf,
					Signers: []accounttypes.AccountID{
						accounttypes.AccountID(suite.chainA.SenderAccount.GetAddress()),
					},
					CallInfo: samplemodtypes.NewContractCallRequest("counter").ContractCallInfo(suite.chainA.App.AppCodec()),
				},
				{
					CrossChainChannel: xccB,
					Signers: []accounttypes.AccountID{
						accounttypes.AccountID(suite.chainB.SenderAccount.GetAddress()),
					},
					CallInfo: samplemodtypes.NewContractCallRequest("counter").ContractCallInfo(suite.chainB.App.AppCodec()),
				},
			},
			clienttypes.NewHeight(0, uint64(suite.chainA.CurrentHeader.Height)+100),
			0,
		)
		res0, err := sendMsgs(suite.coordinator, suite.chainA, suite.chainB, clientB, msg0)
		suite.Require().NoError(err)
		suite.chainA.NextBlock()

		var txMsgData sdk.TxMsgData
		var initiateTxRes initiatortypes.MsgInitiateTxResponse
		suite.Require().NoError(proto.Unmarshal(res0.Data, &txMsgData))
		suite.Require().NoError(proto.Unmarshal(txMsgData.Data[0].Data, &initiateTxRes))
		suite.Require().Equal(initiatortypes.INITIATE_TX_STATUS_PENDING, initiateTxRes.Status)
		txID = initiateTxRes.TxID
	}

	// Send a MsgIBCSignTx to chainB & receive the MsgIBCSignTx to run the transaction on chainA
	var packetCall *channeltypes.Packet
	{
		msg1 := initiatortypes.MsgIBCSignTx{
			CrossChainChannel: xccA,
			TxID:              txID,
			Signers:           []accounttypes.AccountID{suite.chainB.SenderAccount.GetAddress().Bytes()},
			TimeoutHeight:     clienttypes.NewHeight(0, uint64(suite.chainB.CurrentHeader.Height)+100),
			TimeoutTimestamp:  0,
		}
		res1, err := sendMsgs(suite.coordinator, suite.chainB, suite.chainA, clientA, &msg1)
		suite.Require().NoError(err)
		suite.chainB.NextBlock()

		p, err := getPacketFromResult(res1)
		suite.Require().NoError(err)

		res2, err := recvPacket(
			suite.coordinator, suite.chainB, suite.chainA, clientB, *p,
		)
		suite.Require().NoError(err)
		suite.chainA.NextBlock()
		packetCall, err = getPacketFromResult(res2)
		suite.Require().NoError(err)

		ack, err := getACKFromResult(res2)
		suite.Require().NoError(err)
		_, err = acknowledgePacket(
			suite.coordinator,
			suite.chainB,
			suite.chainA,
			clientA,
			*p,
			ack,
		)
		suite.Require().NoError(err)
		suite.chainB.NextBlock()
	}

	// Send a PacketDataCall to chainB
	{
		suite.Require().NoError(
			suite.coordinator.UpdateClient(suite.chainB, suite.chainA, clientB, ibctesting.Tendermint),
		)
		res, err := recvPacket(
			suite.coordinator, suite.chainA, suite.chainB, clientA, *packetCall,
		)
		suite.Require().NoError(err)
		suite.chainB.NextBlock()

		ack, err := getACKFromResult(res)
		suite.Require().NoError(err)
		_, err = acknowledgePacket(
			suite.coordinator,
			suite.chainA,
			suite.chainB,
			clientB,
			*packetCall,
			ack,
		)
		suite.Require().NoError(err)
		suite.chainB.NextBlock()
	}
}

func sendMsgs(coord *ibctesting.Coordinator, source, counterparty *ibctesting.TestChain, counterpartyClientID string, msgs ...sdk.Msg) (*sdk.Result, error) {
	res, err := source.SendMsgs(msgs...)
	if err != nil {
		return nil, err
	}

	coord.IncrementTime()
	err = coord.UpdateClient(
		counterparty, source,
		counterpartyClientID, ibctesting.Tendermint,
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func recvPacket(coord *ibctesting.Coordinator, source, counterparty *ibctesting.TestChain, sourceClient string, packet channeltypes.Packet) (*sdk.Result, error) {
	// get proof of packet commitment on source
	packetKey := host.KeyPacketCommitment(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	proof, proofHeight := source.QueryProof(packetKey)

	recvMsg := channeltypes.NewMsgRecvPacket(packet, proof, proofHeight, counterparty.SenderAccount.GetAddress())

	// receive on counterparty and update source client
	return sendMsgs(coord, counterparty, source, sourceClient, recvMsg)
}

func acknowledgePacket(coord *ibctesting.Coordinator,
	source, counterparty *ibctesting.TestChain,
	counterpartyClient string,
	packet channeltypes.Packet, ack []byte,
) (*sdk.Result, error) {
	packetKey := host.KeyPacketAcknowledgement(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	proof, proofHeight := counterparty.QueryProof(packetKey)

	ackMsg := channeltypes.NewMsgAcknowledgement(packet, ack, proof, proofHeight, source.SenderAccount.GetAddress())
	return sendMsgs(coord, source, counterparty, counterpartyClient, ackMsg)
}

func getPacketFromResult(res *sdk.Result) (*channeltypes.Packet, error) {
	var packet channeltypes.Packet

	events := sdk.StringifyEvents(res.GetEvents().ToABCIEvents())
	for _, ev := range events {
		if ev.Type == channeltypes.EventTypeSendPacket {
			for _, attr := range ev.Attributes {
				switch attr.Key {
				case channeltypes.AttributeKeyData:
					packet.Data = []byte(attr.Value)
				case channeltypes.AttributeKeyTimeoutHeight:
					parts := strings.Split(attr.Value, "-")
					packet.TimeoutHeight = clienttypes.NewHeight(
						strToUint64(parts[0]),
						strToUint64(parts[1]),
					)
				case channeltypes.AttributeKeyTimeoutTimestamp:
					packet.TimeoutTimestamp = strToUint64(attr.Value)
				case channeltypes.AttributeKeySequence:
					packet.Sequence = strToUint64(attr.Value)
				case channeltypes.AttributeKeySrcPort:
					packet.SourcePort = attr.Value
				case channeltypes.AttributeKeySrcChannel:
					packet.SourceChannel = attr.Value
				case channeltypes.AttributeKeyDstPort:
					packet.DestinationPort = attr.Value
				case channeltypes.AttributeKeyDstChannel:
					packet.DestinationChannel = attr.Value
				}
			}
		}
	}
	if err := packet.ValidateBasic(); err != nil {
		return nil, err
	}
	return &packet, nil
}

func getACKFromResult(res *sdk.Result) ([]byte, error) {
	events := sdk.StringifyEvents(res.GetEvents().ToABCIEvents())
	for _, ev := range events {
		if ev.Type == channeltypes.EventTypeWriteAck {
			for _, attr := range ev.Attributes {
				switch attr.Key {
				case channeltypes.AttributeKeyAck:
					return []byte(attr.Value), nil
				}
			}
		}
	}

	return nil, errors.New("ack not found")
}

func strToUint64(s string) uint64 {
	v, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return uint64(v)
}

func TestCrossTestSuite(t *testing.T) {
	suite.Run(t, new(CrossTestSuite))
}
