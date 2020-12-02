package keeper

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/tendermint/tendermint/libs/log"

	basekeeper "github.com/datachainlab/cross/x/core/atomic/protocol/base/keeper"
	"github.com/datachainlab/cross/x/core/atomic/protocol/tpc/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
	"github.com/datachainlab/cross/x/packets"
)

const (
	TypeName = "tpc"
)

type Keeper struct {
	cdc codec.Marshaler

	cm          txtypes.ContractManager
	xccResolver xcctypes.XCCResolver

	basekeeper.Keeper
}

func NewKeeper(
	cdc codec.Marshaler,
	cm txtypes.ContractManager,
	xccResolver xcctypes.XCCResolver,
	baseKeeper basekeeper.Keeper,
) Keeper {
	return Keeper{
		cdc:         cdc,
		cm:          cm,
		xccResolver: xccResolver,
		Keeper:      baseKeeper,
	}
}

func (k Keeper) SendPrepare(
	ctx sdk.Context,
	packetSender packets.PacketSender,
	txID txtypes.TxID,
	transactions []txtypes.ResolvedContractTransaction,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) error {
	if len(transactions) > 0 {
		return errors.New("the number of contract transactions must be greater than 1")
	} else if uint64(ctx.BlockHeight()) >= timeoutHeight.GetVersionHeight() {
		return fmt.Errorf("the given timeoutHeight is in the past: current=%v timeout=%v", ctx.BlockHeight(), timeoutHeight.GetVersionHeight())
	} else if _, found := k.GetCoordinatorState(ctx, txID); found {
		return fmt.Errorf("txID '%X' already exists", txID)
	}

	for i, tx := range transactions {
		pd := types.NewPacketDataPrepare(
			txID,
			tx,
			txtypes.TxIndex(i),
		)
		// TODO send a packet to participants
		_ = pd
	}

	return nil
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("cross/core/atomic/%s", TypeName))
}
