package cross

import (
	"github.com/datachainlab/cross/x/ibc/cross/internal/keeper"
	"github.com/datachainlab/cross/x/ibc/cross/internal/types"
)

// nolint
const (
	ModuleName = types.ModuleName
	RouterKey  = types.RouterKey
	StoreKey   = types.StoreKey

	CO_STATUS_NONE    = keeper.CO_STATUS_NONE
	CO_STATUS_INIT    = keeper.CO_STATUS_INIT
	CO_STATUS_DECIDED = keeper.CO_STATUS_DECIDED
	CO_STATUS_COMMIT  = keeper.CO_STATUS_COMMIT

	CO_DECISION_NONE   = keeper.CO_DECISION_NONE
	CO_DECISION_COMMIT = keeper.CO_DECISION_COMMIT
	CO_DECISION_ABORT  = keeper.CO_DECISION_ABORT

	TX_STATUS_PREPARE = keeper.TX_STATUS_PREPARE
	TX_STATUS_COMMIT  = keeper.TX_STATUS_COMMIT
	TX_STATUS_ABORT   = keeper.TX_STATUS_ABORT

	PREPARE_STATUS_FAILED = types.PREPARE_STATUS_FAILED
	PREPARE_STATUS_OK     = types.PREPARE_STATUS_OK
)

// nolint
var (
	NewKeeper                  = keeper.NewKeeper
	NewQuerier                 = keeper.NewQuerier
	ModuleCdc                  = types.ModuleCdc
	RegisterCodec              = types.RegisterCodec
	SignersFromContext         = types.SignersFromContext
	WithSigners                = types.WithSigners
	NewMsgInitiate             = types.NewMsgInitiate
	NewContractTransaction     = types.NewContractTransaction
	NewChannelInfo             = types.NewChannelInfo
	NewPacketDataPrepare       = types.NewPacketDataPrepare
	NewPacketDataPrepareResult = types.NewPacketDataPrepareResult
	NewPreparePacket           = types.NewPreparePacket
	NewPacketDataCommit        = types.NewPacketDataCommit
)

// nolint
type (
	Keeper                  = keeper.Keeper
	ContractHandler         = keeper.ContractHandler
	MsgInitiate             = types.MsgInitiate
	PreparePacket           = types.PreparePacket
	PacketDataPrepare       = types.PacketDataPrepare
	PacketDataPrepareResult = types.PacketDataPrepareResult
	PacketDataCommit        = types.PacketDataCommit
	OP                      = types.OP
	OPs                     = types.OPs
	State                   = types.State
	Store                   = types.Store
	Committer               = types.Committer
	ChannelInfo             = types.ChannelInfo
	ContractTransaction     = types.ContractTransaction
	ContractTransactions    = types.ContractTransactions
)
