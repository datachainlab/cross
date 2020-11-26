package keeper

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	"github.com/tendermint/tendermint/libs/log"

	basekeeper "github.com/datachainlab/cross/x/core/atomic/protocol/base/keeper"
	"github.com/datachainlab/cross/x/core/atomic/protocol/simple/types"
	commontypes "github.com/datachainlab/cross/x/core/atomic/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
	"github.com/datachainlab/cross/x/packets"
)

const (
	TxIndexCoordinator txtypes.TxIndex = 0
	TxIndexParticipant txtypes.TxIndex = 1
)

const (
	TypeName = "simple"
)

type Keeper struct {
	cdc codec.Marshaler

	cm          txtypes.ContractManager
	xccResolver xcctypes.XCCResolver

	basekeeper.Keeper
}

func NewKeeper(
	cdc codec.Marshaler,
	cm txtypes.ContractManager,
	xccResolver xcctypes.XCCResolver,
	baseKeeper basekeeper.Keeper,
) Keeper {
	return Keeper{
		cdc:         cdc,
		cm:          cm,
		xccResolver: xccResolver,
		Keeper:      baseKeeper,
	}
}

// SendCall starts a simple commit flow
// caller is Coordinator
func (k Keeper) SendCall(
	ctx sdk.Context,
	packetSender packets.PacketSender,
	txID txtypes.TxID,
	transactions []txtypes.ResolvedContractTransaction,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) error {
	if len(transactions) != 2 {
		return errors.New("the number of contract transactions must be 2")
	} else if !k.xccResolver.Capabilities().CrossChainCalls(ctx) && (len(transactions[0].Objects) > 0 || len(transactions[1].Objects) > 0) {
		return errors.New("the chainResolver cannot resolve cannot support the cross-chain calls feature")
	}

	if _, found := k.GetCoordinatorState(ctx, txID); found {
		return fmt.Errorf("txID '%X' already exists", txID)
	}

	tx0 := transactions[TxIndexCoordinator]
	tx1 := transactions[TxIndexParticipant]

	xcc0, err := tx0.GetCrossChainChannel(k.cdc)
	if err != nil {
		return err
	}

	// check if xcc0 indicates our chain
	if !k.xccResolver.IsSelfCrossChainChannel(ctx, xcc0) {
		return fmt.Errorf("a cross-chain channel that txIndex '%v' indicates must be our chain", TxIndexCoordinator)
	}

	xcc1, err := tx1.GetCrossChainChannel(k.cdc)
	if err != nil {
		return err
	}
	ch0, err := k.xccResolver.ResolveCrossChainChannel(ctx, xcc0)
	if err != nil {
		return err
	}
	ch1, err := k.xccResolver.ResolveCrossChainChannel(ctx, xcc1)
	if err != nil {
		return err
	}

	c, found := k.ChannelKeeper().GetChannel(ctx, ch1.Port, ch1.Channel)
	if !found {
		return sdkerrors.Wrap(channeltypes.ErrChannelNotFound, ch1.Channel)
	}
	if err := k.cm.PrepareCommit(ctx, txID, TxIndexCoordinator, tx0); err != nil {
		return err
	}

	payload := types.NewPacketDataCall(txID, tx1)
	if err := k.SendPacket(
		ctx,
		packetSender,
		&payload,
		ch1.Port, ch1.Channel,
		c.Counterparty.PortId, c.Counterparty.ChannelId,
		timeoutHeight,
		timeoutTimestamp,
	); err != nil {
		return err
	}

	cs := commontypes.NewCoordinatorState(
		txtypes.COMMIT_PROTOCOL_SIMPLE,
		commontypes.COORDINATOR_PHASE_PREPARE,
		[]xcctypes.ChannelInfo{*ch0, *ch1},
	)
	if err := cs.Confirm(TxIndexCoordinator, *ch0); err != nil {
		return err
	}
	k.SetCoordinatorState(ctx, txID, cs)

	cTxState := commontypes.NewContractTransactionState(
		commontypes.CONTRACT_TRANSACTION_STATUS_PREPARE,
		commontypes.PREPARE_RESULT_OK,
		*ch0,
	)
	k.SetContractTransactionState(ctx, txID, TxIndexCoordinator, cTxState)
	return nil
}

