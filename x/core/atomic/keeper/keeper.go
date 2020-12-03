package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"

	basekeeper "github.com/datachainlab/cross/x/core/atomic/protocol/base/keeper"
	simplekeeper "github.com/datachainlab/cross/x/core/atomic/protocol/simple/keeper"
	tpckeeper "github.com/datachainlab/cross/x/core/atomic/protocol/tpc/keeper"
	"github.com/datachainlab/cross/x/core/atomic/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
	"github.com/datachainlab/cross/x/packets"
)

var _ txtypes.TxRunner = (*Keeper)(nil)

type Keeper struct {
	baseKeeper   basekeeper.Keeper
	simpleKeeper simplekeeper.Keeper
	tpcKeeper    tpckeeper.Keeper

	packetMiddleware packets.PacketMiddleware
	packetSender     packets.PacketSender
}

func NewKeeper(
	cdc codec.Marshaler,
	storeKey sdk.StoreKey,
	channelKeeper types.ChannelKeeper,
	portKeeper types.PortKeeper,
	scopedKeeper capabilitykeeper.ScopedKeeper,
	cm txtypes.ContractManager,
	xccResolver xcctypes.XCCResolver,
	packetMiddleware packets.PacketMiddleware,
) Keeper {
	baseKeeper := basekeeper.NewKeeper(cdc, storeKey, channelKeeper, portKeeper, scopedKeeper)
	simpleKeeper := simplekeeper.NewKeeper(cdc, cm, xccResolver, baseKeeper)
	tpcKeeper := tpckeeper.NewKeeper(cdc, cm, xccResolver, baseKeeper)
	return Keeper{
		baseKeeper:       baseKeeper,
		simpleKeeper:     simpleKeeper,
		tpcKeeper:        tpcKeeper,
		packetSender:     packets.NewBasicPacketSender(channelKeeper),
		packetMiddleware: packetMiddleware,
	}
}

func (k Keeper) RunTx(ctx sdk.Context, tx txtypes.Tx, ps packets.PacketSender) error {
	// Run a transaction

	switch tx.CommitProtocol {
	case txtypes.COMMIT_PROTOCOL_SIMPLE:
		err := k.simpleKeeper.SendCall(ctx, ps, tx.Id, tx.ContractTransactions, tx.TimeoutHeight, tx.TimeoutTimestamp)
		if err != nil {
			return sdkerrors.Wrap(types.ErrFailedInitiateTx, err.Error())
		}
	default:
		return fmt.Errorf("unknown commit protocol '%v'", tx.CommitProtocol)
	}

	return nil
}

func (k Keeper) SimpleKeeper() simplekeeper.Keeper {
	return k.simpleKeeper
}

func (k Keeper) TPCKeeper() tpckeeper.Keeper {
	return k.tpcKeeper
}
