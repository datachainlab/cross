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

	commonkeeper "github.com/datachainlab/cross/x/atomic/common/keeper"
	commontypes "github.com/datachainlab/cross/x/atomic/common/types"
	"github.com/datachainlab/cross/x/atomic/simple/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
	"github.com/datachainlab/cross/x/packets"
)

const (
	TxIndexCoordinator crosstypes.TxIndex = 0
	TxIndexParticipant crosstypes.TxIndex = 1
)

const (
	TypeName = "simple"
)

type Keeper struct {
	cdc codec.Marshaler

	commonkeeper.Keeper
}

func NewKeeper(
	cdc codec.Marshaler,
	ck commonkeeper.Keeper,
) Keeper {
	return Keeper{
		cdc:    cdc,
		Keeper: ck,
	}
}

// SendCall starts a simple commit flow
// caller is Coordinator
func (k Keeper) SendCall(
	ctx sdk.Context,
	packetSender packets.PacketSender,
	txID crosstypes.TxID,
	transactions []crosstypes.ContractTransaction,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) error {
	if len(transactions) != 2 {
		return errors.New("the number of contract transactions must be 2")
	}

	tx0 := transactions[TxIndexCoordinator]
	tx1 := transactions[TxIndexParticipant]

	if !k.ChainResolver().Capabilities().CrossChainCalls() && (len(tx0.Links) > 0 || len(tx1.Links) > 0) {
		return errors.New("this chainResolver cannot resolve cannot support the cross-chain calls feature")
	}

	if _, found := k.GetCoordinatorState(ctx, txID); found {
		return fmt.Errorf("txID '%X' already exists", txID)
	}

	chain0, err := tx0.GetChainID(k.cdc)
	if err != nil {
		return err
	}

	// check if chain0 indicates our chain
	if !k.ChainResolver().IsOurChain(ctx, chain0) {
		return fmt.Errorf("a chainID that txIndex '%v' indicates must be our chain", TxIndexCoordinator)
	}

	chain1, err := tx1.GetChainID(k.cdc)
	if err != nil {
		return err
	}
	ch0, err := k.ChainResolver().ResolveChainID(ctx, chain0)
	if err != nil {
		return err
	}
	ch1, err := k.ChainResolver().ResolveChainID(ctx, chain1)
	if err != nil {
		return err
	}

	lkr, err := crosstypes.MakeLinker(k.cdc, k.ChainResolver(), transactions)
	if err != nil {
		return err
	}
	// chain0 always indicates our chain
	objs0, err := lkr.Resolve(ctx, chain0, tx0.Links)
	if err != nil {
		return err
	}
	objs1, err := lkr.Resolve(ctx, chain1, tx1.Links)
	if err != nil {
		return err
	}

	c, found := k.ChannelKeeper().GetChannel(ctx, ch1.Port, ch1.Channel)
	if !found {
		return sdkerrors.Wrap(channeltypes.ErrChannelNotFound, ch1.Channel)
	}
	if err := k.PrepareCommit(ctx, txID, TxIndexCoordinator, tx0, objs0); err != nil {
		return err
	}

	payload := types.NewPacketDataCall(txID, crosstypes.NewContractTransactionInfo(tx1, objs1))
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
		commontypes.COMMIT_FLOW_SIMPLE,
		commontypes.COORDINATOR_PHASE_PREPARE,
		[]crosstypes.ChannelInfo{*ch0, *ch1},
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
) (*crosstypes.ContractCallResult, *types.PacketAcknowledgementCall, error) {
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

	if !k.ChainResolver().Capabilities().CrossChainCalls() && len(data.TxInfo.Tx.Links) > 0 {
		return nil, nil, errors.New("this chainResolver cannot resolve cannot support the cross-chain calls feature")
	}

	var (
		prepareStatus commontypes.PrepareResult
		commitStatus  types.CommitStatus
		ctxStatus     commontypes.ContractTransactionStatus
	)
	res, err := k.CommitImmediately(ctx, data.TxId, TxIndexParticipant, data.TxInfo.Tx, data.TxInfo.UnpackObjects(k.cdc))
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
		crosstypes.ChannelInfo{Port: destPort, Channel: destChannel},
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
	txID crosstypes.TxID,
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
	if err := cs.Confirm(TxIndexParticipant, crosstypes.ChannelInfo{Port: sourcePort, Channel: sourceChannel}); err != nil {
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
	txID crosstypes.TxID,
	isCommittable bool,
) (*crosstypes.ContractCallResult, error) {
	_, err := k.EnsureContractTransactionStatus(ctx, txID, TxIndexCoordinator, commontypes.CONTRACT_TRANSACTION_STATUS_PREPARE)
	if err != nil {
		return nil, err
	}
	var (
		res    *crosstypes.ContractCallResult
		status commontypes.ContractTransactionStatus
	)
	if isCommittable {
		res, err = k.Commit(ctx, txID, TxIndexCoordinator)
		if err != nil {
			return nil, err
		}
		status = commontypes.CONTRACT_TRANSACTION_STATUS_COMMIT
	} else {
		err = k.Abort(ctx, txID, TxIndexCoordinator)
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
	return ctx.Logger().With("module", fmt.Sprintf("cross/atomic/%s", TypeName))
}
