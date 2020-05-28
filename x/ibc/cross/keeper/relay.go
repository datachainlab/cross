package keeper

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
	"github.com/datachainlab/cross/x/ibc/cross/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

func (k Keeper) MulticastPreparePacket(
	ctx sdk.Context,
	sender sdk.AccAddress,
	msg types.MsgInitiate,
	transactions []types.ContractTransaction,
) (types.TxID, error) {
	if ctx.ChainID() != msg.ChainID {
		return types.TxID{}, fmt.Errorf("unexpected chainID: '%v' != '%v'", ctx.ChainID(), msg.ChainID)
	} else if ctx.BlockHeight() >= msg.TimeoutHeight {
		return types.TxID{}, fmt.Errorf("this msg is already timeout: current=%v timeout=%v", ctx.BlockHeight(), msg.TimeoutHeight)
	}

	txID := MakeTxID(ctx, msg)
	if _, ok := k.GetCoordinator(ctx, txID); ok {
		return types.TxID{}, fmt.Errorf("coordinator '%x' already exists", txID)
	}

	channelInfos := make([]types.ChannelInfo, len(transactions))
	tss := make([]string, len(transactions))
	lkr, err := types.MakeLinker(transactions)
	if err != nil {
		return types.TxID{}, err
	}
	for id, t := range transactions {
		src := t.Source
		c, found := k.channelKeeper.GetChannel(ctx, src.Port, src.Channel)
		if !found {
			return types.TxID{}, sdkerrors.Wrap(channel.ErrChannelNotFound, t.Source.Channel)
		}
		objs, err := lkr.Lookup(t.Links)
		if err != nil {
			return types.TxID{}, err
		}
		data := types.NewPacketDataPrepare(sender, txID, types.TxIndex(id), types.NewContractTransactionInfo(t, objs))
		if err := k.sendPacket(
			ctx,
			data.GetBytes(),
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

func (k Keeper) PrepareTransaction(
	ctx sdk.Context,
	contractHandler types.ContractHandler,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	data types.PacketDataPrepare,
) (uint8, error) {
	if _, ok := k.GetTx(ctx, data.TxID, data.TxIndex); ok {
		return 0, fmt.Errorf("txID '%x' already exists", data.TxID)
	}

	result := types.PREPARE_RESULT_OK
	if err := k.prepareTransaction(ctx, contractHandler, sourcePort, sourceChannel, destinationPort, destinationChannel, data); err != nil {
		result = types.PREPARE_RESULT_FAILED
		k.Logger(ctx).Info("failed to prepare transaction", "error", err.Error())
	}

	c, found := k.channelKeeper.GetChannel(ctx, destinationPort, destinationChannel)
	if !found {
		return 0, fmt.Errorf("channel(port=%v channel=%v) not found", destinationPort, destinationChannel)
	}
	hops := c.GetConnectionHops()
	connID := hops[len(hops)-1]

	txinfo := types.NewTxInfo(types.TX_STATUS_PREPARE, result, connID, data.TxInfo.Transaction.CallInfo)
	k.SetTx(ctx, data.TxID, data.TxIndex, txinfo)
	return result, nil
}

func (k Keeper) prepareTransaction(
	ctx sdk.Context,
	contractHandler types.ContractHandler,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	data types.PacketDataPrepare,
) error {
	constraint := data.TxInfo.Transaction.StateConstraint

	rs, err := types.MakeResolver(data.TxInfo.LinkObjects)
	if err != nil {
		return err
	}
	store, res, err := contractHandler.Handle(
		types.WithSigners(ctx, data.TxInfo.Transaction.Signers),
		data.TxInfo.Transaction.CallInfo,
		types.ContractRuntimeInfo{StateConstraintType: constraint.Type, ExternalObjectResolver: rs},
	)
	if err != nil {
		return err
	}

	if rv := data.TxInfo.Transaction.ReturnValue; !rv.IsNil() && !rv.Equal(res.GetData()) {
		return fmt.Errorf("unexpected return-value: expected='%X' actual='%X'", *rv, res.GetData())
	}

	id := MakeStoreTransactionID(data.TxID, data.TxIndex)
	if err := store.Precommit(id); err != nil {
		return err
	}
	if ops := store.OPs(); !ops.Equal(constraint.OPs) {
		return fmt.Errorf("unexpected ops: actual(%v) != declation(%v)", ops.String(), constraint.OPs.String())
	}
	k.SetContractResult(ctx, data.TxID, data.TxIndex, res)
	return nil
}

func (k Keeper) ReceivePrepareAcknowledgement(
	ctx sdk.Context,
	sourcePort string,
	sourceChannel string,
	ack types.PacketPrepareAcknowledgement,
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

	c, found := k.channelKeeper.GetChannel(ctx, sourcePort, sourceChannel)
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
		ch, found := k.channelKeeper.GetChannel(ctx, c.Port, c.Channel)
		if !found {
			return sdkerrors.Wrap(channel.ErrChannelNotFound, c.Channel)
		}
		data := types.NewPacketDataCommit(txID, types.TxIndex(id), isCommittable)
		if err := k.sendPacket(
			ctx,
			data.GetBytes(),
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
	data types.PacketDataCommit,
) (types.ContractHandlerResult, error) {
	tx, err := k.EnsureTxStatus(ctx, data.TxID, data.TxIndex, types.TX_STATUS_PREPARE)
	if err != nil {
		return nil, err
	}
	c, found := k.channelKeeper.GetChannel(ctx, destinationPort, destinationChannel)
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
	id := MakeStoreTransactionID(data.TxID, data.TxIndex)
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

func (k Keeper) sendPacket(
	ctx sdk.Context,
	data []byte,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	timeoutHeight uint64,
	timeoutTimestamp uint64,
) error {
	// get the next sequence
	seq, found := k.channelKeeper.GetNextSequenceSend(ctx, sourcePort, sourceChannel)
	if !found {
		return channel.ErrSequenceSendNotFound
	}
	packet := channel.NewPacket(
		data,
		seq,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
		timeoutHeight,
		timeoutTimestamp,
	)
	channelCap, ok := k.scopedKeeper.GetCapability(ctx, ibctypes.ChannelCapabilityPath(sourcePort, sourceChannel))
	if !ok {
		return sdkerrors.Wrap(channel.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	if err := k.channelKeeper.SendPacket(ctx, channelCap, packet); err != nil {
		return err
	}

	k.SetUnacknowledgedPacket(ctx, sourcePort, sourceChannel, seq)
	return nil
}

func makePacketKey(port, channel string, seq uint64) string {
	return fmt.Sprintf("packet/%v/%v/%v", port, channel, seq)
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

// PacketExecuted defines a wrapper function for the channel Keeper's function
// in order to expose it to the cross handler.
// Keeper retreives channel capability and passes it into channel keeper for authentication
func (k Keeper) PacketExecuted(ctx sdk.Context, packet channelexported.PacketI, acknowledgement []byte) error {
	chanCap, ok := k.scopedKeeper.GetCapability(ctx, ibctypes.ChannelCapabilityPath(packet.GetDestPort(), packet.GetDestChannel()))
	if !ok {
		return sdkerrors.Wrap(channel.ErrChannelCapabilityNotFound, "channel capability could not be retrieved for packet")
	}
	return k.channelKeeper.PacketExecuted(ctx, chanCap, packet, acknowledgement)
}

func MakeTxID(ctx sdk.Context, msg types.MsgInitiate) types.TxID {
	var txID [32]byte

	a := tmhash.Sum(msg.GetSignBytes())
	b := tmhash.Sum(types.MakeHashFromABCIHeader(ctx.BlockHeader()).Hash())

	h := tmhash.New()
	h.Write(a)
	h.Write(b)
	copy(txID[:], h.Sum(nil))
	return txID
}

func MakeStoreTransactionID(txID types.TxID, txIndex uint8) []byte {
	size := len(txID)
	bz := make([]byte, size+1)
	copy(bz[:size], txID[:])
	bz[size] = txIndex
	return bz
}
