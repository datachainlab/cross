package types

import (
	"errors"
	"fmt"
	"sync"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/datachainlab/cross/x/core/auth/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

// Linker resolves links that each ContractTransaction has.
type Linker struct {
	cdc                       codec.Codec
	crossChainChannelResolver xcctypes.XCCResolver
	objects                   map[crosstypes.TxIndex]lazyObject
}

// MakeLinker returns Linker
func MakeLinker(cdc codec.Codec, xccResolver xcctypes.XCCResolver, txs []ContractTransaction) (*Linker, error) {
	lkr := Linker{cdc: cdc, crossChainChannelResolver: xccResolver, objects: make(map[crosstypes.TxIndex]lazyObject, len(txs))}
	for i, tx := range txs {
		idx := crosstypes.TxIndex(i)
		tx := tx
		lkr.objects[idx] = makeLazyObject(func() returnObject {
			if tx.ReturnValue.IsNil() {
				return returnObject{err: errors.New("On cross-chain call, each contractTransaction must be specified a return-value")}
			}
			xcc, err := tx.GetCrossChainChannel(lkr.cdc)
			if err != nil {
				return returnObject{err: err}
			}
			obj := txtypes.MakeConstantValueObject(xcc, MakeObjectKey(tx.CallInfo, tx.Signers), tx.ReturnValue.Value)
			return returnObject{obj: &obj}
		})
	}
	return &lkr, nil
}

// Resolve resolves given links and returns resolved Object
func (lkr Linker) Resolve(ctx sdk.Context, callerc xcctypes.XCC, lks []Link) ([]txtypes.Object, error) {
	var objects []txtypes.Object
	for _, lk := range lks {
		idx := lk.GetSrcIndex()
		lzObj, ok := lkr.objects[idx]
		if !ok {
			return nil, fmt.Errorf("idx '%v' not found", idx)
		}
		ret := lzObj()
		if ret.err != nil {
			return nil, ret.err
		}
		calleeID := ret.obj.GetCrossChainChannel(lkr.cdc)
		xcc, err := lkr.crossChainChannelResolver.ConvertCrossChainChannel(ctx, calleeID, callerc)
		if err != nil {
			return nil, err
		}
		obj := ret.obj.WithCrossChainChannel(lkr.cdc, xcc)
		objects = append(objects, obj)
	}
	return objects, nil
}

type returnObject struct {
	obj txtypes.Object
	err error
}

type lazyObject func() returnObject

func makeLazyObject(f func() returnObject) lazyObject {
	var v returnObject
	var once sync.Once
	return func() returnObject {
		once.Do(func() {
			v = f()
			f = nil // so that f can now be GC'ed
		})
		return v
	}
}

// MakeObjectKey returns a key that can be used to identify a contract call
func MakeObjectKey(callInfo txtypes.ContractCallInfo, signers []authtypes.AccountID) []byte {
	h := tmhash.New()
	h.Write(callInfo)
	for _, signer := range signers {
		h.Write(signer)
	}
	return h.Sum(nil)
}
