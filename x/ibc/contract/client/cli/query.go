package cli

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/datachainlab/cross/x/ibc/contract/internal/types"
	"github.com/datachainlab/cross/x/ibc/cross"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetQueryCmd returns the transaction commands for this module
func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the contract module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(GetSimulationCmd(cdc))
	return cmd
}

func GetSimulationCmd(cdc *codec.Codec) *cobra.Command {
	const (
		flagSave = "save"
	)

	cmd := &cobra.Command{
		Use:   "call [contract_id] [contract_method] [[contract_arg_hex]...]",
		Short: "simulate a contract transaction",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)
			var cargs [][]byte
			for _, a := range args[2:] {
				if strings.HasPrefix(a, "0x") {
					b, err := hex.DecodeString(a)
					if err != nil {
						return err
					}
					cargs = append(cargs, b)
				} else {
					cargs = append(cargs, []byte(a))
				}
			}
			ci := types.NewContractCallInfo(
				args[0],
				args[1],
				cargs,
			)
			msg := types.NewMsgContractCall(
				cliCtx.GetFromAddress(),
				nil,
				ci.Bytes(),
			)
			bz, err := cdc.MarshalBinaryLengthPrefixed(msg)
			if err != nil {
				return err
			}
			route := fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QuerySimulation)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}
			var (
				ops    cross.OPs
				result sdk.Result
			)
			cdc.MustUnmarshalBinaryLengthPrefixed(res, &result)
			cdc.MustUnmarshalBinaryLengthPrefixed(result.Data, &ops)
			fmt.Println(ops.String())
			return ioutil.WriteFile(viper.GetString(flagSave), result.Data, 0644)
		},
	}
	cmd = flags.PostCommands(cmd)[0]
	cmd.Flags().String(flagSave, "", "Save a result to this file")
	cmd.MarkFlagRequired(flags.FlagFrom)
	cmd.MarkFlagRequired(flagSave)
	return cmd
}
