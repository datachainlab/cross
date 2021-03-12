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
	atomictypes "github.com/datachainlab/cross/x/core/atomic/types"
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
	} else if uint64(ctx.BlockHeight()) >= timeoutHeight.GetVersionHeight() {
		return fmt.Errorf("the given timeoutHeight is in the past: current=%v timeout=%v", ctx.BlockHeight(), timeoutHeight.GetVersionHeight())
	} else if _, found := k.GetCoordinatorState(ctx, txID); found {
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

	var prepareResult atomictypes.PrepareResult

	// TODO returns a result of contract call
	_, err = k.cm.PrepareCommit(ctx, txID, TxIndexCoordinator, tx0)
	if err != nil {
		prepareResult = atomictypes.PREPARE_RESULT_FAILED
		k.Logger(ctx).Info("failed to PrepareCommit", "err", err)
	} else {
		prepareResult = atomictypes.PREPARE_RESULT_OK
		if err := k.SendPacket(
			ctx,
			packetSender,
			types.NewPacketDataCall(txID, tx1),
			ch1.Port, ch1.Channel,
			c.Counterparty.PortId, c.Counterparty.ChannelId,
			timeoutHeight,
			timeoutTimestamp,
		); err != nil {
			return err
		}
	}

	k.SetCoordinatorState(ctx, txID, makeCoordinatorState([]xcctypes.ChannelInfo{*ch0, *ch1}, prepareResult))
	k.SetContractTransactionState(ctx, txID, TxIndexCoordinator, makeSenderContractTransactionState(prepareResult, *ch0))
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

	var commitStatus types.CommitStatus
	res, err := k.cm.CommitImmediately(ctx, data.TxId, TxIndexParticipant, data.Tx)
	if err != nil {
		commitStatus = types.COMMIT_STATUS_FAILED
		k.Logger(ctx).Info("failed to CommitImmediatelyTransaction", "err", err)
	} else {
		commitStatus = types.COMMIT_STATUS_OK
	}
	k.SetContractTransactionState(
		ctx,
		data.TxId,
		TxIndexParticipant,
		makeReceiverContractTransactionState(
			xcctypes.ChannelInfo{Port: destPort, Channel: destChannel},
			commitStatus,
		),
	)
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
	} else if cs.Phase != atomictypes.COORDINATOR_PHASE_PREPARE {
		return false, fmt.Errorf("coordinator status must be '%v'", atomictypes.COORDINATOR_PHASE_PREPARE.String())
	} else if cs.IsConfirmedALLPrepares() {
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
		cs.Decision = atomictypes.COORDINATOR_DECISION_COMMIT
		isCommittable = true
	case types.COMMIT_STATUS_FAILED:
		cs.Decision = atomictypes.COORDINATOR_DECISION_ABORT
		isCommittable = false
	default:
		panic("unreachable")
	}
	cs.Phase = atomictypes.COORDINATOR_PHASE_COMMIT
	cs.AddAck(TxIndexCoordinator)
	cs.AddAck(TxIndexParticipant)
	if !cs.IsConfirmedALLPrepares() || !cs.IsConfirmedALLCommits() {
		panic("fatal error")
	}
	k.SetCoordinatorState(ctx, txID, *cs)
	return isCommittable, nil
}

// TryCommit try to commit or abort a transaction
// caller is coordinator
func (k Keeper) TryCommit(
	ctx sdk.Context,
	txID txtypes.TxID,
	isCommittable bool,
) (*txtypes.ContractCallResult, error) {
	_, err := k.EnsureContractTransactionStatus(ctx, txID, TxIndexCoordinator, atomictypes.CONTRACT_TRANSACTION_STATUS_PREPARE)
	if err != nil {
		return nil, err
	}
	var (
		res    *txtypes.ContractCallResult
		status atomictypes.ContractTransactionStatus
	)
	if isCommittable {
		res, err = k.cm.Commit(ctx, txID, TxIndexCoordinator)
		if err != nil {
			return nil, err
		}
		status = atomictypes.CONTRACT_TRANSACTION_STATUS_COMMIT
	} else {
		err = k.cm.Abort(ctx, txID, TxIndexCoordinator)
		if err != nil {
			return nil, err
		}
		status = atomictypes.CONTRACT_TRANSACTION_STATUS_ABORT
	}
	k.UpdateContractTransactionStatus(ctx, txID, TxIndexCoordinator, status)
	return res, nil
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("cross/core/atomic/%s", TypeName))
}

