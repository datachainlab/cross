package simple

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/core/keeper/common"
	"github.com/datachainlab/cross/x/core/types"
	"github.com/datachainlab/cross/x/packets"
)

const (
	TxIndexCoordinator types.TxIndex = 0
	TxIndexParticipant types.TxIndex = 1
)

type Keeper struct {
	cdc      codec.Marshaler
	storeKey sdk.StoreKey

	common.Keeper
}

func NewKeeper(
	cdc codec.Marshaler,
	storeKey sdk.StoreKey,
	ck common.Keeper,
) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		Keeper:   ck,
	}
}

// SendCall starts a simple commit flow
// caller is Coordinator
func (k Keeper) SendCall(
	ctx sdk.Context,
	packetSender packets.PacketSender,
	txID types.TxID,
	transactions []types.ContractTransaction,
) ([]byte, error) {
	lkr, err := types.MakeLinker(k.cdc, transactions)
	if err != nil {
		return nil, err
	}

	tx0 := transactions[TxIndexCoordinator]
	tx1 := transactions[TxIndexParticipant]

	// if !k.ChannelResolver().Capabilities().CrossChainCalls() && (len(tx0.Links) > 0 || len(tx1.Links) > 0) {
	// 	return nil, errors.New("this channelResolver cannot resolve cannot support the cross-chain calls feature")
	// }

	objs0, err := lkr.Resolve(tx0.Links)
	if err != nil {
		return nil, err
	}
	_, _ = tx1, objs0

	panic("not implemented error")
}
