package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	accounttypes "github.com/datachainlab/cross/x/core/account/types"
	"github.com/datachainlab/cross/x/core/initiator/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	"github.com/datachainlab/cross/x/packets"
	"github.com/gogo/protobuf/proto"
)

func (k Keeper) initTx(ctx sdk.Context, msg *types.MsgInitiateTx) (txtypes.TxID, bool, error) {
	txID := types.MakeTxID(msg)
	_, found := k.getTxState(ctx, txID)
	if found {
		return nil, false, fmt.Errorf("txID '%X' already exists", txID)
	}

	signers := msg.GetAccounts(k.xccResolver.GetSelfCrossChainChannel(ctx))
	required := msg.GetRequiredAccounts()
	remaining := getRemainingAccounts(signers, required)

	state, err := k.initTxState(ctx, txID, msg, remaining)
	if err != nil {
		return nil, false, err
	}

	return txID, state.Status == types.INITIATE_TX_STATUS_VERIFIED, nil
}

func (k Keeper) initTxState(ctx sdk.Context, txID txtypes.TxID, msg *types.MsgInitiateTx, remainingSigners []accounttypes.Account) (*types.InitiateTxState, error) {
	k.setTxMsg(ctx, txID, msg)
	state := types.NewInitiateTxState(remainingSigners)
	k.setTxState(ctx, txID, state)
	return &state, nil
}

func (k Keeper) verifyTx(ctx sdk.Context, txID txtypes.TxID, signers []accounttypes.Account) (bool, error) {
	txState, found := k.getTxState(ctx, txID)
	if !found {
		return false, fmt.Errorf("txState '%x' not found", txID)
	}
	if txState.Status != types.INITIATE_TX_STATUS_PENDING {
		return false, fmt.Errorf("the status of txState '%x' must be %v", txID, types.INITIATE_TX_STATUS_PENDING)
	}
	return k.updateTxState(ctx, txID, *txState, getRemainingAccounts(signers, txState.RemainingSigners)), nil
}

func (k Keeper) setTxMsg(ctx sdk.Context, txID txtypes.TxID, msg *types.MsgInitiateTx) {
	bz, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	prefix.NewStore(k.store(ctx), types.KeyInitiateTx()).Set(txID, bz)
}

func (k Keeper) getTxMsg(ctx sdk.Context, txID txtypes.TxID) (*types.MsgInitiateTx, bool) {
	bz := prefix.NewStore(k.store(ctx), types.KeyInitiateTx()).Get(txID)
	if bz == nil {
		return nil, false
	}
	var msg types.MsgInitiateTx
	if err := proto.Unmarshal(bz, &msg); err != nil {
		panic(err)
	}
	return &msg, true
}

func (k Keeper) setTxState(ctx sdk.Context, txID txtypes.TxID, state types.InitiateTxState) {
	bz, err := proto.Marshal(&state)
	if err != nil {
		panic(err)
	}
	prefix.NewStore(k.store(ctx), types.KeyInitiateTxState()).Set(txID, bz)
}

func (k Keeper) updateTxState(ctx sdk.Context, txID txtypes.TxID, state types.InitiateTxState, remainingSigners []accounttypes.Account) bool {
	state.RemainingSigners = remainingSigners
	if len(remainingSigners) == 0 {
		state.Status = types.INITIATE_TX_STATUS_VERIFIED
	}
	k.setTxState(ctx, txID, state)
	return state.Status == types.INITIATE_TX_STATUS_VERIFIED
}

func (k Keeper) getTxState(ctx sdk.Context, txID txtypes.TxID) (*types.InitiateTxState, bool) {
	var state types.InitiateTxState
	bz := prefix.NewStore(k.store(ctx), types.KeyInitiateTxState()).Get(txID)
	if bz == nil {
		return nil, false
	}
	if err := proto.Unmarshal(bz, &state); err != nil {
		panic(err)
	}
	return &state, true
}

func (k Keeper) signTx(ctx sdk.Context, txID txtypes.TxID, signers []accounttypes.Account) (types.InitiateTxStatus, error) {
	state, found := k.getTxState(ctx, txID)
	if !found {
		return 0, fmt.Errorf("txID '%X' not found", txID)
	} else if state.Status != types.INITIATE_TX_STATUS_PENDING {
		return 0, fmt.Errorf("status must be %v", types.INITIATE_TX_STATUS_PENDING)
	}
	remaining := getRemainingAccounts(signers, state.RemainingSigners)
	if len(remaining) == 0 {
		state.Status = types.INITIATE_TX_STATUS_VERIFIED
	} else {
		state.Status = types.INITIATE_TX_STATUS_PENDING
	}
	k.setTxState(ctx, txID, *state)
	return state.Status, nil
}

func (k Keeper) runTx(ctx sdk.Context, txID txtypes.TxID, msg *types.MsgInitiateTx) error {
	ctx, ps, err := k.packetMiddleware.HandleMsg(ctx, msg, packets.NewBasicPacketSender(k.channelKeeper))
	if err != nil {
		return err
	}

	rtxs, err := k.ResolveTransactions(ctx, msg.ContractTransactions)
	if err != nil {
		return err
	}

	tx := txtypes.NewTx(txID, msg.CommitProtocol, rtxs, msg.TimeoutHeight, msg.TimeoutTimestamp)
	return k.txProcessor.RunTx(ctx, tx, ps)
}

func (k Keeper) ResolveTransactions(ctx sdk.Context, ctxs []types.ContractTransaction) ([]txtypes.ResolvedContractTransaction, error) {
	lkr, err := types.MakeLinker(k.m, k.xccResolver, ctxs)
	if err != nil {
		return nil, err
	}

	var rtxs []txtypes.ResolvedContractTransaction

	for _, ct := range ctxs {
		xcc, err := ct.GetCrossChainChannel(k.m)
		if err != nil {
			return nil, err
		}
		objs, err := lkr.Resolve(ctx, xcc, ct.Links)
		if err != nil {
			return nil, err
		}
		anyObjs, err := txtypes.PackObjects(objs)
		if err != nil {
			return nil, err
		}
		rt := txtypes.ResolvedContractTransaction{
			CrossChainChannel: ct.CrossChainChannel,
			Signers:           ct.Signers,
			CallInfo:          ct.CallInfo,
			ReturnValue:       ct.ReturnValue,
			Objects:           anyObjs,
		}
		rtxs = append(rtxs, rt)
	}

	return rtxs, nil
}

func getRemainingAccounts(signers, required []accounttypes.Account) []accounttypes.Account {
	var state = make([]bool, len(required))
	for i, acc := range required {
		for _, s := range signers {
			if acc.Equal(s) {
				state[i] = true
			}
		}
	}
	var remaining []accounttypes.Account
	for i, acc := range required {
		if !state[i] {
			remaining = append(remaining, acc)
		}
	}
	return remaining
}