// ReceiveCallPacket receives a PacketDataCall to commit a transaction
// caller is participant
func (k Keeper) ReceiveCallPacket(
	ctx sdk.Context,
	destPort,
	destChannel string,
	data types.PacketDataCall,
) (*txtypes.ContractCallResult, *types.PacketAcknowledgementCall, error) {
	// validate packet data upon receiving
	if err := data.ValidateBasic(); err != nil {
		return nil, nil, err
	}

	_, found := k.ChannelKeeper().GetChannel(ctx, destPort, destChannel)
	if !found {
		return nil, nil, fmt.Errorf("channel(port=%v channel=%v) not found", destPort, destChannel)
	}

	if _, ok := k.GetContractTransactionState(ctx, data.TxId, TxIndexParticipant); ok {
		return nil, nil, fmt.Errorf("txID '%x' already exists", data.TxId)
	}

	// if !k.xccResolver.Capabilities().CrossChainCalls(ctx, uint32(txtypes.COMMIT_PROTOCOL_SIMPLE)) && len(data.TxInfo.Tx.Links) > 0 {
	// 	return nil, nil, errors.New("CrossChainResolver cannot resolve cannot support the cross-chain calls feature")
	// }

	var (
		prepareStatus commontypes.PrepareResult
		commitStatus  types.CommitStatus
		ctxStatus     commontypes.ContractTransactionStatus
	)
	res, err := k.cm.CommitImmediately(ctx, data.TxId, TxIndexParticipant, data.Tx)
	if err != nil {
		prepareStatus = commontypes.PREPARE_RESULT_FAILED
		commitStatus = types.COMMIT_STATUS_FAILED
		ctxStatus = commontypes.CONTRACT_TRANSACTION_STATUS_ABORT
		k.Logger(ctx).Info("failed to CommitImmediatelyTransaction", "err", err)
	} else {
		prepareStatus = commontypes.PREPARE_RESULT_OK
		commitStatus = types.COMMIT_STATUS_OK
		ctxStatus = commontypes.CONTRACT_TRANSACTION_STATUS_COMMIT
	}

	txinfo := commontypes.NewContractTransactionState(
		ctxStatus,
		prepareStatus,
		xcctypes.ChannelInfo{Port: destPort, Channel: destChannel},
	)
	k.SetContractTransactionState(ctx, data.TxId, TxIndexParticipant, txinfo)
	return res, types.NewPacketAcknowledgementCall(commitStatus), nil
}

// ReceiveCallAcknowledgement receives PacketAcknowledgementCall to updates CoordinatorState
// caller is coordinator
func (k Keeper) ReceiveCallAcknowledgement(
	ctx sdk.Context,
	sourcePort string,
	sourceChannel string,
	ack types.PacketAcknowledgementCall,
	txID txtypes.TxID,
) (isCommittable bool, err error) {
	cs, found := k.GetCoordinatorState(ctx, txID)
	if !found {
		return false, fmt.Errorf("txID '%x' not found", txID)
	} else if cs.Phase != commontypes.COORDINATOR_PHASE_PREPARE {
		return false, fmt.Errorf("coordinator status must be '%v'", commontypes.COORDINATOR_PHASE_PREPARE.String())
	} else if cs.IsCompleted() {
		return false, errors.New("all transactions are already confirmed")
	}

	_, found = k.ChannelKeeper().GetChannel(ctx, sourcePort, sourceChannel)
	if !found {
		return false, sdkerrors.Wrap(channeltypes.ErrChannelNotFound, sourceChannel)
	}
	if err := cs.Confirm(TxIndexParticipant, xcctypes.ChannelInfo{Port: sourcePort, Channel: sourceChannel}); err != nil {
		return false, err
	}
	switch ack.Status {
	case types.COMMIT_STATUS_OK:
		cs.Decision = commontypes.COORDINATOR_DECISION_COMMIT
		isCommittable = true
	case types.COMMIT_STATUS_FAILED:
		cs.Decision = commontypes.COORDINATOR_DECISION_ABORT
		isCommittable = false
	default:
		panic("unreachable")
	}
	cs.Phase = commontypes.COORDINATOR_PHASE_COMMIT
	cs.AddAck(TxIndexCoordinator)
	cs.AddAck(TxIndexParticipant)
	if !cs.IsCompleted() || !cs.IsReceivedALLAcks() {
		panic("fatal error")
	}
	k.SetCoordinatorState(ctx, txID, *cs)
	return isCommittable, nil
}

func (k Keeper) TryCommit(
	ctx sdk.Context,
	txID txtypes.TxID,
	isCommittable bool,
) (*txtypes.ContractCallResult, error) {
	_, err := k.EnsureContractTransactionStatus(ctx, txID, TxIndexCoordinator, commontypes.CONTRACT_TRANSACTION_STATUS_PREPARE)
	if err != nil {
		return nil, err
	}
	var (
		res    *txtypes.ContractCallResult
		status commontypes.ContractTransactionStatus
	)
	if isCommittable {
		res, err = k.cm.Commit(ctx, txID, TxIndexCoordinator)
		if err != nil {
			return nil, err
		}
		status = commontypes.CONTRACT_TRANSACTION_STATUS_COMMIT
	} else {
		err = k.cm.Abort(ctx, txID, TxIndexCoordinator)
		if err != nil {
			return nil, err
		}
		status = commontypes.CONTRACT_TRANSACTION_STATUS_ABORT
	}
	k.UpdateContractTransactionStatus(ctx, txID, TxIndexCoordinator, status)
	return res, nil
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("cross/core/atomic/%s", TypeName))
}
