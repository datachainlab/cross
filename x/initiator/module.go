package initiator

import (
	"github.com/datachainlab/cross/x/initiator/client/cli"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the root tx command for the IBC connections.
func GetTxCmd() *cobra.Command {
	return cli.NewTxCmd()
}

// GetQueryCmd returns the root query command for the IBC connections.
func GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}
