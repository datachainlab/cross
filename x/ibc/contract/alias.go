package contract

import (
	"github.com/datachainlab/cross/x/ibc/contract/keeper"
	"github.com/datachainlab/cross/x/ibc/contract/types"
)

const (
	ModuleName = types.ModuleName
	StoreKey   = types.StoreKey
	RouterKey  = types.RouterKey
)

type (
	Keeper               = keeper.Keeper
	Method               = keeper.Method
	StateProvider        = keeper.StateProvider
	Context              = keeper.Context
	Contract             = keeper.Contract
	MsgContractCall      = types.MsgContractCall
	ContractCallResponse = types.ContractCallResponse
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
