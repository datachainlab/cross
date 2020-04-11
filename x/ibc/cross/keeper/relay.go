package keeper

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
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
	for id, t := range transactions {
		c, found := k.channelKeeper.GetChannel(ctx, t.Source.Port, t.Source.Channel)
		if !found {
			return types.TxID{}, sdkerrors.Wrap(channel.ErrChannelNotFound, t.Source.Channel)
		}

		data := types.NewPacketDataPrepare(sender, txID, types.TxIndex(id), transactions[id])
		s := transactions[id].Source
		err := k.sendPacket(
			ctx,
			data.GetBytes(),
			s.Port, s.Channel,
			c.Counterparty.PortID, c.Counterparty.ChannelID,
			data.GetTimeoutHeight(),
		)
		if err != nil {
			return types.TxID{}, err
		}
		hops := c.GetConnectionHops()
		tss[id] = hops[len(hops)-1]
		channelInfos[id] = types.NewChannelInfo(s.Port, s.Channel)
	}

	k.SetCoordinator(ctx, txID, types.NewCoordinatorInfo(types.CO_STATUS_INIT, tss, channelInfos))

	return txID, nil
}

func (k Keeper) CreatePreparePacket(
	ctx sdk.Context,
	seq uint64,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	txID types.TxID,
	txIndex types.TxIndex,
	transaction types.ContractTransaction,
	sender sdk.AccAddress,
) channel.Packet {
	packetData := types.NewPacketDataPrepare(sender, txID, txIndex, transaction)
	packet := channel.NewPacket(
		packetData.GetBytes(),
		seq,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
		packetData.GetTimeoutHeight(),
	)
	return packet
}

func (k Keeper) PrepareTransaction(
	ctx sdk.Context,
	contractHandler types.ContractHandler,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	data types.PacketDataPrepare,
) error {
	if _, ok := k.GetTx(ctx, data.TxID, data.TxIndex); ok {
		return fmt.Errorf("txID '%x' already exists", data.TxID)
	}

	status := types.PREPARE_STATUS_OK
	if err := k.prepareTransaction(ctx, contractHandler, sourcePort, sourceChannel, destinationPort, destinationChannel, data); err != nil {
		status = types.PREPARE_STATUS_FAILED
	}

	// Send a Prepared Packet to coordinator (reply to source channel)
	packetData := types.NewPacketDataPrepareResult(data.TxID, data.TxIndex, status)
	if err := k.sendPacket(
		ctx,
		packetData.GetBytes(),
		destinationPort, destinationChannel,
		sourcePort, sourceChannel,
		packetData.GetTimeoutHeight(),
	); err != nil {
		return err
	}

	c, found := k.channelKeeper.GetChannel(ctx, destinationPort, destinationChannel)
	if !found {
		return errors.New("channel not found")
	}
	hops := c.GetConnectionHops()
	connID := hops[len(hops)-1]

	txinfo := types.NewTxInfo(types.TX_STATUS_PREPARE, connID, data.ContractTransaction.Contract)
	k.SetTx(ctx, data.TxID, data.TxIndex, txinfo)
	return nil
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
	store, res, err := contractHandler.Handle(
		types.WithSigners(ctx, data.ContractTransaction.Signers),
		data.ContractTransaction.Contract,
	)
	if err != nil {
		return err
	}

	id := MakeStoreTransactionID(data.TxID, data.TxIndex)
	if err := store.Precommit(id); err != nil {
		return err
	}
	if ops := store.OPs(); !ops.Equal(data.ContractTransaction.OPs) {
		return fmt.Errorf("unexpected ops: %v != %v", ops.String(), data.ContractTransaction.OPs.String())
	}
	k.SetContractResult(ctx, data.TxID, data.TxIndex, res)
	return nil
}

