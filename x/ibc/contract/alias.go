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
	HTTPServerRouter     = types.HTTPServerRouter
	MsgContractCall      = types.MsgContractCall
	ContractCallResponse = types.ContractCallResponse
)

var (
	NewKeeper              = keeper.NewKeeper
	NewQuerier             = keeper.NewQuerier
	NewContractHandler     = keeper.NewContractHandler
	NewContract            = keeper.NewContract
	CallExternalFunc       = keeper.CallExternalFunc
	NewHTTPServerRouter    = types.NewHTTPServerRouter
	EncodeContractCallInfo = types.EncodeContractCallInfo
	DecodeContractCallInfo = types.DecodeContractCallInfo
	NewContractCallInfo    = types.NewContractCallInfo
	NewMsgContractCall     = types.NewMsgContractCall

	ModuleCdc     = types.ModuleCdc
	RegisterCodec = types.RegisterCodec
)
