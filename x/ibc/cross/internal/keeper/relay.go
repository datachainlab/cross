package keeper

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	"github.com/datachainlab/cross/x/ibc/cross/internal/types"
)

func (k Keeper) MulticastPreparePacket(
	ctx sdk.Context,
	sender sdk.AccAddress,
	msg types.MsgInitiate,
	transactions []types.ContractTransaction,
) error {
	if ctx.ChainID() != msg.ChainID {
		return fmt.Errorf("unexpected chainID: '%v' != '%v'", ctx.ChainID(), msg.ChainID)
	} else if ctx.BlockHeight() >= msg.TimeoutHeight {
		return fmt.Errorf("this msg is already timeout: current=%v timeout=%v", ctx.BlockHeight(), msg.TimeoutHeight)
	}

	txID := msg.GetTxID()
	if _, ok := k.GetTx(ctx, txID); ok {
		return fmt.Errorf("txID '%x' already exists", txID)
	}
	if _, ok := k.GetCoordinator(ctx, txID); ok {
		return fmt.Errorf("coordinator '%x' already exists", txID)
	}

	var channels []channel.Channel
	var sequences []uint64
	for _, t := range transactions {
		c, found := k.channelKeeper.GetChannel(ctx, t.Source.Port, t.Source.Channel)
		if !found {
			return sdkerrors.Wrap(channel.ErrChannelNotFound, t.Source.Channel)
		}

		// get the next sequence
		seq, found := k.channelKeeper.GetNextSequenceSend(ctx, t.Source.Port, t.Source.Channel)
		if !found {
			return channel.ErrSequenceSendNotFound
		}

		channels = append(channels, c)
		sequences = append(sequences, seq)
	}
	if len(transactions) != len(channels) || len(channels) != len(sequences) {
		panic("unreachable")
	}

	channelInfos := make([]types.ChannelInfo, len(transactions))
	tss := make([]string, len(transactions))
	for id, c := range channels {
		s := transactions[id].Source
		p := k.CreatePreparePacket(
			ctx,
			sequences[id],
			s.Port,
			s.Channel,
			c.Counterparty.PortID,
			c.Counterparty.ChannelID,
			txID,
			id,
			transactions[id],
			sender,
		)
		if err := k.channelKeeper.SendPacket(ctx, p); err != nil {
			return err
		}
		hops := c.GetConnectionHops()
		tss[id] = hops[len(hops)-1]
		channelInfos[id] = types.NewChannelInfo(s.Port, s.Channel)
	}

	k.SetCoordinator(ctx, txID, NewCoordinatorInfo(CO_STATUS_INIT, tss, channelInfos))

	return nil
}

func (k Keeper) CreatePreparePacket(
	ctx sdk.Context,
	seq uint64,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	txID []byte,
	transactionID int,
	transaction types.ContractTransaction,
	sender sdk.AccAddress,
) channel.Packet {
	packetData := types.NewPacketDataPrepare(sender, txID, transactionID, transaction)
	packet := channel.NewPacket(
		packetData,
		seq,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
	)
	return packet
}

func (k Keeper) PrepareTransaction(
	ctx sdk.Context,
	contractHandler ContractHandler,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	data types.PacketDataPrepare,
	sender sdk.AccAddress,
) error {
	if _, ok := k.GetTx(ctx, data.TxID); ok {
		return fmt.Errorf("txID '%x' already exists", data.TxID)
	}

	status := types.PREPARE_STATUS_OK
	if err := k.prepareTransaction(ctx, contractHandler, sourcePort, sourceChannel, destinationPort, destinationChannel, data, sender); err != nil {
		status = types.PREPARE_STATUS_FAILED
	}

	// get the next sequence
	seq, found := k.channelKeeper.GetNextSequenceSend(ctx, destinationPort, destinationChannel)
	if !found {
		return channel.ErrSequenceSendNotFound
	}

	// Send a Prepared Packet to coordinator (reply to source channel)
	packet := k.CreatePrepareResultPacket(seq, destinationPort, destinationChannel, sourcePort, sourceChannel, sender, data.TxID, data.TransactionID, status)
	if err := k.channelKeeper.SendPacket(ctx, packet); err != nil {
		return err
	}

	c, found := k.channelKeeper.GetChannel(ctx, destinationPort, destinationChannel)
	if !found {
		return errors.New("channel not found")
	}
	hops := c.GetConnectionHops()
	connID := hops[len(hops)-1]

	txinfo := NewTxInfo(TX_STATUS_PREPARE, connID, data.ContractTransaction.Contract)
	k.SetTx(ctx, data.TxID, txinfo)
	return nil
}

func (k Keeper) prepareTransaction(
	ctx sdk.Context,
	contractHandler ContractHandler,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	data types.PacketDataPrepare,
	sender sdk.AccAddress,
) error {
	store, err := contractHandler.Handle(
		types.WithSigners(ctx, data.ContractTransaction.Signers),
		data.ContractTransaction.Contract,
	)
	if err != nil {
		return err
	}
	if err := store.Precommit(data.TxID); err != nil {
		return err
	}
	if !store.OPs().Equal(data.ContractTransaction.OPs) {
		return fmt.Errorf("unexpected ops")
	}
	return nil
}

