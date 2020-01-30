package crossccc

import (
	"github.com/bluele/crossccc/x/ibc/crossccc/internal/keeper"
	"github.com/bluele/crossccc/x/ibc/crossccc/internal/types"
)

// nolint
const (
	ModuleName = types.ModuleName
	RouterKey  = types.RouterKey
	StoreKey   = types.StoreKey

	CO_STATUS_INIT   = keeper.CO_STATUS_INIT
	CO_STATUS_COMMIT = keeper.CO_STATUS_COMMIT

	TX_STATUS_PREPARE = keeper.TX_STATUS_PREPARE
	TX_STATUS_COMMIT  = keeper.TX_STATUS_COMMIT
	TX_STATUS_ABORT   = keeper.TX_STATUS_ABORT
)

// nolint
var (
	NewKeeper          = keeper.NewKeeper
	NewQuerier         = keeper.NewQuerier
	ModuleCdc          = types.ModuleCdc
	RegisterCodec      = types.RegisterCodec
	SignerFromContext  = types.SignerFromContext
	WithSigner         = types.WithSigner
	NewMsgInitiate     = types.NewMsgInitiate
	NewMsgConfirm      = types.NewMsgConfirm
	NewStateTransition = types.NewStateTransition
	NewChannelInfo     = types.NewChannelInfo
)

// nolint
type (
	Keeper             = keeper.Keeper
	ContractHandler    = keeper.ContractHandler
	MsgInitiate        = types.MsgInitiate
	MsgConfirm         = types.MsgConfirm
	PacketDataInitiate = types.PacketDataInitiate
	PacketDataCommit   = types.PacketDataCommit
	OP                 = types.OP
	OPs                = types.OPs
	State              = types.State
	Store              = types.Store
	Committer          = types.Committer
	ChannelInfo        = types.ChannelInfo
	StateTransition    = types.StateTransition
)
