package cli

import (
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
)

func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:   "cross",
		Short: "Cross chain contract calls query subcommands",
	}

	queryCmd.AddCommand(flags.PostCommands()...)

	return queryCmd
}
