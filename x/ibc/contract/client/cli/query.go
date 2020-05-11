package cli

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/datachainlab/cross/x/ibc/contract/types"
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
	cmd.AddCommand(
		GetContractCallSimulationCmd(cdc),
		GetShowContractCallResult(cdc),
	)
	return cmd
}

func GetContractCallSimulationCmd(cdc *codec.Codec) *cobra.Command {
	const (
		flagStateConstraintType = "state-constraint"
		flagSave                = "save"
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
					b, err := hex.DecodeString(a[2:])
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
			scType := uint8(viper.GetUint(flagStateConstraintType))
			msg := types.NewMsgContractCall(
				cliCtx.GetFromAddress(),
				nil,
				ci.Bytes(),
				scType,
			)
			bz, err := cdc.MarshalJSON(msg)
			if err != nil {
				return err
			}
			route := fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QuerySimulation)
			res, height, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}
			var (
				response types.ContractCallResponse
				result   sdk.Result
			)
			cdc.MustUnmarshalJSON(res, &result)
			cdc.MustUnmarshalJSON(result.Data, &response)
			callResult := cross.ContractCallResult{
				ChainID:         cliCtx.ChainID,
				Height:          height,
				Signers:         []sdk.AccAddress{cliCtx.GetFromAddress()},
				CallInfo:        ci.Bytes(),
				StateConstraint: cross.NewStateConstraint(scType, response.OPs),
			}
			bz, err = cdc.MarshalJSON(callResult)
			if err != nil {
				return err
			}
			fmt.Println(callResult.String())
			return ioutil.WriteFile(viper.GetString(flagSave), bz, 0644)
		},
	}
	cmd = flags.PostCommands(cmd)[0]
	cmd.Flags().Uint8(flagStateConstraintType, cross.ExactMatchStateConstraint, "state constraint type")
	cmd.Flags().String(flagSave, "", "Save a result to this file")
	cmd.MarkFlagRequired(flags.FlagFrom)
	cmd.MarkFlagRequired(flagSave)
	return cmd
}

func GetShowContractCallResult(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-result [path to result file]",
		Short: "show result saved in file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bz, err := ioutil.ReadFile(args[0])
			if err != nil {
				return err
			}
			var res cross.ContractCallResult
			if err := cdc.UnmarshalJSON(bz, &res); err != nil {
				return err
			}
			fmt.Println(res.String())
			return nil
		},
	}
	cmd = flags.PostCommands(cmd)[0]
	return cmd
}
