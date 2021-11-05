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
	"github.com/gogo/protobuf/proto"
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
			res := txtypes.NewConstantValueCallResult(xcc, MakeCallResultKey(tx.CallInfo, tx.Signers), tx.ReturnValue.Value)
			return returnObject{res: &res}
		})
	}
	return &lkr, nil
}

// Resolve resolves given links and returns resolved Object
func (lkr Linker) Resolve(ctx sdk.Context, callerc xcctypes.XCC, lks []Link) ([]txtypes.CallResult, error) {
	var results []txtypes.CallResult
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
		calleeID := ret.res.GetCrossChainChannel(lkr.cdc)
		xcc, err := lkr.crossChainChannelResolver.ConvertCrossChainChannel(ctx, calleeID, callerc)
		if err != nil {
			return nil, err
		}
		result := ret.res.WithCrossChainChannel(lkr.cdc, xcc)
		results = append(results, result)
	}
	return results, nil
}

type returnObject struct {
	res txtypes.CallResult
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

// MakeCallResultKey returns a key that can be used to identify a contract call
func MakeCallResultKey(callInfo txtypes.ContractCallInfo, signers []authtypes.Account) []byte {
	h := tmhash.New()
	h.Write(callInfo)
	for _, signer := range signers {
		bz, err := proto.Marshal(&signer)
		if err != nil {
			panic(err)
		}
		h.Write(bz)
	}
	return h.Sum(nil)
}
