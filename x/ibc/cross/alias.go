package cross

import (
	"github.com/datachainlab/cross/x/ibc/cross/keeper"
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

	TX_STATUS_PREPARE = types.TX_STATUS_PREPARE
	TX_STATUS_COMMIT  = types.TX_STATUS_COMMIT
	TX_STATUS_ABORT   = types.TX_STATUS_ABORT

	PREPARE_RESULT_OK     = types.PREPARE_RESULT_OK
	PREPARE_RESULT_FAILED = types.PREPARE_RESULT_FAILED

	TypeInitiate      = types.TypeInitiate
	TypePrepare       = types.TypePrepare
	TypePrepareResult = types.TypePrepareResult
	TypeCommit        = types.TypeCommit

	NoStateConstraint         = types.NoStateConstraint
	ExactMatchStateConstraint = types.ExactMatchStateConstraint
	PreStateConstraint        = types.PreStateConstraint
	PostStateConstraint       = types.PostStateConstraint
)

// nolint
var (
	NewKeeper                       = keeper.NewKeeper
	NewQuerier                      = keeper.NewQuerier
	MakeTxID                        = keeper.MakeTxID
	MakeStoreTransactionID          = keeper.MakeStoreTransactionID
	ModuleCdc                       = types.ModuleCdc
	RegisterCodec                   = types.RegisterCodec
	SignersFromContext              = types.SignersFromContext
	WithSigners                     = types.WithSigners
	NewMsgInitiate                  = types.NewMsgInitiate
	NewContractTransaction          = types.NewContractTransaction
	NewStateConstraint              = types.NewStateConstraint
	NewChannelInfo                  = types.NewChannelInfo
	NewPacketDataPrepare            = types.NewPacketDataPrepare
	NewPacketPrepareAcknowledgement = types.NewPacketPrepareAcknowledgement
	NewPacketDataCommit             = types.NewPacketDataCommit
	NewPacketCommitAcknowledgement  = types.NewPacketCommitAcknowledgement
)

// nolint
type (
	Keeper                       = keeper.Keeper
	ContractHandler              = types.ContractHandler
	ContractHandlerResult        = types.ContractHandlerResult
	MsgInitiate                  = types.MsgInitiate
	PacketData                   = types.PacketData
	PacketAcknowledgement        = types.PacketAcknowledgement
	PacketDataPrepare            = types.PacketDataPrepare
	PacketPrepareAcknowledgement = types.PacketPrepareAcknowledgement
	PacketDataCommit             = types.PacketDataCommit
	PacketCommitAcknowledgement  = types.PacketCommitAcknowledgement
	OP                           = types.OP
	OPs                          = types.OPs
	State                        = types.State
	Store                        = types.Store
	Committer                    = types.Committer
	ChannelInfo                  = types.ChannelInfo
	ContractTransaction          = types.ContractTransaction
	ContractTransactions         = types.ContractTransactions
	ContractCallResult           = types.ContractCallResult
	ContractCallInfo             = types.ContractCallInfo
	StateConstraint              = types.StateConstraint
	StateConstraintType          = types.StateConstraintType
	TxID                         = types.TxID
	TxIndex                      = types.TxIndex
)
