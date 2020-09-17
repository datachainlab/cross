package tpc

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	"github.com/datachainlab/cross/x/ibc/cross/keeper/common"
	"github.com/datachainlab/cross/x/ibc/cross/types"
	tpctypes "github.com/datachainlab/cross/x/ibc/cross/types/tpc"
	"github.com/tendermint/tendermint/libs/log"
)

const TypeName = "tpc"

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

func (k Keeper) MulticastPreparePacket(
	ctx sdk.Context,
	packetSender types.PacketSender,
	sender sdk.AccAddress,
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

	channelInfos := make([]types.ChannelInfo, len(transactions))
	tss := make([]string, len(transactions))
	lkr, err := types.MakeLinker(transactions)
	if err != nil {
		return types.TxID{}, err
	}
	for id, tx := range transactions {
		var objs []types.Object
		if len(tx.Links) > 0 {
			if !k.ChannelResolver().Capabilities().CrossChainCalls() {
				return types.TxID{}, errors.New("this channelResolver cannot support the cross-chain calls feature")
			}
			objs, err = lkr.Resolve(tx.Links)
			if err != nil {
				return types.TxID{}, err
			}
		} else {
			objs = nil
		}

		src, err := k.ChannelResolver().Resolve(ctx, tx.ChainID)
		if err != nil {
			return types.TxID{}, err
		}
		c, found := k.ChannelKeeper().GetChannel(ctx, src.Port, src.Channel)
		if !found {
			return types.TxID{}, sdkerrors.Wrap(channel.ErrChannelNotFound, src.Channel)
		}

		data := tpctypes.NewPacketDataPrepare(sender, txID, types.TxIndex(id), types.NewContractTransactionInfo(tx, objs))
		if err := k.SendPacket(
			ctx,
			packetSender,
			data,
			src.Port, src.Channel,
			c.Counterparty.PortID, c.Counterparty.ChannelID,
			data.GetTimeoutHeight(),
			data.GetTimeoutTimestamp(),
		); err != nil {
			return types.TxID{}, err
		}
		hops := c.GetConnectionHops()
		tss[id] = hops[len(hops)-1]
		channelInfos[id] = types.NewChannelInfo(src.Port, src.Channel)
	}

	k.SetCoordinator(ctx, txID, types.NewCoordinatorInfo(types.CO_STATUS_INIT, tss, channelInfos))

	return txID, nil
}

func (k Keeper) Prepare(
	ctx sdk.Context,
	contractHandler types.ContractHandler,
	sourcePort,
	sourceChannel string,
	data tpctypes.PacketDataPrepare,
) (uint8, error) {
	if _, ok := k.GetTx(ctx, data.TxID, data.TxIndex); ok {
		return 0, fmt.Errorf("txID '%x' already exists", data.TxID)
	}

	if !k.ChannelResolver().Capabilities().CrossChainCalls() && len(data.TxInfo.LinkObjects) > 0 {
		return 0, errors.New("this channelResolver cannot resolve cannot support the cross-chain calls feature")
	}

	result := types.PREPARE_RESULT_OK
	if err := k.PrepareTransaction(ctx, contractHandler, data.TxID, data.TxIndex, data.TxInfo.Transaction, data.TxInfo.LinkObjects); err != nil {
		result = types.PREPARE_RESULT_FAILED
		k.Logger(ctx).Info("failed to prepare transaction", "error", err.Error())
	}

	c, found := k.ChannelKeeper().GetChannel(ctx, sourcePort, sourceChannel)
	if !found {
		return 0, fmt.Errorf("channel(port=%v channel=%v) not found", sourcePort, sourceChannel)
	}
	hops := c.GetConnectionHops()
	connID := hops[len(hops)-1]

	txinfo := types.NewTxInfo(types.TX_STATUS_PREPARE, result, connID, data.TxInfo.Transaction.CallInfo)
	k.SetTx(ctx, data.TxID, data.TxIndex, txinfo)
	return result, nil
}