// makeCoordinatorState returns a state of coordinator after `prepare`
// CONTRACT:
// - `channels` length must be 2
// - `prepareResult` must be PREPARE_RESULT_OK or PREPARE_RESULT_FAILED
func makeCoordinatorState(channels []xcctypes.ChannelInfo, prepareResult atomictypes.PrepareResult) atomictypes.CoordinatorState {
	if len(channels) != 2 {
		panic(fmt.Errorf("channels length must be 2"))
	}

	var (
		coordinatorPhase    atomictypes.CoordinatorPhase
		coordinatorDecision atomictypes.CoordinatorDecision
	)

	if prepareResult == atomictypes.PREPARE_RESULT_OK {
		coordinatorPhase = atomictypes.COORDINATOR_PHASE_PREPARE
		coordinatorDecision = atomictypes.COORDINATOR_DECISION_UNKNOWN
	} else if prepareResult == atomictypes.PREPARE_RESULT_FAILED {
		coordinatorPhase = atomictypes.COORDINATOR_PHASE_COMMIT
		coordinatorDecision = atomictypes.COORDINATOR_DECISION_ABORT
	} else {
		panic(fmt.Errorf("unexpected value: %v", prepareResult))
	}

	cs := atomictypes.NewCoordinatorState(
		txtypes.COMMIT_PROTOCOL_SIMPLE,
		coordinatorPhase,
		channels,
	)
	cs.Decision = coordinatorDecision
	if err := cs.Confirm(TxIndexCoordinator, channels[0]); err != nil {
		panic(err)
	}
	return cs
}

// makeSenderContractTransactionState returns a ContractTransactionState of coordinator after `prepare`
// CONTRACT:
// - `prepareResult` must be PREPARE_RESULT_OK or PREPARE_RESULT_FAILED
func makeSenderContractTransactionState(prepareResult atomictypes.PrepareResult, channel xcctypes.ChannelInfo) atomictypes.ContractTransactionState {
	if prepareResult == atomictypes.PREPARE_RESULT_OK {
		return atomictypes.NewContractTransactionState(atomictypes.CONTRACT_TRANSACTION_STATUS_PREPARE, prepareResult, channel)
	} else if prepareResult == atomictypes.PREPARE_RESULT_FAILED {
		return atomictypes.NewContractTransactionState(atomictypes.CONTRACT_TRANSACTION_STATUS_ABORT, prepareResult, channel)
	} else {
		panic(fmt.Errorf("unexpected value: %v", prepareResult))
	}
}

// makeReceiverContractTransactionState returns a ContractTransactionState of participant(doesn't have the coordinator role) after `commit`
// CONTRACT:
// - `commitStatus` must be COMMIT_STATUS_OK or COMMIT_STATUS_FAILED
func makeReceiverContractTransactionState(channel xcctypes.ChannelInfo, commitStatus types.CommitStatus) atomictypes.ContractTransactionState {
	if commitStatus == types.COMMIT_STATUS_OK {
		return atomictypes.NewContractTransactionState(
			atomictypes.CONTRACT_TRANSACTION_STATUS_COMMIT,
			atomictypes.PREPARE_RESULT_OK,
			channel,
		)
	} else if commitStatus == types.COMMIT_STATUS_FAILED {
		return atomictypes.NewContractTransactionState(
			atomictypes.CONTRACT_TRANSACTION_STATUS_ABORT,
			atomictypes.PREPARE_RESULT_FAILED,
			channel,
		)
	} else {
		panic(fmt.Errorf("unexpected value: %v", commitStatus))
	}
}
