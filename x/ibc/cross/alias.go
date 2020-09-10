package cross

import (
	"github.com/datachainlab/cross/x/ibc/cross/keeper"
	"github.com/datachainlab/cross/x/ibc/cross/keeper/common"
	"github.com/datachainlab/cross/x/ibc/cross/types"
)

// nolint
const (
	ModuleName   = types.ModuleName
	RouterKey    = types.RouterKey
	StoreKey     = types.StoreKey
	QuerierRoute = types.QuerierRoute

	CO_STATUS_NONE    = types.CO_STATUS_NONE
	CO_STATUS_INIT    = types.CO_STATUS_INIT
	CO_STATUS_DECIDED = types.CO_STATUS_DECIDED

	CO_DECISION_NONE   = types.CO_DECISION_NONE
	CO_DECISION_COMMIT = types.CO_DECISION_COMMIT
	CO_DECISION_ABORT  = types.CO_DECISION_ABORT

	COMMIT_PROTOCOL_SIMPLE = types.COMMIT_PROTOCOL_SIMPLE
	COMMIT_PROTOCOL_TPC    = types.COMMIT_PROTOCOL_TPC

	TX_STATUS_PREPARE = types.TX_STATUS_PREPARE
	TX_STATUS_COMMIT  = types.TX_STATUS_COMMIT
	TX_STATUS_ABORT   = types.TX_STATUS_ABORT

	TypeInitiate = types.TypeInitiate

	NoStateConstraint         = types.NoStateConstraint
	ExactMatchStateConstraint = types.ExactMatchStateConstraint
	PreStateConstraint        = types.PreStateConstraint
	PostStateConstraint       = types.PostStateConstraint
)

// nolint
var (
	NewKeeper               = keeper.NewKeeper
	NewQuerier              = keeper.NewQuerier
	MakeTxID                = common.MakeTxID
	MakeStoreTransactionID  = common.MakeStoreTransactionID
	ModuleCdc               = types.ModuleCdc
	RegisterCodec           = types.RegisterCodec
	SignersFromContext      = types.SignersFromContext
	WithSigners             = types.WithSigners
	NewMsgInitiate          = types.NewMsgInitiate
	NewContractTransaction  = types.NewContractTransaction
	NewStateConstraint      = types.NewStateConstraint
	NewReturnValue          = types.NewReturnValue
	NewChannelInfo          = types.NewChannelInfo
	NewCallResultLink       = types.NewCallResultLink
	MakeObjectKey           = types.MakeObjectKey
	DefaultResolverProvider = types.DefaultResolverProvider
	NewFakeResolver         = types.NewFakeResolver
)

// nolint
type (
	Keeper                  = keeper.Keeper
	ContractHandler         = types.ContractHandler
	ContractHandlerResult   = types.ContractHandlerResult
	MsgInitiate             = types.MsgInitiate
	PacketData              = types.PacketData
	PacketAcknowledgement   = types.PacketAcknowledgement
	OP                      = types.OP
	OPs                     = types.OPs
	State                   = types.State
	Store                   = types.Store
	Committer               = types.Committer
	ChannelInfo             = types.ChannelInfo
	ContractTransaction     = types.ContractTransaction
	ContractTransactions    = types.ContractTransactions
	ContractTransactionInfo = types.ContractTransactionInfo
	ContractCallResult      = types.ContractCallResult
	ContractCallInfo        = types.ContractCallInfo
	ContractRuntimeInfo     = types.ContractRuntimeInfo
	StateConstraint         = types.StateConstraint
	StateConstraintType     = types.StateConstraintType
	TxID                    = types.TxID
	TxIndex                 = types.TxIndex
	ObjectResolver          = types.ObjectResolver
	ObjectResolverProvider  = types.ObjectResolverProvider
	SequentialResolver      = types.SequentialResolver
	FakeResolver            = types.FakeResolver
	Object                  = types.Object
	ObjectType              = types.ObjectType
	Link                    = types.Link
	LinkType                = types.LinkType
)
