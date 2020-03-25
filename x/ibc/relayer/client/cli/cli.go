package cli

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
)

func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "relayer",
		Short: "Relayer transaction subcommands",
	}

	txCmd.AddCommand(
		GetRelayPacket(cdc),
	)

	return txCmd
}
