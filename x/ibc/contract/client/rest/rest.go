package rest

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/gorilla/mux"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(ctx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/cross/contract/call", QueryContractCallRequestHandlerFn(ctx)).Methods("POST")
}
