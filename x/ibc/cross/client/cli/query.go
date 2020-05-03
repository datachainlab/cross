package cli

import (
	"bufio"
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/datachainlab/cross/x/ibc/cross/types"
	"github.com/spf13/cobra"
)

func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for Cross module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(
		flags.PostCommands(GetCoordinatorStatus(cdc), GetUnacknowledgedPackets(cdc))...,
	)
	return cmd
}

func GetCoordinatorStatus(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "coordinator [tx_id]",
		Short: "get the coordinator's status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)
			bz, err := hex.DecodeString(args[0])
			if err != nil {
				return err
			}
			var txID types.TxID
			var response types.QueryCoordinatorStatusResponse

			if len(bz) != len(txID) {
				return fmt.Errorf("expected length is %v, but got %v", len(txID), len(bz))
			}
			copy(txID[:], bz)
			req := types.QueryCoordinatorStatusRequest{TxID: txID}
			bz, err = cdc.MarshalBinaryLengthPrefixed(req)
			if err != nil {
				return err
			}
			route := fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryCoordinatorStatus)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}
			cdc.MustUnmarshalBinaryLengthPrefixed(res, &response)
			fmt.Println(string(cdc.MustMarshalJSON(response)))
			return nil
		},
	}

	return cmd
}

func GetUnacknowledgedPackets(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unacknowledged-packets",
		Short: "get all unacknowledged packets",
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			req := types.QueryUnacknowledgedPacketsRequest{}
			bz, err := cdc.MarshalBinaryLengthPrefixed(req)
			if err != nil {
				return err
			}
			route := fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryUnacknowledgedPackets)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}
			var response types.QueryUnacknowledgedPacketsResponse
			cdc.MustUnmarshalBinaryLengthPrefixed(res, &response)
			fmt.Println(string(cdc.MustMarshalJSON(response)))
			return nil
		},
	}

	return cmd
}
