package contract

import (
	"github.com/bluele/crossccc/x/ibc/contract/internal/keeper"
	"github.com/bluele/crossccc/x/ibc/contract/internal/types"
)

const (
	ModuleName = types.ModuleName
	StoreKey   = types.StoreKey
	RouterKey  = types.RouterKey
)

type (
	Keeper          = keeper.Keeper
	MsgContractCall = types.MsgContractCall
)

var (
	NewKeeper          = keeper.NewKeeper
	NewQuerier         = keeper.NewQuerier
	NewMsgContractCall = types.NewMsgContractCall
)
