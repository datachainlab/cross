package contract

import (
	"github.com/datachainlab/cross/x/ibc/contract/internal/keeper"
	"github.com/datachainlab/cross/x/ibc/contract/internal/types"
)

const (
	ModuleName = types.ModuleName
	StoreKey   = types.StoreKey
	RouterKey  = types.RouterKey
)

type (
	Keeper           = keeper.Keeper
	Method           = keeper.Method
	StateProvider    = keeper.StateProvider
	Context          = keeper.Context
	MsgContractCall  = types.MsgContractCall
	ContractResponse = types.ContractResponse
)

var (
	NewKeeper               = keeper.NewKeeper
	NewQuerier              = keeper.NewQuerier
	NewContractHandler      = keeper.NewContractHandler
	EncodeContractSignature = types.EncodeContractSignature
	DecodeContractSignature = types.DecodeContractSignature
	NewContractCallInfo     = types.NewContractCallInfo
	NewMsgContractCall      = types.NewMsgContractCall
	NewContract             = keeper.NewContract

	ModuleCdc     = types.ModuleCdc
	RegisterCodec = types.RegisterCodec
)
