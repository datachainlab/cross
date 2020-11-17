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
	xcc types.CrossChainChannel,
	txID types.TxID,
	signers []types.AccountID,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) error {
	ci, err := k.CrossChainChannelResolver().ResolveCrossChainChannel(ctx, xcc)
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

// ReceiveIBCSignTx receives PacketDataIBCSignTx to verify the transaction
// If all required signs are completed, run the transaction
func (k Keeper) ReceiveIBCSignTx(
	ctx sdk.Context,
	destPort,
	destChannel string,
	data types.PacketDataIBCSignTx,
) (bool, error) {
	// Validations

	if err := data.ValidateBasic(); err != nil {
		return false, err
	}
	_, found := k.ChannelKeeper().GetChannel(ctx, destPort, destChannel)
	if !found {
		return false, fmt.Errorf("channel(port=%v channel=%v) not found", destPort, destChannel)
	}
	xcc, err := k.CrossChainChannelResolver().ResolveChannel(ctx, &types.ChannelInfo{Port: destPort, Channel: destChannel})
	if err != nil {
		return false, err
	}

	// Verify the signers of transaction

	completed, err := k.verifyTx(ctx, data.TxID, makeExternalAccounts(xcc, data.Signers))
	if err != nil {
		return false, err
	} else if !completed {
		// TODO returns a status code of tx
		return false, nil
	}

	// Run the transaction

	msg, found := k.getTxMsg(ctx, data.TxID)
	if !found {
		return false, fmt.Errorf("txMsg '%x' not found", data.TxID)
	}
	if err := k.runTx(ctx, data.TxID, msg); err != nil {
		return false, err
	}
	return true, nil
}

func makeExternalAccounts(xcc types.CrossChainChannel, signers []types.AccountID) []types.Account {
	var accs []types.Account
	for _, id := range signers {
		accs = append(accs, types.NewAccount(xcc, id))
	}
	return accs
}
