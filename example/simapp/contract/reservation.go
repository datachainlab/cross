package contract

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/ibc/contract"
	"github.com/datachainlab/cross/x/ibc/cross"
	"github.com/datachainlab/cross/x/ibc/store/lock"
)

const (
	ReserveFnName   = "reserve"
	TrainContractID = "train"
	HotelContractID = "hotel"
)

func HotelReservationContractHandler(k contract.Keeper) cross.ContractHandler {
	contractHandler := contract.NewContractHandler(k, func(store sdk.KVStore) cross.State {
		return lock.NewStore(store)
	})

	contractHandler.AddRoute(HotelContractID, GetHotelContract())
	return contractHandler
}

func TrainReservationContractHandler(k contract.Keeper) cross.ContractHandler {
	contractHandler := contract.NewContractHandler(k, func(store sdk.KVStore) cross.State {
		return lock.NewStore(store)
	})

	contractHandler.AddRoute(TrainContractID, GetTrainContract())
	return contractHandler
}

func MakeRoomKey(id int32) []byte {
	return []byte(fmt.Sprintf("room/%v", id))
}

func GetHotelContract() contract.Contract {
	return contract.NewContract([]contract.Method{
		{
			Name: ReserveFnName,
			F: func(ctx contract.Context, store cross.Store) error {
				reserver := ctx.Signers()[0]
				roomID := contract.Int32(ctx.Args()[0])
				key := MakeRoomKey(roomID)
				if store.Has(key) {
					return fmt.Errorf("room %v is already reserved", roomID)
				} else {
					store.Set(key, reserver)
				}
				return nil
			},
		},
	})
}

func MakeSeatKey(id int32) []byte {
	return []byte(fmt.Sprintf("seat/%v", id))
}

func GetTrainContract() contract.Contract {
	return contract.NewContract([]contract.Method{
		{
			Name: ReserveFnName,
			F: func(ctx contract.Context, store cross.Store) error {
				reserver := ctx.Signers()[0]
				seatID := contract.Int32(ctx.Args()[0])
				key := MakeSeatKey(seatID)
				if store.Has(key) {
					return fmt.Errorf("seat %v is already reserved", seatID)
				} else {
					store.Set(key, reserver)
				}
				return nil
			},
		},
	})
}
