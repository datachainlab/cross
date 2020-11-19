package cli

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	"github.com/datachainlab/cross/x/core/types"
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
				// This channel info indicates a channel betwwen this chain to initiator chain(source chain?)
				ci, err := parseChannelInfoFromString(initiatorChannel)
				if err != nil {
					return err
				}
				channelClient := channeltypes.NewQueryClient(clientCtx)
				xcc, err := resolveXCCFromChannel(channelClient, *ci)
				if err != nil {
					return err
				}
				anyXCC, err = types.PackCrossChainChannel(xcc)
				if err != nil {
					return err
				}
			}

			var signers []types.AccountID
			for _, s := range viper.GetStringSlice(flagSigners) {
				keyInfo, err := clientCtx.Keyring.Key(s)
				if err != nil {
					return err
				}
				signers = append(signers, types.AccountIDFromAccAddress(keyInfo.GetAddress()))
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

	cmd.Flags().Bool(flagInitiatorChain, false, "")
	cmd.Flags().String(flagInitiatorChainChannel, "", "")
	cmd.Flags().StringSlice(flagSigners, nil, "")
	cmd.Flags().String(flagCallInfo, "", "")
	cmd.Flags().String(flags.FlagOutputDocument, "", "The document will be written to the given file instead of STDOUT")

	cmd.MarkFlagRequired(flagSigners)
	cmd.MarkFlagRequired(flagCallInfo)

	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")
	return cmd
}

func parseChannelInfoFromString(s string) (*types.ChannelInfo, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return nil, errors.New("channel format must be follow a format: '<channelID>:<portID>'")
	}
	return &types.ChannelInfo{Channel: parts[0], Port: parts[1]}, nil
}

func resolveXCCFromChannel(queryClient channeltypes.QueryClient, ci types.ChannelInfo) (types.CrossChainChannel, error) {
	// Get a source chainId of initiator
	ctx := context.Background()
	res, err := queryClient.Channel(ctx, &channeltypes.QueryChannelRequest{PortId: ci.Port, ChannelId: ci.Channel})
	if err != nil {
		return nil, err
	}
	return &types.ChannelInfo{Port: res.Channel.Counterparty.PortId, Channel: res.Channel.Counterparty.ChannelId}, nil
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
