package keeper

import (
	"fmt"

	"github.com/bluele/crossccc/x/ibc/crossccc/internal/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
)

func (k Keeper) MulticastInitiatePacket(
	ctx sdk.Context,
	sender sdk.AccAddress,
	msg types.MsgInitiate,
	transitions []types.StateTransition,
) error {
	// Fetch each channel info and its next sequence
	// So, Instantiator must be *HUB* with concerned networks
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

	for i, c := range channels {
		s := transitions[i].Source
		err := k.createInitiatePacket(
			ctx,
			sequences[i],
			s.Port,
			s.Channel,
			c.Counterparty.PortID,
			c.Counterparty.ChannelID,
			txID,
			transitions[i],
			sender,
		)
		if err != nil {
			return err
		}
	}

	k.SetCoordinator(ctx, txID, CoordinatorInfo{CO_STATUS_INIT})

	return nil
}

func (k Keeper) createInitiatePacket(
	ctx sdk.Context,
	seq uint64,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	txID []byte,
	transition types.StateTransition,
	sender sdk.AccAddress,
) error {
	packetData := types.NewPacketDataInitiate(sender, txID, transition)
	packet := channel.NewPacket(
		packetData,
		seq,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
	)
	return k.channelKeeper.SendPacket(ctx, packet)
}

func (k Keeper) PrepareTransaction(
	ctx sdk.Context,
	contractHandler ContractHandler,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	data types.PacketDataInitiate,
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
	packet := k.CreatePreparePacket(seq, destinationPort, destinationChannel, sourcePort, sourceChannel, sender, data.TxID, status)
	if err := k.channelKeeper.SendPacket(ctx, packet); err != nil {
		return err
	}

	txinfo := NewTxInfo(TX_STATUS_PREPARE, types.ChannelInfo{Port: sourcePort, Channel: sourceChannel}, data.StateTransition.Contract)
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
	data types.PacketDataInitiate,
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
	// TODO set contract info
	return nil
}

func (k Keeper) CreatePreparePacket(
	seq uint64,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	sender sdk.AccAddress,
	txID []byte,
	status uint8,
) channel.Packet {
	packetData := types.NewPacketDataPrepare(sender, txID, status)
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
	prepareInfoList []types.PrepareInfo,
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

	var channels []channel.Channel
	var sequences []uint64
	for _, info := range prepareInfoList {
		c, found := k.channelKeeper.GetChannel(ctx, info.Source.Port, info.Source.Channel)
		if !found {
			return sdkerrors.Wrap(channel.ErrChannelNotFound, info.Source.Channel)
		}

		// get the next sequence
		seq, found := k.channelKeeper.GetNextSequenceSend(ctx, info.Source.Port, info.Source.Channel)
		if !found {
			return channel.ErrSequenceSendNotFound
		}

		channels = append(channels, c)
		sequences = append(sequences, seq)
	}
	if len(prepareInfoList) != len(channels) || len(channels) != len(sequences) {
		panic("unreachable")
	}

	for i, c := range channels {
		s := prepareInfoList[i].Source
		err := k.createCommitPacket(
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
		if err != nil {
			return err
		}
	}

	k.UpdateCoordinatorStatus(ctx, txID, CO_STATUS_COMMIT)

	return nil
}

func (k Keeper) createCommitPacket(
	ctx sdk.Context,
	seq uint64,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	sender sdk.AccAddress,
	txID []byte,
	isCommitable bool,
) error {
	packetData := types.NewPacketDataCommit(sender, txID, isCommitable)
	packet := channel.NewPacket(
		packetData,
		seq,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
	)
	return k.channelKeeper.SendPacket(ctx, packet)
}

// Precondition:
// - PacketCommit is included at coordinator chain
func (k Keeper) ReceiveCommitPacket(
	ctx sdk.Context,
	contractHandler ContractHandler,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	data types.PacketDataCommit,
	sender sdk.AccAddress,
) error {
	tx, err := k.EnsureTxStatus(ctx, data.TxID, TX_STATUS_PREPARE)
	if err != nil {
		return err
	}
	if tx.Coordinator.Port != sourcePort || tx.Coordinator.Channel != sourceChannel {
		return fmt.Errorf("expected coordinator is {%v, %v}, but got {%v, %v}", tx.Coordinator.Port, tx.Coordinator.Channel, sourcePort, sourceChannel)
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

	// Save Ack packet?

	return nil
}
