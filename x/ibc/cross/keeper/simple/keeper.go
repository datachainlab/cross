package simple

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	"github.com/datachainlab/cross/x/ibc/cross/keeper/common"
	"github.com/datachainlab/cross/x/ibc/cross/types"
	"github.com/datachainlab/cross/x/ibc/cross/types/simple"
	simpletypes "github.com/datachainlab/cross/x/ibc/cross/types/simple"
	"github.com/tendermint/tendermint/libs/log"
)

const (
	TypeName                = "simple"
	CoordinatorConnectionID = ""
)

const (
	TX_INDEX_COORDINATOR types.TxIndex = 0
	TX_INDEX_PARTICIPANT types.TxIndex = 1
)

type Keeper struct {
	cdc      *codec.Codec // The wire codec for binary encoding/decoding.
	storeKey sdk.StoreKey // Unexposed key to access store from sdk.Context

	common.Keeper
}

func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, ck common.Keeper) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		Keeper:   ck,
	}
}

// caller is Coordinator
func (k Keeper) SendCall(
	ctx sdk.Context,
	packetSender types.PacketSender,
	contractHandler types.ContractHandler,
	msg types.MsgInitiate,
	transactions []types.ContractTransaction,
) (types.TxID, error) {
	if ctx.ChainID() != msg.ChainID {
		return types.TxID{}, fmt.Errorf("unexpected chainID: '%v' != '%v'", ctx.ChainID(), msg.ChainID)
	} else if ctx.BlockHeight() >= msg.TimeoutHeight {
		return types.TxID{}, fmt.Errorf("this msg is already timeout: current=%v timeout=%v", ctx.BlockHeight(), msg.TimeoutHeight)
	}

	txID := common.MakeTxID(ctx, msg)
	if _, ok := k.GetCoordinator(ctx, txID); ok {
		return types.TxID{}, fmt.Errorf("coordinator '%x' already exists", txID)
	}

	lkr, err := types.MakeLinker(transactions)
	if err != nil {
		return types.TxID{}, err
	}

	tx0 := transactions[TX_INDEX_COORDINATOR]
	tx1 := transactions[TX_INDEX_PARTICIPANT]

	if !k.ChannelResolver().Capabilities().CrossChainCalls() && (len(tx0.Links) > 0 || len(tx1.Links) > 0) {
		return types.TxID{}, errors.New("this channelResolver cannot resolve cannot support the cross-chain calls feature")
	}

	objs0, err := lkr.Resolve(tx0.Links)
	if err != nil {
		return types.TxID{}, err
	}
	if err := k.PrepareTransaction(ctx, contractHandler, txID, TX_INDEX_COORDINATOR, tx0, objs0); err != nil {
		return types.TxID{}, err
	}

	objs1, err := lkr.Resolve(tx1.Links)
	if err != nil {
		return types.TxID{}, err
	}

	ch0, err := k.ChannelResolver().Resolve(ctx, tx0.ChainID)
	if err != nil {
		return types.TxID{}, err
	}
	ch1, err := k.ChannelResolver().Resolve(ctx, tx1.ChainID)
	if err != nil {
		return types.TxID{}, err
	}

	c, found := k.ChannelKeeper().GetChannel(ctx, ch1.Port, ch1.Channel)
	if !found {
		return types.TxID{}, sdkerrors.Wrap(channel.ErrChannelNotFound, ch1.Channel)
	}

	data := simpletypes.NewPacketDataCall(msg.Sender, txID, types.NewContractTransactionInfo(tx1, objs1))
	if err := k.SendPacket(
		ctx,
		packetSender,
		data,
		ch1.Port, ch1.Channel,
		c.Counterparty.PortID, c.Counterparty.ChannelID,
		data.GetTimeoutHeight(),
		data.GetTimeoutTimestamp(),
	); err != nil {
		return types.TxID{}, err
	}
	hops := c.GetConnectionHops()
	co := types.NewCoordinatorInfo(
		types.CO_STATUS_INIT,
		[]string{CoordinatorConnectionID, hops[len(hops)-1]},
		[]types.ChannelInfo{*ch0, *ch1},
	)
	if err := co.Confirm(TX_INDEX_COORDINATOR, CoordinatorConnectionID); err != nil {
		return types.TxID{}, err
	}
	k.SetCoordinator(
		ctx,
		txID,
		co,
	)
	txinfo := types.NewTxInfo(types.TX_STATUS_PREPARE, types.PREPARE_RESULT_OK, CoordinatorConnectionID, tx0.CallInfo)
	k.SetTx(ctx, txID, TX_INDEX_COORDINATOR, txinfo)
	return txID, nil
}

