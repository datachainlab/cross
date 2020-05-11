package cli

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/datachainlab/cross/x/ibc/cross/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagSigners  = "signers"
	flagOPs      = "ops"
	flagContract = "contract"
)

/*
GetCreateMsgInitiateCmd returns a command that executes to initiate a distributed transaction
This command implemetation follows under some assumptions.
Assumption:
	- All keys that are used to create a signature exists on this keychain
*/
func GetCreateMsgInitiateCmd(cdc *codec.Codec) *cobra.Command {
	const (
		flagContractTransaction = "contract"
		flagSourceChannel       = "channel"
	)

	cmd := &cobra.Command{
		Use:   "create [timeout-height] [nonce]",
		Short: "Create a MsgInitiate transaction",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc).WithBroadcastMode(flags.BroadcastBlock)
			sender := cliCtx.GetFromAddress()

			var txs []types.ContractTransaction
			txPaths := viper.GetStringSlice(flagContractTransaction)
			channels := viper.GetStringSlice(flagSourceChannel)
			if len(txPaths) != len(channels) {
				return fmt.Errorf("The number of contracts and channels don't match")
			}
			for i, path := range txPaths {
				var res types.ContractCallResult
				bz, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				cdc.MustUnmarshalJSON(bz, &res)
				channelInfo := channels[i]
				parts := strings.Split(channelInfo, ":")
				tx := types.ContractTransaction{
					Source: types.ChannelInfo{
						Channel: parts[0],
						Port:    parts[1],
					},
					Signers:         res.Signers,
					CallInfo:        res.CallInfo,
					StateConstraint: res.StateConstraint,
				}
				txs = append(txs, tx)
			}

			timeout, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			nonce, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}
			msg := types.NewMsgInitiate(
				sender,
				cliCtx.ChainID,
				txs,
				timeout,
				nonce,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd = flags.PostCommands(cmd)[0]
	cmd.Flags().StringSlice(flagContractTransaction, nil, "Save a result to this file")
	cmd.Flags().StringSlice(flagSourceChannel, nil, `"[channel]:[port]"`)
	cmd.MarkFlagRequired(flagContractTransaction)
	cmd.MarkFlagRequired(flagSourceChannel)
	cmd.MarkFlagRequired(flags.FlagFrom)
	cmd.MarkFlagRequired(flags.FlagChainID)
	return cmd
}
