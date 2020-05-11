package cli

import (
	"bufio"
	"encoding/hex"
	"strings"

	"github.com/datachainlab/cross/x/ibc/contract/types"
	"github.com/datachainlab/cross/x/ibc/cross"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Contract transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(
		CallTxCmd(cdc),
	)
	return txCmd
}

func CallTxCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "call [contract_id] [contract_method] [[contract_arg_hex]...]",
		Short: "Create and sign a send tx",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
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
			msg := types.NewMsgContractCall(
				cliCtx.GetFromAddress(),
				nil,
				ci.Bytes(),
				cross.NoStateConstraint,
			)
			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd = flags.PostCommands(cmd)[0]
	cmd.MarkFlagRequired(flags.FlagFrom)
	return cmd
}