func (k Keeper) CreatePrepareResultPacket(
	seq uint64,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	sender sdk.AccAddress,
	txID []byte,
	transactionID int,
	status uint8,
) channel.Packet {
	packetData := types.NewPacketDataPrepareResult(sender, txID, transactionID, status)
	return channel.NewPacket(
		packetData,
		seq,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
	)
}

func (k Keeper) ReceivePrepareResultPacket(
	ctx sdk.Context,
	packet channel.Packet,
	data types.PacketDataPrepareResult,
) (canDecide bool, isCommittable bool, err error) {
	co, ok := k.GetCoordinator(ctx, data.TxID)
	if !ok {
		return false, false, fmt.Errorf("coordinator '%x' not found", data.TxID)
	} else if co.Status == CO_STATUS_NONE {
		return false, false, errors.New("coordinator status must not be CO_STATUS_NONE")
	} else if co.IsCompleted() {
		return false, false, errors.New("all transactions are already confirmed")
	}

	c, found := k.channelKeeper.GetChannel(ctx, packet.DestinationPort, packet.DestinationChannel)
	if !found {
		return false, false, sdkerrors.Wrap(channel.ErrChannelNotFound, packet.DestinationChannel)
	}
	hops := c.GetConnectionHops()
	if err := co.Confirm(data.TransactionID, hops[len(hops)-1]); err != nil {
		return false, false, err
	}

	if data.Status == types.PREPARE_STATUS_FAILED {
		co.Status = CO_STATUS_DECIDED
		co.Decision = CO_DECISION_ABORT
	} else if data.Status == types.PREPARE_STATUS_OK {
		if co.IsCompleted() {
			co.Status = CO_STATUS_DECIDED
			co.Decision = CO_DECISION_COMMIT
		}
	} else {
		panic("unreachable")
	}

	k.SetCoordinator(ctx, data.TxID, *co)
	return co.Status == CO_STATUS_DECIDED, co.Decision == CO_DECISION_COMMIT, nil
}

func (k Keeper) MulticastCommitPacket(
	ctx sdk.Context,
	txID []byte,
	sender sdk.AccAddress,
	isCommittable bool,
) error {
	co, ok := k.GetCoordinator(ctx, txID)
	if !ok {
		return fmt.Errorf("coordinator '%x' not found", txID)
	} else if co.Status != CO_STATUS_DECIDED {
		return errors.New("coordinator status must be CO_STATUS_DECIDED")
	}

	for _, c := range co.Channels {
		ch, found := k.channelKeeper.GetChannel(ctx, c.Port, c.Channel)
		if !found {
			return sdkerrors.Wrap(channel.ErrChannelNotFound, c.Channel)
		}
		// get the next sequence
		seq, found := k.channelKeeper.GetNextSequenceSend(ctx, c.Port, c.Channel)
		if !found {
			return channel.ErrSequenceSendNotFound
		}
		packet := k.CreateCommitPacket(
			ctx,
			seq,
			c.Port,
			c.Channel,
			ch.GetCounterparty().GetPortID(),
			ch.GetCounterparty().GetChannelID(),
			sender,
			txID,
			isCommittable,
		)
		if err := k.channelKeeper.SendPacket(ctx, packet); err != nil {
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
	sender sdk.AccAddress,
	txID []byte,
	isCommitable bool,
) channel.Packet {
	packetData := types.NewPacketDataCommit(sender, txID, isCommitable)
	return channel.NewPacket(
		packetData,
		seq,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
	)
}

func (k Keeper) ReceiveCommitPacket(
	ctx sdk.Context,
	contractHandler ContractHandler,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	data types.PacketDataCommit,
) error {
	tx, err := k.EnsureTxStatus(ctx, data.TxID, TX_STATUS_PREPARE)
	if err != nil {
		return err
	}
	c, found := k.channelKeeper.GetChannel(ctx, destinationPort, destinationChannel)
	if !found {
		return fmt.Errorf("channel not found: port=%v channel=%v", destinationPort, destinationChannel)
	}
	hops := c.GetConnectionHops()
	connID := hops[len(hops)-1]

	if tx.CoordinatorConnectionID != connID {
		return fmt.Errorf("expected coordinatorConnectionID is %v, but got %v", tx.CoordinatorConnectionID, connID)
	}

	state, err := contractHandler.GetState(ctx, tx.Contract)
	if err != nil {
		return err
	}

	var status uint8
	if data.IsCommittable {
		if err := state.Commit(data.TxID); err != nil {
			return err
		}
		status = TX_STATUS_COMMIT
	} else {
		if err := state.Discard(data.TxID); err != nil {
			return err
		}
		status = TX_STATUS_ABORT
	}

	if err := k.UpdateTxStatus(ctx, data.TxID, status); err != nil {
		return err
	}

	return nil
}
