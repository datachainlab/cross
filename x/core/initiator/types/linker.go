package types

import (
	"errors"
	"fmt"
	"sync"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	contracttypes "github.com/datachainlab/cross/x/contract/types"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
)

// Linker resolves links that each ContractTransaction has.
type Linker struct {
	cdc                       codec.Marshaler
	crossChainChannelResolver xcctypes.XCCResolver
	objects                   map[txtypes.TxIndex]lazyObject
}

// MakeLinker returns Linker
func MakeLinker(cdc codec.Marshaler, xccResolver xcctypes.XCCResolver, txs []ContractTransaction) (*Linker, error) {
	lkr := Linker{cdc: cdc, crossChainChannelResolver: xccResolver, objects: make(map[txtypes.TxIndex]lazyObject, len(txs))}
	for i, tx := range txs {
		idx := txtypes.TxIndex(i)
		tx := tx
		lkr.objects[idx] = makeLazyObject(func() returnObject {
			if tx.ReturnValue.IsNil() {
				return returnObject{err: errors.New("On cross-chain call, each contractTransaction must be specified a return-value")}
			}
			xcc, err := tx.GetCrossChainChannel(lkr.cdc)
			if err != nil {
				return returnObject{err: err}
			}
			obj := txtypes.MakeConstantValueObject(xcc, contracttypes.MakeObjectKey(tx.CallInfo, tx.Signers), tx.ReturnValue.Value)
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
