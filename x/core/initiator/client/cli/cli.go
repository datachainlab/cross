package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/datachainlab/cross/x/core/initiator/types"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	// Group bridge queries under a subcommand
	queryCmd := &cobra.Command{
		Use:                        types.SubModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.SubModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	queryCmd.AddCommand(
		GetCreateContractTransaction(),
	)

	return queryCmd
}

// NewTxCmd returns the transaction commands for IBC fungible token transfer
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.SubModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.SubModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewInitiateTxCmd(),
		NewIBCSignTxCmd(),
	)

	return txCmd
}
