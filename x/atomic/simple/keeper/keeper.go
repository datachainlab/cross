package keeper

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"

	commonkeeper "github.com/datachainlab/cross/x/atomic/common/keeper"
	commontypes "github.com/datachainlab/cross/x/atomic/common/types"
	types "github.com/datachainlab/cross/x/atomic/simple/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
	"github.com/datachainlab/cross/x/packets"
)

const (
	TxIndexCoordinator crosstypes.TxIndex = 0
	TxIndexParticipant crosstypes.TxIndex = 1
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
) error {
	tx0 := transactions[TxIndexCoordinator]
	tx1 := transactions[TxIndexParticipant]

	if !k.ChannelResolver().Capabilities().CrossChainCalls() && (len(tx0.Links) > 0 || len(tx1.Links) > 0) {
		return errors.New("this channelResolver cannot resolve cannot support the cross-chain calls feature")
	}

	chain0, err := tx0.GetChainID(k.cdc)
	if err != nil {
		return err
	}
	// TODO check if chain0 indicates our chain

	chain1, err := tx1.GetChainID(k.cdc)
	if err != nil {
		return err
	}
	ch0, err := k.ChannelResolver().Resolve(ctx, chain0)
	if err != nil {
		return err
	}
	ch1, err := k.ChannelResolver().Resolve(ctx, chain1)
	if err != nil {
		return err
	}

	lkr, err := crosstypes.MakeLinker(k.cdc, transactions)
	if err != nil {
		return err
	}
	objs0, err := lkr.Resolve(tx0.Links)
	if err != nil {
		return err
	}
	objs1, err := lkr.Resolve(tx1.Links)
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
		clienttypes.NewHeight(0, 0),
		0,
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

	ctxState := commontypes.NewContractTransactionState(
		commontypes.CONTRACT_TRANSACTION_STATUS_PREPARE,
		commontypes.PREPARE_RESULT_OK,
		*ch0,
	)
	k.SetContractTransactionState(ctx, txID, TxIndexCoordinator, ctxState)
	return nil
}
