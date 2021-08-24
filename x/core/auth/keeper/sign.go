package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	"github.com/datachainlab/cross/x/core/auth/types"
	authtypes "github.com/datachainlab/cross/x/core/auth/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
	"github.com/datachainlab/cross/x/packets"
)

// SendIBCSign sends PacketDataIBCSignTx
func (k Keeper) SendIBCSignTx(
	ctx sdk.Context,
	packetSender packets.PacketSender,
	xcc xcctypes.XCC,
	txID crosstypes.TxID,
	signers []authtypes.AccountID,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) error {
	ci, err := k.xccResolver.ResolveCrossChainChannel(ctx, xcc)
	if err != nil {
		return err
	}

	c, found := k.channelKeeper.GetChannel(ctx, ci.Port, ci.Channel)
	if !found {
		return fmt.Errorf("channel '%v' not found", ci.String())
	}

	payload := types.NewPacketDataIBCSignTx(txID, signers, timeoutHeight, timeoutTimestamp)
	return k.SendPacket(
		ctx,
		packetSender,
		&payload,
		ci.Port, ci.Channel,
		c.Counterparty.PortId, c.Counterparty.ChannelId,
		timeoutHeight,
		timeoutTimestamp,
	)
}

// ReceiveIBCSignTx receives PacketDataIBCSignTx to verify the transaction
// If all required signs are completed, run the transaction
func (k Keeper) ReceiveIBCSignTx(
	ctx sdk.Context,
	destPort,
	destChannel string,
	data types.PacketDataIBCSignTx,
) (completed bool, err error) {
	// Validations

	if err := data.ValidateBasic(); err != nil {
		return false, err
	}
	_, found := k.channelKeeper.GetChannel(ctx, destPort, destChannel)
	if !found {
		return false, fmt.Errorf("channel(port=%v channel=%v) not found", destPort, destChannel)
	}
	xcc, err := k.xccResolver.ResolveChannel(ctx, &xcctypes.ChannelInfo{Port: destPort, Channel: destChannel})
	if err != nil {
		return false, err
	}

	// Verify the signers of transaction

	completed, err = k.Sign(ctx, data.TxID, makeExternalAccounts(xcc, data.Signers))
	if err != nil {
		return false, err
	} else if !completed {
		// TODO returns a status code of tx
		return false, nil
	}

	return true, nil
}

func makeExternalAccounts(xcc xcctypes.XCC, signers []authtypes.AccountID) []authtypes.Account {
	var accs []authtypes.Account
	for _, id := range signers {
		accs = append(accs, authtypes.NewAccount(xcc, id))
	}
	return accs
}