func (k Keeper) ReceivePrepareAcknowledgement(
	ctx sdk.Context,
	sourcePort string,
	sourceChannel string,
	ack tpctypes.PacketPrepareAcknowledgement,
	txID types.TxID,
	txIndex types.TxIndex,
) (canMulticast bool, isCommittable bool, err error) {
	co, ok := k.GetCoordinator(ctx, txID)
	if !ok {
		return false, false, fmt.Errorf("coordinator '%x' not found", txID)
	} else if co.Status == types.CO_STATUS_NONE {
		return false, false, errors.New("coordinator status must not be CO_STATUS_NONE")
	} else if co.IsCompleted() {
		return false, false, errors.New("all transactions are already confirmed")
	}

	c, found := k.ChannelKeeper().GetChannel(ctx, sourcePort, sourceChannel)
	if !found {
		return false, false, sdkerrors.Wrap(channel.ErrChannelNotFound, sourceChannel)
	}
	hops := c.GetConnectionHops()
	if err := co.Confirm(txIndex, hops[len(hops)-1]); err != nil {
		return false, false, err
	}

	if co.Status == types.CO_STATUS_INIT {
		if ack.Status == types.PREPARE_RESULT_FAILED {
			co.Status = types.CO_STATUS_DECIDED
			co.Decision = types.CO_DECISION_ABORT
		} else if ack.Status == types.PREPARE_RESULT_OK {
			if co.IsCompleted() {
				co.Status = types.CO_STATUS_DECIDED
				co.Decision = types.CO_DECISION_COMMIT
			}
		} else {
			panic("unreachable")
		}
		canMulticast = co.Status == types.CO_STATUS_DECIDED
	} else if co.Status == types.CO_STATUS_DECIDED {
		canMulticast = false
	} else {
		panic("unreachable")
	}

	k.SetCoordinator(ctx, txID, *co)
	return canMulticast, co.Decision == types.CO_DECISION_COMMIT, nil
}

func (k Keeper) MulticastCommitPacket(
	ctx sdk.Context,
	packetSender types.PacketSender,
	txID types.TxID,
	isCommittable bool,
) error {
	co, ok := k.GetCoordinator(ctx, txID)
	if !ok {
		return fmt.Errorf("coordinator '%x' not found", txID)
	} else if co.Status != types.CO_STATUS_DECIDED {
		return errors.New("coordinator status must be CO_STATUS_DECIDED")
	}

	for id, c := range co.Channels {
		ch, found := k.ChannelKeeper().GetChannel(ctx, c.Port, c.Channel)
		if !found {
			return sdkerrors.Wrap(channel.ErrChannelNotFound, c.Channel)
		}
		data := tpctypes.NewPacketDataCommit(txID, types.TxIndex(id), isCommittable)
		if err := k.SendPacket(
			ctx,
			packetSender,
			data,
			c.Port,
			c.Channel,
			ch.GetCounterparty().GetPortID(),
			ch.GetCounterparty().GetChannelID(),
			data.GetTimeoutHeight(),
			data.GetTimeoutTimestamp(),
		); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) ReceiveCommitPacket(
	ctx sdk.Context,
	contractHandler types.ContractHandler,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	data tpctypes.PacketDataCommit,
) (types.ContractHandlerResult, error) {
	tx, err := k.EnsureTxStatus(ctx, data.TxID, data.TxIndex, types.TX_STATUS_PREPARE)
	if err != nil {
		return nil, err
	}
	c, found := k.ChannelKeeper().GetChannel(ctx, destinationPort, destinationChannel)
	if !found {
		return nil, fmt.Errorf("channel not found: port=%v channel=%v", destinationPort, destinationChannel)
	}
	hops := c.GetConnectionHops()
	connID := hops[len(hops)-1]

	if tx.CoordinatorConnectionID != connID {
		return nil, fmt.Errorf("expected coordinatorConnectionID is %v, but got %v", tx.CoordinatorConnectionID, connID)
	}

	state, err := contractHandler.GetState(ctx, tx.ContractCallInfo, types.ContractRuntimeInfo{StateConstraintType: types.NoStateConstraint})
	if err != nil {
		return nil, err
	}

	var res types.ContractHandlerResult
	var status uint8
	id := common.MakeStoreTransactionID(data.TxID, data.TxIndex)
	if data.IsCommittable {
		if err := state.Commit(id); err != nil {
			return nil, err
		}
		status = types.TX_STATUS_COMMIT
		r, ok := k.GetContractResult(ctx, data.TxID, data.TxIndex)
		if !ok {
			return nil, fmt.Errorf("Can't find the execution result of contract handler")
		}
		res = contractHandler.OnCommit(ctx, r)
	} else {
		if tx.PrepareResult == types.PREPARE_RESULT_OK {
			if err := state.Discard(id); err != nil {
				return nil, err
			}
		} else if tx.PrepareResult == types.PREPARE_RESULT_FAILED {
			// nop
		} else {
			panic("unreachable")
		}
		status = types.TX_STATUS_ABORT
		res = types.ContractHandlerAbortResult{}
	}

	k.RemoveContractResult(ctx, data.TxID, data.TxIndex)
	if err := k.UpdateTxStatus(ctx, data.TxID, data.TxIndex, status); err != nil {
		return nil, err
	}

	return res, nil
}

func (k Keeper) PacketCommitAcknowledgement(ctx sdk.Context, txID types.TxID, txIndex types.TxIndex) error {
	ci, err := k.EnsureCoordinatorStatus(ctx, txID, types.CO_STATUS_DECIDED)
	if err != nil {
		return err
	}
	if !ci.AddAck(txIndex) {
		return fmt.Errorf("transactionID '%v' is already received", txIndex)
	}
	k.SetCoordinator(ctx, txID, *ci)
	return nil
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("cross/%s", TypeName))
}
