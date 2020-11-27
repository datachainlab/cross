package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/core/contract/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
)

type contractManager struct {
	cdc      codec.Marshaler
	storeKey sdk.StoreKey

	mod                     types.ContractModule
	commitStore             types.CommitStoreI
	resolverProvider        txtypes.ObjectResolverProvider
	contractHandleDecorator types.ContractHandleDecorator
}

var _ txtypes.ContractManager = (*contractManager)(nil)

func NewContractManager(
	cdc codec.Marshaler,
	storeKey sdk.StoreKey,
	mod types.ContractModule,
	commitStore types.CommitStoreI,
	contractHandleDecorator types.ContractHandleDecorator,
) txtypes.ContractManager {

	return contractManager{
		cdc:                     cdc,
		storeKey:                storeKey,
		mod:                     mod,
		commitStore:             commitStore,
		resolverProvider:        txtypes.DefaultResolverProvider(),
		contractHandleDecorator: contractHandleDecorator,
	}
}

func (k contractManager) PrepareCommit(
	ctx sdk.Context,
	txID txtypes.TxID,
	txIndex txtypes.TxIndex,
	tx txtypes.ResolvedContractTransaction,
) error {
	ctx, err := k.setupContext(ctx, tx, types.AtomicMode)
	if err != nil {
		return err
	}
	res, err := k.processTransaction(ctx, tx)
	if err != nil {
		return err
	}
	k.setContractCallResult(ctx, txID, txIndex, *res)
	return k.commitStore.Precommit(ctx, makeContractTransactionID(txID, txIndex))
}

func (k contractManager) setupContext(
	ctx sdk.Context,
	tx txtypes.ResolvedContractTransaction,
	commitMode types.CommitMode,
) (sdk.Context, error) {
	rs, err := k.resolverProvider(k.cdc, tx.UnpackObjects(k.cdc))
	if err != nil {
		return ctx, err
	}

	// Setup a context
	ctx = types.SetupContractContext(
		ctx,
		tx.Signers,
		types.ContractRuntimeInfo{
			CommitMode:             commitMode,
			ExternalObjectResolver: rs,
		},
	)
	goCtx, err := k.contractHandleDecorator.Handle(ctx.Context(), tx.CallInfo)
	if err != nil {
		return ctx, err
	}
	return ctx.WithContext(goCtx), nil
}

func (k contractManager) processTransaction(
	ctx sdk.Context,
	tx txtypes.ResolvedContractTransaction,
) (res *txtypes.ContractCallResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = types.NewErrContractCall(e)
			} else {
				err = types.NewErrContractCall(fmt.Errorf("type=%T value=%#v", e, e))
			}
		}
	}()

	res, err = k.mod.OnContractCall(
		sdk.WrapSDKContext(ctx),
		tx.CallInfo,
	)
	if err != nil {
		return nil, err
	}

	if !tx.ReturnValue.IsNil() && !tx.ReturnValue.Equal(txtypes.NewReturnValue(res.Data)) {
		return nil, fmt.Errorf("unexpected return-value: expected='%X' actual='%X'", *tx.ReturnValue, res.Data)
	}

	return res, nil
}

func (k contractManager) CommitImmediately(
	ctx sdk.Context,
	txID txtypes.TxID,
	txIndex txtypes.TxIndex,
	tx txtypes.ResolvedContractTransaction,
) (*txtypes.ContractCallResult, error) {
	ctx, err := k.setupContext(ctx, tx, types.BasicMode)
	if err != nil {
		return nil, err
	}
	res, err := k.processTransaction(ctx, tx)
	if err != nil {
		return nil, err
	}
	k.commitStore.CommitImmediately(ctx)
	return res, nil
}

// Commit commits the transaction
func (k contractManager) Commit(
	ctx sdk.Context,
	txID txtypes.TxID,
	txIndex txtypes.TxIndex,
) (*txtypes.ContractCallResult, error) {
	if err := k.commitStore.Commit(ctx, makeContractTransactionID(txID, txIndex)); err != nil {
		return nil, err
	}
	res := k.getContractCallResult(ctx, txID, txIndex)
	// TODO calls OnCommit handler
	k.removeContractCallResult(ctx, txID, txIndex)
	return res, nil
}

// Abort aborts the transaction
func (k contractManager) Abort(
	ctx sdk.Context,
	txID txtypes.TxID,
	txIndex txtypes.TxIndex,
) error {
	if err := k.commitStore.Abort(ctx, makeContractTransactionID(txID, txIndex)); err != nil {
		return err
	}
	// TODO calls OnAbort handler
	k.removeContractCallResult(ctx, txID, txIndex)
	return nil
}
