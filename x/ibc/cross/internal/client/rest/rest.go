package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
)

const (
// RestChannelID = "channel-id"
// RestPortID    = "port-id"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, queryRoute string) {
	// registerQueryRoutes(cliCtx, r, queryRoute)
	// registerTxRoutes(cliCtx, r)
}
