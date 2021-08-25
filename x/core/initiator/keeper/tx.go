package keeper

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"

	"github.com/datachainlab/cross/x/core/initiator/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
	"github.com/datachainlab/cross/x/packets"
)

func (k Keeper) initTx(ctx sdk.Context, msg *types.MsgInitiateTx) (txtypes.TxID, bool, error) {
	txID := types.MakeTxID(msg)
	_, found := k.GetTxState(ctx, txID)
	if found {
		return nil, false, fmt.Errorf("txID '%X' already exists", txID)
	}

	state := types.NewInitiateTxState(*msg)

	if err := k.authenticator.InitAuthState(ctx, txID, msg.GetRequiredAccounts()); err != nil {
		return nil, false, err
	}

	xcc, err := func (ctx sdk.Context, msg *types.MsgInitiateTx) (xcctypes.XCC, error) {
		for _, tx := range msg.ContractTransactions {
			for _, signer := range tx.Signers {
				if bytes.Equal(msg.Sender, signer) {
					if xcc, err := xcctypes.UnpackCrossChainChannel(k.m, *tx.CrossChainChannel); err != nil {
						return nil, err
					} else {
						return k.xccResolver.ResolveCrossChainChannel(ctx, xcc)
					}
				}
			}
		}

		return nil, fmt.Errorf("the same signer does not exist. %v:%v", msg.Signers, msg.Sender)
	}(ctx, msg)
	if err != nil {
		return nil, false, err
	}

	completed, err := k.authenticator.Sign(ctx, txID, msg.GetAccounts(xcc))
	if err != nil {
		return nil, false, err
	}

	if completed {
		state.Status = types.INITIATE_TX_STATUS_VERIFIED
	}

	k.setTxState(ctx, txID, state)
	return txID, completed, nil
}

func (k Keeper) setTxState(ctx sdk.Context, txID txtypes.TxID, state types.InitiateTxState) {
	bz, err := proto.Marshal(&state)
	if err != nil {
		panic(err)
	}
	prefix.NewStore(k.store(ctx), types.KeyInitiateTxState()).Set(txID, bz)
}

func (k Keeper) GetTxState(ctx sdk.Context, txID txtypes.TxID) (*types.InitiateTxState, bool) {
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

func (k Keeper) runTx(ctx sdk.Context, txID txtypes.TxID, msg *types.MsgInitiateTx) error {
	wctx, ps, err := k.packetMiddleware.HandleMsg(ctx, msg, packets.NewBasicPacketSender(k.channelKeeper))
	if err != nil {
		return err
	}

	rtxs, err := k.ResolveTransactions(wctx, msg.ContractTransactions)
	if err != nil {
		return err
	}

	tx := txtypes.NewTx(txID, msg.CommitProtocol, rtxs, msg.TimeoutHeight, msg.TimeoutTimestamp)
	if err := k.txRunner.RunTx(wctx, tx, ps); err != nil {
		return err
	} else {
		return nil
	}
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
