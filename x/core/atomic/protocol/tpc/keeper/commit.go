package keeper

import (
	"fmt"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/modules/core/04-channel/types"

	"github.com/datachainlab/cross/x/core/atomic/protocol/tpc/types"
	atomictypes "github.com/datachainlab/cross/x/core/atomic/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
	"github.com/datachainlab/cross/x/packets"
)

func (k Keeper) SendCommit(
	ctx sdk.Context,
	packetSender packets.PacketSender,
	txID crosstypes.TxID,
	isCommittable bool,
) error {
	cs, found := k.GetCoordinatorState(ctx, txID)
	if !found {
		return fmt.Errorf("txID '%x' not found", txID)
	} else if cs.Phase != atomictypes.COORDINATOR_PHASE_PREPARE {
		return fmt.Errorf("coordinator status must be '%v'", atomictypes.COORDINATOR_PHASE_PREPARE.String())
	} else if cs.Decision == atomictypes.COORDINATOR_DECISION_UNKNOWN {
		return fmt.Errorf("coordinator must decide any status")
	}

	// NOTE: packet-timeout isn't supported in the two-phase commit
	timeoutHeight := clienttypes.NewHeight(
		clienttypes.ParseChainID(ctx.ChainID()),
		math.MaxUint64,
	)
	for id, c := range cs.Channels {
		ch, found := k.ChannelKeeper().GetChannel(ctx, c.Port, c.Channel)
		if !found {
			return sdkerrors.Wrap(channeltypes.ErrChannelNotFound, c.Channel)
		}
		pd := types.NewPacketDataCommit(txID, crosstypes.TxIndex(id), isCommittable)
		if err := k.SendPacket(
			ctx,
			packetSender,
			pd,
			c.Port,
			c.Channel,
			ch.GetCounterparty().GetPortID(),
			ch.GetCounterparty().GetChannelID(),
			timeoutHeight,
			0,
		); err != nil {
			return err
		}
	}
	cs.Phase = atomictypes.COORDINATOR_PHASE_COMMIT
	k.SetCoordinatorState(ctx, txID, *cs)
	return nil
}

func (k Keeper) ReceivePacketCommit(
	ctx sdk.Context,
	destPort,
	destChannel string,
	data types.PacketDataCommit,
) (*txtypes.ContractCallResult, *types.PacketAcknowledgementCommit, error) {
	// Validations

	txState, err := k.EnsureContractTransactionStatus(
		ctx,
		data.TxId, data.TxIndex,
		atomictypes.CONTRACT_TRANSACTION_STATUS_PREPARE,
	)
	if err != nil {
		return nil, nil, err
	}
	_, found := k.ChannelKeeper().GetChannel(ctx, destPort, destChannel)
	if !found {
		return nil, nil, fmt.Errorf("channel not found: port=%v channel=%v", destPort, destChannel)
	}

	ci := &xcctypes.ChannelInfo{Channel: destChannel, Port: destPort}
	if !txState.CoordinatorChannel.Equal(ci) {
		return nil, nil, fmt.Errorf("expected CoordinatorChannel is %v, but got %v", txState.CoordinatorChannel, ci)
	}

	// Try to Commit or Abort

	if data.IsCommittable {
		res, err := k.cm.Commit(ctx, data.TxId, data.TxIndex)
		if err != nil {
			ack := types.NewPacketAcknowledgementCommit(types.COMMIT_STATUS_FAILED)
			ack.ErrorMessage = err.Error()
			return nil, ack, nil
		}
		return res, types.NewPacketAcknowledgementCommit(types.COMMIT_STATUS_OK), nil
	} else {
		err := k.cm.Abort(ctx, data.TxId, data.TxIndex)
		if err != nil {
			return nil, nil, err
		}
		return nil, types.NewPacketAcknowledgementCommit(types.COMMIT_STATUS_OK), nil
	}
}

func (k Keeper) ReceiveCommitAcknowledgement(
	ctx sdk.Context,
	txID crosstypes.TxID,
	txIndex crosstypes.TxIndex,
) error {
	cs, found := k.GetCoordinatorState(ctx, txID)
	if !found {
		return fmt.Errorf("txID '%x' not found", txID)
	} else if cs.Phase != atomictypes.COORDINATOR_PHASE_COMMIT {
		return fmt.Errorf("coordinator status must be '%v'", atomictypes.COORDINATOR_PHASE_COMMIT.String())
	} else if cs.Decision == atomictypes.COORDINATOR_DECISION_UNKNOWN {
		return fmt.Errorf("coordinator must decide any status")
	}

	if !cs.AddAck(txIndex) {
		return fmt.Errorf("tx '%v' already exists", txIndex)
	}

	k.SetCoordinatorState(ctx, txID, *cs)
	return nil
}
