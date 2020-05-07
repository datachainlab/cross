package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/datachainlab/cross/x/ibc/cross/types"
	"github.com/gorilla/mux"
)

func QueryCoordinatorStatusHandlerFn(ctx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, ctx, r)
		if !ok {
			return
		}
		txIDStr := mux.Vars(r)["tx_id"]
		var txID types.TxID
		if strings.HasPrefix(txIDStr, "0x") {
			txID.FromString(txIDStr[2:])
		} else {
			txID.FromString(txIDStr)
		}

		req := types.QueryCoordinatorStatusRequest{TxID: txID}
		route := fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryCoordinatorStatus)
		res, _, err := ctx.QueryWithData(route, ctx.Codec.MustMarshalJSON(req))
		if rest.CheckInternalServerError(w, err) {
			return
		}
		var response types.QueryCoordinatorStatusResponse
		ctx.Codec.MustUnmarshalJSON(res, &response)
		rest.PostProcessResponse(w, ctx, response)
	}
}

func QueryUnacknowledgedPacketsHandlerFn(ctx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, ctx, r)
		if !ok {
			return
		}
		req := types.QueryUnacknowledgedPacketsRequest{}
		route := fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryUnacknowledgedPackets)
		res, _, err := ctx.QueryWithData(route, ctx.Codec.MustMarshalJSON(req))
		if rest.CheckInternalServerError(w, err) {
			return
		}
		var response types.QueryUnacknowledgedPacketsResponse
		ctx.Codec.MustUnmarshalJSON(res, &response)
		rest.PostProcessResponse(w, ctx, response)
	}
}
