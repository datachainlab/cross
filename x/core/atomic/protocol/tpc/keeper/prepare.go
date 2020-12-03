package keeper

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"

	"github.com/datachainlab/cross/x/core/atomic/protocol/tpc/types"
	atomictypes "github.com/datachainlab/cross/x/core/atomic/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
	"github.com/datachainlab/cross/x/packets"
)

func (k Keeper) SendPrepare(
	ctx sdk.Context,
	packetSender packets.PacketSender,
	txID txtypes.TxID,
	transactions []txtypes.ResolvedContractTransaction,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) error {
	if len(transactions) == 0 {
		return errors.New("the number of contract transactions must be greater than 1")
	} else if uint64(ctx.BlockHeight()) >= timeoutHeight.GetVersionHeight() {
		return fmt.Errorf("the given timeoutHeight is in the past: current=%v timeout=%v", ctx.BlockHeight(), timeoutHeight.GetVersionHeight())
	} else if _, found := k.GetCoordinatorState(ctx, txID); found {
		return fmt.Errorf("txID '%X' already exists", txID)
	}

	var channels []xcctypes.ChannelInfo
	for i, tx := range transactions {
		data := types.NewPacketDataPrepare(
			txID,
			tx,
			txtypes.TxIndex(i),
		)
		xcc, err := tx.GetCrossChainChannel(k.cdc)
		if err != nil {
			return err
		}
		ci, err := k.xccResolver.ResolveCrossChainChannel(ctx, xcc)
		if err != nil {
			return err
		}
		ch, found := k.ChannelKeeper().GetChannel(ctx, ci.Port, ci.Channel)
		if !found {
			return sdkerrors.Wrap(channeltypes.ErrChannelNotFound, ci.String())
		}
		if err := k.SendPacket(
			ctx,
			packetSender,
			&data,
			ci.Port, ci.Channel, ch.Counterparty.PortId, ch.Counterparty.ChannelId,
			timeoutHeight, timeoutTimestamp,
		); err != nil {
			return err
		}
		channels = append(channels, *ci)
	}

	cs := atomictypes.NewCoordinatorState(
		txtypes.COMMIT_PROTOCOL_TPC,
		atomictypes.COORDINATOR_PHASE_PREPARE,
		channels,
	)
	k.SetCoordinatorState(ctx, txID, cs)
	return nil
}

func (k Keeper) ReceivePacketPrepare(
	ctx sdk.Context,
	destPort,
	destChannel string,
	data types.PacketDataPrepare,
) (*txtypes.ContractCallResult, *types.PacketAcknowledgementPrepare, error) {
	// validate packet data upon receiving
	if err := data.ValidateBasic(); err != nil {
		return nil, nil, err
	}

	_, found := k.ChannelKeeper().GetChannel(ctx, destPort, destChannel)
	if !found {
		return nil, nil, fmt.Errorf("channel(port=%v channel=%v) not found", destPort, destChannel)
	}

	if _, ok := k.GetContractTransactionState(ctx, data.TxId, data.TxIndex); ok {
		return nil, nil, fmt.Errorf("txID '%x' already exists", data.TxId)
	}

	var prepareResult atomictypes.PrepareResult
	res, err := k.cm.PrepareCommit(ctx, data.TxId, data.TxIndex, data.Tx)
	if err != nil {
		k.Logger(ctx).Info("failed to prepare a commit", "error", err.Error())
		prepareResult = atomictypes.PREPARE_RESULT_FAILED
	} else {
		prepareResult = atomictypes.PREPARE_RESULT_OK
	}

	txState := atomictypes.NewContractTransactionState(
		atomictypes.CONTRACT_TRANSACTION_STATUS_PREPARE,
		prepareResult,
		xcctypes.ChannelInfo{Channel: destChannel, Port: destPort},
	)
	k.SetContractTransactionState(ctx, data.TxId, data.TxIndex, txState)

	return res, types.NewPacketAcknowledgementPayload(prepareResult), nil
}

func (k Keeper) HandlePacketAcknowledgementPrepare(
	ctx sdk.Context,
	sourcePort string,
	sourceChannel string,
	ack types.PacketAcknowledgementPrepare,
	txID txtypes.TxID,
	txIndex txtypes.TxIndex,
	ps packets.PacketSender,
) (*sdk.Result, error) {
	state, err := k.receivePrepareAcknowledgement(
		ctx,
		sourcePort, sourceChannel,
		ack,
		txID, txIndex,
	)
	if err != nil {
		return nil, err
	}

	if state.GoCommit && state.GoAbort {
		panic("fatal error")
	} else if state.GoCommit || state.GoAbort {
		if err := k.SendCommit(
			ctx,
			ps,
			txID,
			state.GoCommit,
		); err != nil {
			return nil, err
		}
	} else if state.AlreadyCommitted {
		// nop
	} else { // wait for more acks
		// nop
	}
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func (k Keeper) receivePrepareAcknowledgement(
	ctx sdk.Context,
	sourcePort string,
	sourceChannel string,
	ack types.PacketAcknowledgementPrepare,
	txID txtypes.TxID,
	txIndex txtypes.TxIndex,
) (*receivePrepareState, error) {
	if err := ack.ValidateBasic(); err != nil {
		return nil, err
	}
	cs, found := k.GetCoordinatorState(ctx, txID)
	if !found {
		return nil, fmt.Errorf("txID '%x' not found", txID)
	} else if cs.Phase != atomictypes.COORDINATOR_PHASE_PREPARE {
		return nil, fmt.Errorf("coordinator status must be '%v'", atomictypes.COORDINATOR_PHASE_PREPARE.String())
	} else if cs.IsCompleted() {
		return nil, errors.New("all transactions are already confirmed")
	}

	_, found = k.ChannelKeeper().GetChannel(ctx, sourcePort, sourceChannel)
	if !found {
		return nil, sdkerrors.Wrap(channeltypes.ErrChannelNotFound, sourceChannel)
	}

	if err := cs.Confirm(txIndex, xcctypes.ChannelInfo{Port: sourcePort, Channel: sourceChannel}); err != nil {
		return nil, err
	}

	var state receivePrepareState
	switch cs.Phase {
	case atomictypes.COORDINATOR_PHASE_PREPARE:
		switch ack.Result {
		case atomictypes.PREPARE_RESULT_OK:
			if cs.IsCompleted() {
				cs.Decision = atomictypes.COORDINATOR_DECISION_COMMIT
				state.GoCommit = true
			} else {
				// nop
			}
		case atomictypes.PREPARE_RESULT_FAILED:
			cs.Decision = atomictypes.COORDINATOR_DECISION_ABORT
			state.GoAbort = true
		default:
			panic(fmt.Sprintf("unexpected result %v", ack.Result))
		}
	case atomictypes.COORDINATOR_PHASE_COMMIT:
		state.AlreadyCommitted = true
	default:
		panic(fmt.Sprintf("unexpected phase %v", cs.Phase))
	}
	k.SetCoordinatorState(ctx, txID, *cs)
	return &state, nil
}

// receivePrepareState keeps a packet receiving state
type receivePrepareState struct {
	AlreadyCommitted bool // AlreadyCommitted indicates a boolean whether the tx had already reach the commit phase
	GoCommit         bool // GoCommit indicates a boolean whether the tx should be committed at next step
	GoAbort          bool // GoAbort indicates a boolean whether the tx should be aborted at next step
}
