package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	accounttypes "github.com/datachainlab/cross/x/core/account/types"
	"github.com/datachainlab/cross/x/core/initiator/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
	"github.com/datachainlab/cross/x/packets"
)

// SendIBCSign sends PacketDataIBCSignTx
func (k Keeper) SendIBCSignTx(
	ctx sdk.Context,
	packetSender packets.PacketSender,
	xcc xcctypes.XCC,
	txID txtypes.TxID,
	signers []accounttypes.AccountID,
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
) (bool, error) {
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

func makeExternalAccounts(xcc xcctypes.XCC, signers []accounttypes.AccountID) []accounttypes.Account {
	var accs []accounttypes.Account
	for _, id := range signers {
		accs = append(accs, accounttypes.NewAccount(xcc, id))
	}
	return accs
}