// caller is participant
func (k Keeper) ReceiveCallPacket(
	ctx sdk.Context,
	contractHandler types.ContractHandler,
	sourcePort,
	sourceChannel string,
	data simple.PacketDataCall,
) (uint8, error) {
	if _, ok := k.GetTx(ctx, data.TxID, TX_INDEX_PARTICIPANT); ok {
		return 0, fmt.Errorf("txID '%x' already exists", data.TxID)
	}

	if !k.ChannelResolver().Capabilities().CrossChainCalls() && len(data.TxInfo.LinkObjects) > 0 {
		return 0, errors.New("this channelResolver cannot resolve cannot support the cross-chain calls feature")
	}

	result := types.PREPARE_RESULT_OK
	if err := k.CommitImmediatelyTransaction(ctx, contractHandler, data.TxID, 1, data.TxInfo.Transaction, data.TxInfo.LinkObjects); err != nil {
		result = types.PREPARE_RESULT_FAILED
		k.Logger(ctx).Info("failed to CommitImmediatelyTransaction", "err", err)
	}

	c, found := k.ChannelKeeper().GetChannel(ctx, sourcePort, sourceChannel)
	if !found {
		return 0, fmt.Errorf("channel(port=%v channel=%v) not found", sourcePort, sourceChannel)
	}
	hops := c.GetConnectionHops()
	connID := hops[len(hops)-1]

	txinfo := types.NewTxInfo(types.TX_STATUS_COMMIT, types.PREPARE_RESULT_OK, connID, data.TxInfo.Transaction.CallInfo)
	k.SetTx(ctx, data.TxID, TX_INDEX_PARTICIPANT, txinfo)
	return result, nil
}

// caller is coordinator
func (k Keeper) ReceiveCallAcknowledgement(
	ctx sdk.Context,
	sourcePort string,
	sourceChannel string,
	ack simpletypes.PacketCallAcknowledgement,
	txID types.TxID,
) (isCommittable bool, err error) {
	co, ok := k.GetCoordinator(ctx, txID)
	if !ok {
		return false, fmt.Errorf("coordinator '%x' not found", txID)
	} else if co.Status == types.CO_STATUS_NONE {
		return false, errors.New("coordinator status must not be CO_STATUS_NONE")
	} else if co.IsCompleted() {
		return false, errors.New("all transactions are already confirmed")
	}

	c, found := k.ChannelKeeper().GetChannel(ctx, sourcePort, sourceChannel)
	if !found {
		return false, sdkerrors.Wrap(channel.ErrChannelNotFound, sourceChannel)
	}
	hops := c.GetConnectionHops()
	if err := co.Confirm(TX_INDEX_PARTICIPANT, hops[len(hops)-1]); err != nil {
		return false, err
	}
	switch ack.Status {
	case simple.COMMIT_OK:
		co.Decision = types.CO_DECISION_COMMIT
	case simple.COMMIT_FAILED:
		co.Decision = types.CO_DECISION_ABORT
	default:
		panic("unreachable")
	}
	co.Status = types.CO_STATUS_DECIDED
	co.AddAck(TX_INDEX_COORDINATOR)
	co.AddAck(TX_INDEX_PARTICIPANT)
	if !co.IsCompleted() || !co.IsReceivedALLAcks() {
		panic("fatal error")
	}
	k.SetCoordinator(ctx, txID, *co)
	return true, nil
}

func (k Keeper) TryCommit(
	ctx sdk.Context,
	contractHandler types.ContractHandler,
	txID types.TxID,
	isCommittable bool,
) (types.ContractHandlerResult, error) {
	tx, err := k.EnsureTxStatus(ctx, txID, TX_INDEX_COORDINATOR, types.TX_STATUS_PREPARE)
	if err != nil {
		return nil, err
	}

	var status uint8
	var res types.ContractHandlerResult
	if isCommittable {
		res, err = k.CommitTransaction(ctx, contractHandler, txID, TX_INDEX_COORDINATOR, tx)
		if err != nil {
			return nil, err
		}
		status = types.TX_STATUS_COMMIT
	} else {
		err = k.AbortTransaction(ctx, contractHandler, txID, TX_INDEX_COORDINATOR, tx)
		if err != nil {
			return nil, err
		}
		status = types.TX_STATUS_ABORT
	}

	k.UpdateTxStatus(ctx, txID, TX_INDEX_COORDINATOR, status)
	return res, nil
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("cross/%s", TypeName))
}
