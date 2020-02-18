package keeper

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/bluele/crossccc/x/ibc/crossccc/internal/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
)

func (k Keeper) MulticastPreparePacket(
	ctx sdk.Context,
	sender sdk.AccAddress,
	msg types.MsgInitiate,
	transitions []types.StateTransition,
) error {
	txID := msg.GetTxID()
	if _, ok := k.GetTx(ctx, txID); ok {
		return fmt.Errorf("txID '%x' already exists", txID)
	}
	if _, ok := k.GetCoordinator(ctx, txID); ok {
		return fmt.Errorf("coordinator '%x' already exists", txID)
	}

	var channels []channel.Channel
	var sequences []uint64
	for _, t := range transitions {
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
	if len(transitions) != len(channels) || len(channels) != len(sequences) {
		panic("unreachable")
	}

	tss := make([]string, len(transitions))
	for id, c := range channels {
		s := transitions[id].Source
		p := k.CreatePreparePacket(
			ctx,
			sequences[id],
			s.Port,
			s.Channel,
			c.Counterparty.PortID,
			c.Counterparty.ChannelID,
			txID,
			id,
			transitions[id],
			sender,
		)
		if err := k.channelKeeper.SendPacket(ctx, p); err != nil {
			return err
		}
		hops := c.GetConnectionHops()
		tss[id] = hops[len(hops)-1]
	}

	k.SetCoordinator(ctx, txID, NewCoordinatorInfo(CO_STATUS_INIT, tss))

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
	transitionID int,
	transition types.StateTransition,
	sender sdk.AccAddress,
) channel.Packet {
	packetData := types.NewPacketDataPrepare(sender, txID, transitionID, transition)
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
	packet := k.CreatePrepareResultPacket(seq, destinationPort, destinationChannel, sourcePort, sourceChannel, sender, data.TxID, data.TransitionID, status)
	if err := k.channelKeeper.SendPacket(ctx, packet); err != nil {
		return err
	}

	c, found := k.channelKeeper.GetChannel(ctx, destinationPort, destinationChannel)
	if !found {
		return errors.New("channel not found")
	}
	hops := c.GetConnectionHops()
	connID := hops[len(hops)-1]

	txinfo := NewTxInfo(TX_STATUS_PREPARE, connID, data.StateTransition.Contract)
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
	store, err := contractHandler.Handle(ctx, data.StateTransition.Contract)
	if err != nil {
		return err
	}
	if err := store.Precommit(data.TxID); err != nil {
		return err
	}
	if !store.OPs().Equal(data.StateTransition.OPs) {
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
	transitionID int,
	status uint8,
) channel.Packet {
	packetData := types.NewPacketDataPrepareResult(sender, txID, transitionID, status)
	return channel.NewPacket(
		packetData,
		seq,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
	)
}

func (k Keeper) MulticastCommitPacket(
	ctx sdk.Context,
	txID []byte,
	preparePackets []types.PreparePacket,
	sender sdk.AccAddress,
	isCommittable bool,
) error {

	co, ok := k.GetCoordinator(ctx, txID)
	if !ok {
		return fmt.Errorf("coordinator '%x' not found", txID)
	}
	if co.Status != CO_STATUS_INIT {
		return fmt.Errorf("expected status is %v, but got %v", CO_STATUS_INIT, co.Status)
	}
	tsSet := co.Set()

	var channels []channel.Channel
	var sequences []uint64
	for _, p := range preparePackets {
		data := p.Packet.GetData().(types.PacketDataPrepareResult)
		if !bytes.Equal(txID, data.TxID) {
			return fmt.Errorf("unexpected txID: %x", data.TxID)
		}

		c, found := k.channelKeeper.GetChannel(ctx, p.Source.Port, p.Source.Channel)
		if !found {
			return sdkerrors.Wrap(channel.ErrChannelNotFound, p.Source.Channel)
		}

		// get the next sequence
		seq, found := k.channelKeeper.GetNextSequenceSend(ctx, p.Source.Port, p.Source.Channel)
		if !found {
			return channel.ErrSequenceSendNotFound
		}

		hops := c.GetConnectionHops()
		connID := hops[len(hops)-1]
		pair := ConnectionTransitionPair{connID, data.TransitionID}
		if !tsSet.Contains(pair) {
			return errors.New("unknown packet")
		}
		tsSet.Remove(pair)

		channels = append(channels, c)
		sequences = append(sequences, seq)
	}
	if len(preparePackets) != len(channels) || len(channels) != len(sequences) {
		panic("unreachable")
	} else if size := tsSet.Cardinality(); size != 0 {
		return fmt.Errorf("tsset still have some elements: %v", size)
	}

	for i, c := range channels {
		s := preparePackets[i].Source
		p := k.CreateCommitPacket(
			ctx,
			sequences[i],
			s.Port,
			s.Channel,
			c.Counterparty.PortID,
			c.Counterparty.ChannelID,
			sender,
			txID,
			isCommittable,
		)
		if err := k.channelKeeper.SendPacket(ctx, p); err != nil {
			return err
		}
	}

	if err := k.UpdateCoordinatorStatus(ctx, txID, CO_STATUS_COMMIT); err != nil {
		return err
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

	// TODO Send an Ack packet to coordinator?

	return nil
}
