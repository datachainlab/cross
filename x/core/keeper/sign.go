package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/datachainlab/cross/x/core/types"
	"github.com/datachainlab/cross/x/packets"
)

// SendIBCSign sends PacketDataIBCSignTx
func (k Keeper) SendIBCSignTx(
	ctx sdk.Context,
	packetSender packets.PacketSender,
	chainID types.ChainID,
	txID types.TxID,
	signers []types.AccountID,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) error {
	ci, err := k.ChainResolver().Resolve(ctx, chainID)
	if err != nil {
		return err
	}

	c, found := k.ChannelKeeper().GetChannel(ctx, ci.Port, ci.Channel)
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

func (k Keeper) ReceiveIBCSignTx(
	ctx sdk.Context,
	destPort,
	destChannel string,
	data types.PacketDataIBCSignTx,
) error {

	// steps:
	// 1. lookup a tx with ID
	// 2. call signTx with data.signers

	_, found := k.ChannelKeeper().GetChannel(ctx, destPort, destChannel)
	if !found {
		return fmt.Errorf("channel(port=%v channel=%v) not found", destPort, destChannel)
	}

	// TODO implements
	panic("not implemented error")

	return nil
}
