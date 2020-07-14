package keeper_test

import (
	"fmt"
	"strconv"

	"github.com/datachainlab/cross/x/ibc/contract"
	"github.com/datachainlab/cross/x/ibc/cross"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func parseCoin(ctx contract.Context, denomIdx, amountIdx int) (sdk.Coin, error) {
	denom := string(ctx.Args()[denomIdx])
	amount, err := strconv.Atoi(string(ctx.Args()[amountIdx]))
	if err != nil {
		return sdk.Coin{}, err
	}
	if amount < 0 {
		return sdk.Coin{}, fmt.Errorf("amount must be positive number")
	}
	coin := sdk.NewInt64Coin(denom, int64(amount))
	return coin, nil
}

func marshalCoin(coin sdk.Coin) []byte {
	bz, err := coin.Marshal()
	if err != nil {
		panic(err)
	}
	return bz
}

func marshalCoins(coins sdk.Coins) []byte {
	return testcdc.MustMarshalBinaryLengthPrefixed(coins)
}

func unmarshalCoins(bz []byte) sdk.Coins {
	var coins sdk.Coins
	testcdc.MustUnmarshalBinaryLengthPrefixed(bz, &coins)
	return coins
}

func getBalanceOf(store cross.Store, address sdk.AccAddress) sdk.Coins {
	bz := store.Get(address)
	if bz == nil {
		return sdk.NewCoins()
	}
	return unmarshalCoins(bz)
}

func setBalance(store cross.Store, address sdk.AccAddress, balance sdk.Coins) {
	bz := marshalCoins(balance)
	store.Set(address, bz)
}

var testcdc *codec.Codec

func init() {
	testcdc = codec.New()

	testcdc.RegisterConcrete(sdk.Coin{}, "sdk/Coin", nil)
	testcdc.RegisterConcrete(sdk.Coins{}, "sdk/Coins", nil)
}
