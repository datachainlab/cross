package simapp

import (
	"encoding/binary"
	"fmt"

	"github.com/bluele/crossccc/x/ibc/contract"
	"github.com/bluele/crossccc/x/ibc/crossccc"
	"github.com/bluele/crossccc/x/ibc/store/lock"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func makeContractHandler(k contract.Keeper) crossccc.ContractHandler {
	contractHandler := contract.NewContractHandler(k, func(store sdk.KVStore) crossccc.State {
		return lock.NewStore(store)
	})

	c := contract.NewContract([]contract.Method{
		{
			Name: "f0",
			F: func(ctx contract.Context, store crossccc.Store) error {
				k := []byte("counter")
				v := store.Get(k)
				var counter uint32
				if v == nil {
					counter = 0
				} else {
					counter = binary.BigEndian.Uint32(v)
				}

				fmt.Printf("f0 is called by %v counter=%v\n", ctx.Signers()[0].String(), counter)

				b := make([]byte, 4)
				binary.BigEndian.PutUint32(b, counter+1)

				store.Set(k, b)
				return nil
			},
		},
	})
	contractHandler.AddRoute("c0", c)
	return contractHandler
}
