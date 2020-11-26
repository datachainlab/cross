package cli

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	accounttypes "github.com/datachainlab/cross/x/account/types"
	"github.com/datachainlab/cross/x/core/initiator/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
)

func GetCreateContractTransaction() *cobra.Command {
	const (
		flagInitiatorChain        = "initiator-chain"
		flagInitiatorChainChannel = "initiator-chain-channel"
		flagSigners               = "signers"
		flagCallInfo              = "call-info"
	)

	// TODO add dry-run support
	cmd := &cobra.Command{
		Use:   "create-contract-tx",
		Short: "Create a new contract transaction",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			var anyXCC *codectypes.Any

			// Validations
			initiatorChannel := viper.GetString(flagInitiatorChainChannel)
			isInitiator := viper.GetBool(flagInitiatorChain)
			if isInitiator {
				// Query self-XCC to query server
				crossClient := types.NewQueryClient(clientCtx)
				res, err := crossClient.SelfXCC(context.Background(), &types.QuerySelfXCCRequest{})
				if err != nil {
					return err
				}
				anyXCC = res.Xcc
			} else {
				anyXCC, err = resolveXCCForInitiator(
					channeltypes.NewQueryClient(clientCtx),
					initiatorChannel,
				)
			}

			var signers []accounttypes.AccountID
			for _, s := range viper.GetStringSlice(flagSigners) {
				keyInfo, err := clientCtx.Keyring.Key(s)
				if err != nil {
					return err
				}
				signers = append(signers, accounttypes.AccountIDFromAccAddress(keyInfo.GetAddress()))
			}

			callInfo := []byte(viper.GetString(flagCallInfo))
			cTx := types.ContractTransaction{
				CrossChainChannel: anyXCC,
				Signers:           signers,
				CallInfo:          callInfo,
			}
			// prepare output document
			closeFunc, err := setOutputFile(cmd)
			if err != nil {
				return err
			}
			defer closeFunc()
			return clientCtx.WithOutput(cmd.OutOrStdout()).PrintOutput(&cTx)
		},
	}

	cmd.Flags().Bool(flagInitiatorChain, false, "A boolean value whether the chain is an initiator of the cross-chain tx includes this contract tx")
	cmd.Flags().String(flagInitiatorChainChannel, "", "The channel info: '<channelID>:<portID>'")
	cmd.Flags().StringSlice(flagSigners, nil, "Signers info")
	cmd.Flags().String(flagCallInfo, "", "A contract call info")
	cmd.Flags().String(flags.FlagOutputDocument, "", "The document will be written to the given file instead of STDOUT")

	cmd.MarkFlagRequired(flagSigners)
	cmd.MarkFlagRequired(flagCallInfo)

	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")
	return cmd
}

func parseChannelInfoFromString(s string) (*xcctypes.ChannelInfo, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return nil, errors.New("channel format must be follow a format: '<channelID>:<portID>'")
	}
	return &xcctypes.ChannelInfo{Channel: parts[0], Port: parts[1]}, nil
}

func resolveXCCForInitiator(queryClient channeltypes.QueryClient, s string) (*codectypes.Any, error) {
	ci, err := parseChannelInfoFromString(s)
	if err != nil {
		return nil, err
	}
	xcc, err := resolveXCCFromChannel(queryClient, *ci)
	if err != nil {
		return nil, err
	}
	return xcctypes.PackCrossChainChannel(xcc)
}

func resolveXCCFromChannel(queryClient channeltypes.QueryClient, ci xcctypes.ChannelInfo) (xcctypes.XCC, error) {
	// Get a source chainId of initiator
	ctx := context.Background()
	res, err := queryClient.Channel(ctx, &channeltypes.QueryChannelRequest{PortId: ci.Port, ChannelId: ci.Channel})
	if err != nil {
		return nil, err
	}
	return &xcctypes.ChannelInfo{Port: res.Channel.Counterparty.PortId, Channel: res.Channel.Counterparty.ChannelId}, nil
}

func setOutputFile(cmd *cobra.Command) (func(), error) {
	outputDoc, _ := cmd.Flags().GetString(flags.FlagOutputDocument)
	if outputDoc == "" {
		cmd.SetOut(cmd.OutOrStdout())
		return func() {}, nil
	}

	fp, err := os.OpenFile(outputDoc, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return func() {}, err
	}

	cmd.SetOut(fp)

	return func() { fp.Close() }, nil
}
