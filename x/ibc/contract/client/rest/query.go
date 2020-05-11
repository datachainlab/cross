package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/datachainlab/cross/x/ibc/contract/types"
	"github.com/datachainlab/cross/x/ibc/cross"
)

type ContractCallReq struct {
	From     string                 `json:"from"`
	Signers  []string               `json:"signers"`
	CallInfo types.ContractCallInfo `json:"call_info"`
}

func QueryContractCallRequestHandlerFn(ctx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ContractCallReq
		if !rest.ReadRESTReq(w, r, ctx.Codec, &req) {
			return
		}
		var returnResult bool
		v := r.FormValue("result")
		if v == "true" {
			returnResult = true
		} else {
			returnResult = false
		}

		ctx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, ctx, r)
		if !ok {
			return
		}
		addr, err := sdk.AccAddressFromBech32(req.From)
		if rest.CheckInternalServerError(w, err) {
			return
		}
		var signers []sdk.AccAddress
		for _, s := range req.Signers {
			signer, err := sdk.AccAddressFromBech32(s)
			if rest.CheckInternalServerError(w, err) {
				return
			}
			signers = append(signers, signer)
		}
		msg := types.NewMsgContractCall(
			addr,
			signers,
			req.CallInfo.Bytes(),
			cross.ExactStateCondition,
		)
		bz, err := ctx.Codec.MarshalJSON(msg)
		if rest.CheckBadRequestError(w, err) {
			return
		}
		route := fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QuerySimulation)
		res, height, err := ctx.QueryWithData(route, bz)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		var (
			response types.ContractCallResponse
			result   sdk.Result
		)
		ctx.Codec.MustUnmarshalJSON(res, &result)
		ctx.Codec.MustUnmarshalJSON(result.Data, &response)

		ctx = ctx.WithHeight(height)
		if returnResult {
			rest.PostProcessResponse(w, ctx, result)
		} else {
			rest.PostProcessResponse(w, ctx, response)
		}
	}
}
