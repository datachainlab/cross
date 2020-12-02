package keeper

import (
	"fmt"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"

	"github.com/datachainlab/cross/x/core/atomic/protocol/tpc/types"
	atomictypes "github.com/datachainlab/cross/x/core/atomic/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	"github.com/datachainlab/cross/x/packets"
)

func (k Keeper) SendCommit(
	ctx sdk.Context,
	packetSender packets.PacketSender,
	txID txtypes.TxID,
	isCommittable bool,
) error {
	cs, found := k.GetCoordinatorState(ctx, txID)
	if !found {
		return fmt.Errorf("txID '%x' not found", txID)
	} else if cs.Phase != atomictypes.COORDINATOR_PHASE_PREPARE {
		return fmt.Errorf("coordinator status must be '%v'", atomictypes.COORDINATOR_PHASE_PREPARE.String())
	} else if cs.Decision != atomictypes.COORDINATOR_DECISION_UNKNOWN {
		return fmt.Errorf("coordinator must decide any status")
	}

	timeoutHeight := clienttypes.NewHeight(
		clienttypes.ParseChainID(ctx.ChainID()),
		math.MaxInt64,
	)
	for id, c := range cs.Channels {
		ch, found := k.ChannelKeeper().GetChannel(ctx, c.Port, c.Channel)
		if !found {
			return sdkerrors.Wrap(channeltypes.ErrChannelNotFound, c.Channel)
		}
		pd := types.NewPacketDataCommit(txID, txtypes.TxIndex(id), isCommittable)
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
	if isCommittable {
		cs.Decision = atomictypes.COORDINATOR_DECISION_COMMIT
	} else {
		cs.Decision = atomictypes.COORDINATOR_DECISION_ABORT
	}
	k.SetCoordinatorState(ctx, txID, *cs)
	return nil
}
