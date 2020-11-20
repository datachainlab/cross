package cli

import (
	"context"
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/datachainlab/cross/x/atomic/types"
	"github.com/spf13/cobra"
)

func GetCoordinatorState() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "coordinator-state [TxID: hex encoding]",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			txID, err := hex.DecodeString(args[0])
			if err != nil {
				return err
			}
			q := types.NewQueryClient(clientCtx)
			res, err := q.CoordinatorState(
				context.Background(),
				&types.QueryCoordinatorStateRequest{TxId: txID},
			)
			if err != nil {
				return err
			}
			return clientCtx.PrintOutput(&res.CoodinatorState)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
