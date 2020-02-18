package cli

import (
	"bufio"
	"encoding/hex"

	"github.com/bluele/cross/x/ibc/contract/internal/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
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
		Use:   "call [from_key_or_address] [contract_id] [contract_method] [contract_arg_hex]",
		Short: "Create and sign a send tx",
		Args:  cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInputAndFrom(inBuf, args[0]).WithCodec(cdc)
			sender := cliCtx.GetFromAddress()
			var cargs [][]byte
			for _, a := range args[3:] {
				b, err := hex.DecodeString(a)
				if err != nil {
					return err
				}
				cargs = append(cargs, b)
			}
			ci := types.NewContractInfo(
				args[1],
				args[2],
				cargs,
			)
			msg := types.NewMsgContractCall(
				sender,
				nil,
				ci.Bytes(),
			)
			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd = flags.PostCommands(cmd)[0]
	return cmd
}
