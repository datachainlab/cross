package simple

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/core/keeper/common"
	"github.com/datachainlab/cross/x/core/types"
	"github.com/datachainlab/cross/x/packets"
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
	txID []byte,
	transactions []types.ContractTransaction,
) ([]byte, error) {

	panic("not implemented error")
}