func (k Keeper) ReceivePrepareResultPacket(
	ctx sdk.Context,
	packet channel.Packet,
	data types.PacketDataPrepareResult,
) (canMulticast bool, isCommittable bool, err error) {
	co, ok := k.GetCoordinator(ctx, data.TxID)
	if !ok {
		return false, false, fmt.Errorf("coordinator '%x' not found", data.TxID)
	} else if co.Status == types.CO_STATUS_NONE {
		return false, false, errors.New("coordinator status must not be CO_STATUS_NONE")
	} else if co.IsCompleted() {
		return false, false, errors.New("all transactions are already confirmed")
	}

	c, found := k.channelKeeper.GetChannel(ctx, packet.DestinationPort, packet.DestinationChannel)
	if !found {
		return false, false, sdkerrors.Wrap(channel.ErrChannelNotFound, packet.DestinationChannel)
	}
	hops := c.GetConnectionHops()
	if err := co.Confirm(data.TxIndex, hops[len(hops)-1]); err != nil {
		return false, false, err
	}

	if co.Status == types.CO_STATUS_INIT {
		if data.Status == types.PREPARE_STATUS_FAILED {
			co.Status = types.CO_STATUS_DECIDED
			co.Decision = types.CO_DECISION_ABORT
		} else if data.Status == types.PREPARE_STATUS_OK {
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

	k.SetCoordinator(ctx, data.TxID, *co)
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
		// get the next sequence
		seq, found := k.channelKeeper.GetNextSequenceSend(ctx, c.Port, c.Channel)
		if !found {
			return channel.ErrSequenceSendNotFound
		}
		channelCap, ok := k.scopedKeeper.GetCapability(ctx, ibctypes.ChannelCapabilityPath(c.Port, c.Channel))
		if !ok {
			return sdkerrors.Wrap(channel.ErrChannelCapabilityNotFound, "module does not own channel capability")
		}
		packet := k.CreateCommitPacket(
			ctx,
			seq,
			c.Port,
			c.Channel,
			ch.GetCounterparty().GetPortID(),
			ch.GetCounterparty().GetChannelID(),
			txID,
			types.TxIndex(id),
			isCommittable,
		)
		if err := k.channelKeeper.SendPacket(ctx, channelCap, packet); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) CreateCommitPacket(
	ctx sdk.Context,
	seq uint64,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	txID types.TxID,
	txIndex types.TxIndex,
	isCommitable bool,
) channel.Packet {
	packetData := types.NewPacketDataCommit(txID, txIndex, isCommitable)
	return channel.NewPacket(
		packetData.GetBytes(),
		seq,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
		packetData.GetTimeoutHeight(),
	)
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

	state, err := contractHandler.GetState(ctx, tx.Contract)
	if err != nil {
		return nil, err
	}
	res, ok := k.GetContractResult(ctx, data.TxID, data.TxIndex)
	if !ok {
		return nil, fmt.Errorf("Can't find the execution result of contract handler")
	}

	var status uint8
	id := MakeStoreTransactionID(data.TxID, data.TxIndex)
	if data.IsCommittable {
		if err := state.Commit(id); err != nil {
			return nil, err
		}
		status = types.TX_STATUS_COMMIT
		res = contractHandler.OnCommit(ctx, res)
	} else {
		if err := state.Discard(id); err != nil {
			return nil, err
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

func (k Keeper) SendAckCommitPacket(
	ctx sdk.Context,
	txID types.TxID,
	txIndex types.TxIndex,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
) error {
	data := types.NewPacketDataAckCommit(txID, txIndex)
	return k.sendPacket(ctx, data.GetBytes(), sourcePort, sourceChannel, destinationPort, destinationChannel, data.GetTimeoutHeight())
}

func (k Keeper) sendPacket(
	ctx sdk.Context,
	data []byte,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	timeout uint64,
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
		timeout,
	)
	channelCap, ok := k.scopedKeeper.GetCapability(ctx, ibctypes.ChannelCapabilityPath(sourcePort, sourceChannel))
	if !ok {
		return sdkerrors.Wrap(channel.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}
	return k.channelKeeper.SendPacket(ctx, channelCap, packet)
}

func (k Keeper) ReceiveAckPacket(ctx sdk.Context, txID types.TxID, txIndex types.TxIndex) error {
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
