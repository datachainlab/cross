package cli

import (
	"context"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/modules/core/04-channel/types"
	ibctmtypes "github.com/cosmos/ibc-go/modules/light-clients/07-tendermint/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/datachainlab/cross/x/core/auth/types"
	authtypes "github.com/datachainlab/cross/x/core/auth/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
)

func NewIBCSignTxCmd() *cobra.Command {
	const (
		flagTxID                  = "tx-id"
		flagInitiatorChainChannel = "initiator-chain-channel"
	)

	cmd := &cobra.Command{
		Use:   "ibc-signtx",
		Short: "Sign the cross-chain transaction on other chain via the chain",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			clientCtx = clientCtx.WithOutputFormat("json")
			anyXCC, err := resolveXCC(
				channeltypes.NewQueryClient(clientCtx),
				viper.GetString(flagInitiatorChainChannel),
			)
			if err != nil {
				return err
			}
			signer := authtypes.AccountIDFromAccAddress(clientCtx.FromAddress)
			txID, err := hex.DecodeString(viper.GetString(flagTxID))
			if err != nil {
				return err
			}
			h, height, err := QueryTendermintHeader(clientCtx)
			if err != nil {
				return err
			}
			version := clienttypes.ParseChainID(h.Header.ChainID)
			msg := types.NewMsgIBCSignTx(
				anyXCC,
				txID,
				[]authtypes.AccountID{signer},
				clienttypes.NewHeight(version, uint64(height)+100),
				0,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().String(flagTxID, "", "hex encoding of the TxID")
	cmd.Flags().String(flagInitiatorChainChannel, "", "channel info: '<channelID>:<portID>'")
	cmd.MarkFlagRequired(flagTxID)
	cmd.MarkFlagRequired(flagInitiatorChainChannel)

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func resolveXCC(queryClient channeltypes.QueryClient, s string) (*codectypes.Any, error) {
	ci, err := parseChannelInfoFromString(s)
	if err != nil {
		return nil, err
	}
	return xcctypes.PackCrossChainChannel(ci)
}

func parseChannelInfoFromString(s string) (*xcctypes.ChannelInfo, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return nil, errors.New("channel format must be follow a format: '<channelID>:<portID>'")
	}
	return &xcctypes.ChannelInfo{Channel: parts[0], Port: parts[1]}, nil
}

// QueryTendermintHeader takes a client context and returns the appropriate
// tendermint header
// Original implementation(but has a little) is here: https://github.com/cosmos/cosmos-sdk/blob/300b7393addba8c162cae929db90b083dcf93bd0/x/ibc/core/02-client/client/utils/utils.go#L123
func QueryTendermintHeader(clientCtx client.Context) (ibctmtypes.Header, int64, error) {
	node, err := clientCtx.GetNode()
	if err != nil {
		return ibctmtypes.Header{}, 0, err
	}

	info, err := node.ABCIInfo(context.Background())
	if err != nil {
		return ibctmtypes.Header{}, 0, err
	}

	height := info.Response.LastBlockHeight

	commit, err := node.Commit(context.Background(), &height)
	if err != nil {
		return ibctmtypes.Header{}, 0, err
	}

	page := 1
	count := 10_000

	validators, err := node.Validators(context.Background(), &height, &page, &count)
	if err != nil {
		return ibctmtypes.Header{}, 0, err
	}

	protoCommit := commit.SignedHeader.ToProto()
	protoValset, err := tmtypes.NewValidatorSet(validators.Validators).ToProto()
	if err != nil {
		return ibctmtypes.Header{}, 0, err
	}

	header := ibctmtypes.Header{
		SignedHeader: protoCommit,
		ValidatorSet: protoValset,
	}

	return header, height, nil
}
