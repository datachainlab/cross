package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	abci "github.com/tendermint/tendermint/abci/types"
)

// This method should be moved into query.go
func GetRelayPacket(cdc *codec.Codec) *cobra.Command {
	const (
		flagRelayerAddress = "relayer-address"
	)

	cmd := &cobra.Command{
		Use:   "relay [src-height] [src-port] [src-channel] [src-seq] [dst-port] [dst-channel] [txIdx]",
		Short: "relay each packets",
		Args:  cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc).WithBroadcastMode(flags.BroadcastBlock)

			height, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			srcPort, srcChannel := args[1], args[2]
			srcSeq, err := strconv.Atoi(args[3])
			if err != nil {
				return err
			}
			dstPort, dstChannel := args[4], args[5]
			txIndex, err := strconv.Atoi(args[6])
			if err != nil {
				return err
			}

			events := []string{
				fmt.Sprintf("message.%v = '%v'", "action", "cross_initiate"),
				fmt.Sprintf("%v.%v = '%v'", channeltypes.EventTypeSendPacket, channeltypes.AttributeKeySrcPort, srcPort),
				fmt.Sprintf("%v.%v = '%v'", channeltypes.EventTypeSendPacket, channeltypes.AttributeKeySrcChannel, srcChannel),
				fmt.Sprintf("%v.%v = '%v'", channeltypes.EventTypeSendPacket, channeltypes.AttributeKeySequence, srcSeq),
				fmt.Sprintf("%v.%v = '%v'", channeltypes.EventTypeSendPacket, channeltypes.AttributeKeyDstPort, dstPort),
				fmt.Sprintf("%v.%v = '%v'", channeltypes.EventTypeSendPacket, channeltypes.AttributeKeyDstChannel, dstChannel),
			}
			resTx, err := cliCtx.Client.TxSearch(strings.Join(events, " AND "), true, 1, 2, "")
			if err != nil {
				return err
			}
			if resTx.TotalCount == 0 {
				return fmt.Errorf("no events")
			} else if resTx.TotalCount > 1 {
				return fmt.Errorf("multiple events found")
			}

			var count = 0
			var data []byte
		L:
			for _, ev := range resTx.Txs[0].TxResult.Events {
				if ev.Type == channeltypes.EventTypeSendPacket {
					if count != txIndex {
						count++
						continue
					}
					for _, attr := range ev.GetAttributes() {
						if bytes.Equal(attr.GetKey(), []byte(channeltypes.AttributeKeyData)) {
							data = attr.GetValue()
							break L
						}
					}
				}
			}

			var packetData exported.PacketDataI
			cdc.MustUnmarshalJSON(data, &packetData)

			relayer, err := sdk.AccAddressFromBech32(viper.GetString(flagRelayerAddress))
			if err != nil {
				return err
			}
			packet := channeltypes.NewPacket(packetData, uint64(srcSeq), srcPort, srcChannel, dstPort, dstChannel)
			cliCtx.Height = int64(height - 1)
			res, err := cliCtx.QueryABCI(abci.RequestQuery{
				Path:   "store/ibc/key",
				Data:   ibctypes.KeyPacketCommitment(srcPort, srcChannel, uint64(srcSeq)),
				Height: cliCtx.Height,
				Prove:  true,
			})
			if err != nil {
				return err
			}
			if res.IsErr() {
				return fmt.Errorf("failed to execute QueryABCI")
			}
			if bz := channeltypes.CommitPacket(packet.GetData()); !bytes.Equal(res.Value, bz) {
				return fmt.Errorf("unexpected CommitPacket: %X != %X", res.Value, bz)
			}
			proof := commitment.MerkleProof{Proof: res.Proof}
			msg := channeltypes.NewMsgPacket(packet, proof, uint64(height), relayer)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return authclient.PrintUnsignedStdTx(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}
	cmd = flags.PostCommands(cmd)[0]
	cmd.Flags().String(flagRelayerAddress, "", "")
	cmd.MarkFlagRequired(flagRelayerAddress)
	return cmd
}
