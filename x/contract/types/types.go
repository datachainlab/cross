package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	accounttypes "github.com/datachainlab/cross/x/account/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	xcctypes "github.com/datachainlab/cross/x/xcc/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

type ContractManager interface {
	PrepareCommit(
		ctx sdk.Context,
		txID txtypes.TxID,
		txIndex txtypes.TxIndex,
		tx txtypes.ResolvedContractTransaction,
	) error
	CommitImmediately(
		ctx sdk.Context,
		txID txtypes.TxID,
		txIndex txtypes.TxIndex,
		tx txtypes.ResolvedContractTransaction,
	) (*ContractCallResult, error)
	Commit(
		ctx sdk.Context,
		txID txtypes.TxID,
		txIndex txtypes.TxIndex,
	) (*ContractCallResult, error)
	Abort(
		ctx sdk.Context,
		txID txtypes.TxID,
		txIndex txtypes.TxIndex,
	) error
}

type ContractModule interface {
	OnContractCall(ctx context.Context, callInfo txtypes.ContractCallInfo) (*ContractCallResult, error)
}

type ContractHandler func(ctx context.Context, callInfo txtypes.ContractCallInfo) (*ContractCallResult, error)

type ContractHandleDecorator interface {
	Handle(ctx context.Context, callInfo txtypes.ContractCallInfo) (newCtx context.Context, err error)
}

type ContractHandleDecorators []ContractHandleDecorator

var _ ContractHandleDecorator = (*ContractHandleDecorators)(nil)

func (decs ContractHandleDecorators) Handle(ctx context.Context, callInfo txtypes.ContractCallInfo) (newCtx context.Context, err error) {
	for _, dec := range decs {
		ctx, err = dec.Handle(ctx, callInfo)
		if err != nil {
			return nil, err
		}
	}
	return ctx, nil
}

func NewContractHandler(h ContractHandler, decs ...ContractHandleDecorator) ContractHandler {
	if h == nil {
		panic("ContractHandler cannot be nil")
	}
	return func(ctx context.Context, callInfo txtypes.ContractCallInfo) (*ContractCallResult, error) {
		var err error
		for _, dec := range decs {
			ctx, err = dec.Handle(ctx, callInfo)
			if err != nil {
				return nil, err
			}
		}
		return h(ctx, callInfo)
	}
}

func SetupContractContext(ctx sdk.Context, signers []accounttypes.AccountID, runtimeInfo ContractRuntimeInfo) sdk.Context {
	goCtx := ctx.Context()
	goCtx = ContextWithContractRuntimeInfo(goCtx, runtimeInfo)
	goCtx = ContextWithContractSigners(goCtx, signers)
	return ctx.WithContext(goCtx)
}

type ExternalContractCaller interface {
	Call(ctx sdk.Context, xcc xcctypes.XCC, callInfo txtypes.ContractCallInfo, signers []accounttypes.AccountID) []byte
}

type externalContractCaller struct{}

var _ ExternalContractCaller = (*externalContractCaller)(nil)

func (cc externalContractCaller) Call(ctx sdk.Context, xcc xcctypes.XCC, callInfo txtypes.ContractCallInfo, signers []accounttypes.AccountID) []byte {
	r := ContractRuntimeFromContext(ctx.Context()).ExternalObjectResolver
	key := MakeObjectKey(callInfo, signers)
	obj, err := r.Resolve(xcc, key)
	if err != nil {
		panic(err)
	}
	v, err := obj.Evaluate(key)
	if err != nil {
		panic(err)
	}
	return v
}

func NewExternalContractCaller() ExternalContractCaller {
	return externalContractCaller{}
}

// GetEvents converts Events to sdk.Events
func (res ContractCallResult) GetEvents() sdk.Events {
	events := make(sdk.Events, 0, len(res.Events))
	for _, ev := range res.Events {
		attrs := make([]sdk.Attribute, 0, len(ev.Attributes))
		for _, attr := range ev.Attributes {
			attrs = append(attrs, sdk.NewAttribute(string(attr.Key), string(attr.Value)))
		}
		events = append(events, sdk.NewEvent(ev.Type, attrs...))
	}
	return events
}

type ContractRuntimeInfo struct {
	CommitMode             CommitMode
	ExternalObjectResolver txtypes.ObjectResolver
}

// MakeObjectKey returns a key that can be used to identify a contract call
func MakeObjectKey(callInfo txtypes.ContractCallInfo, signers []accounttypes.AccountID) []byte {
	h := tmhash.New()
	h.Write(callInfo)
	for _, signer := range signers {
		h.Write(signer)
	}
	return h.Sum(nil)
}
