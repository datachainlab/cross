package naive

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	"github.com/datachainlab/cross/x/ibc/cross/keeper/common"
	"github.com/datachainlab/cross/x/ibc/cross/types"
	naivetypes "github.com/datachainlab/cross/x/ibc/cross/types/naive"
	"github.com/tendermint/tendermint/libs/log"
)

const TypeName = "naive"

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

func (k Keeper) SendCall(
	ctx sdk.Context,
	contractHandler types.ContractHandler,
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

	tx0 := transactions[0]
	tx1 := transactions[1]

	lkr, err := types.MakeLinker(transactions)
	if err != nil {
		return types.TxID{}, err
	}

	objs0, err := lkr.Resolve(tx0.Links)
	if err != nil {
		return types.TxID{}, err
	}
	if err := k.PrepareTransaction(ctx, contractHandler, txID, 0, tx0, objs0); err != nil {
		return types.TxID{}, err
	}

	objs1, err := lkr.Resolve(tx1.Links)
	if err != nil {
		return types.TxID{}, err
	}

	c, found := k.ChannelKeeper().GetChannel(ctx, tx1.Source.Port, tx1.Source.Channel)
	if !found {
		return types.TxID{}, sdkerrors.Wrap(channel.ErrChannelNotFound, tx1.Source.Channel)
	}

	data := naivetypes.NewPacketDataCall(msg.Sender, txID, types.NewContractTransactionInfo(tx1, objs1))
	if err := k.SendPacket(
		ctx,
		data.GetBytes(),
		tx1.Source.Port, tx1.Source.Channel,
		c.Counterparty.PortID, c.Counterparty.ChannelID,
		data.GetTimeoutHeight(),
		data.GetTimeoutTimestamp(),
	); err != nil {
		return types.TxID{}, err
	}
	hops := c.GetConnectionHops()
	k.SetCoordinator(
		ctx,
		txID,
		types.NewCoordinatorInfo(
			types.CO_STATUS_INIT,
			[]string{hops[len(hops)-1]},
			[]types.ChannelInfo{types.NewChannelInfo(tx1.Source.Port, tx1.Source.Channel)},
		),
	)
	return txID, nil
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("cross/%s", TypeName))
}
