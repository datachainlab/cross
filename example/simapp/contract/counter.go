package contract

import (
	"encoding/binary"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/ibc/contract"
	"github.com/datachainlab/cross/x/ibc/cross"
	"github.com/datachainlab/cross/x/ibc/store/lock"
)

func CounterContractHandlerProvider(k contract.Keeper) cross.ContractHandler {
	contractHandler := contract.NewContractHandler(k, func(store sdk.KVStore) cross.State {
		return lock.NewStore(store)
	})

	contractHandler.AddRoute(
		"c0",
		contract.NewContract(
			[]contract.Method{
				{
					Name: "f0",
					F: func(ctx contract.Context, store cross.Store) ([]byte, error) {
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
						counter++
						binary.BigEndian.PutUint32(b, counter)
						store.Set(k, b)

						ctx.EventManager().EmitEvent(
							sdk.NewEvent("counter", sdk.NewAttribute("count", fmt.Sprint(counter))),
						)
						return b, nil
					},
				},
			},
		),
	)
	return contractHandler
}
