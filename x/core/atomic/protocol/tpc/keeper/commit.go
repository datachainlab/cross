package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	"github.com/datachainlab/cross/x/packets"
)

func (k Keeper) SendCommit(
	ctx sdk.Context,
	packetSender packets.PacketSender,
	txID txtypes.TxID,
	isCommittable bool,
) error {
	// TODO implements
	return nil
}
