package cli

import (
	"bufio"
	"encoding/hex"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/bluele/crossccc/x/ibc/crossccc/internal/types"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagSigners  = "signers"
	flagOPs      = "ops"
	flagContract = "contract"
)

/*
GetInitiateTxCmd returns a command that executes to initiate a distributed transaction
This command implemetation follows under some assumptions.
Assumption:
	- All keys that are used to create a signature exists on this keychain
*/
func GetInitiateTxCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "initiate [transitions-file] [nonce]",
		Short: "Initiate a distributed transaction",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(nil).WithCodec(cdc).WithBroadcastMode(flags.BroadcastBlock)
			sender := cliCtx.GetFromAddress()
			sts, err := readStateTransitionsFile(cdc, args[0])
			if err != nil {
				return err
			}
			nonce, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}
			msg := types.NewMsgInitiate(
				sender,
				sts,
				nonce,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

func GetCreateStateTransitionFileCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "create-st [dest-file] [src-port] [src-channel] [contract] --signers [signers] --ops [ops]",
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			dest := args[0]
			srcc := types.ChannelInfo{
				Port:    args[1],
				Channel: args[2],
			}
			signers, err := readSignersFromFlag()
			if err != nil {
				return err
			}
			contract, err := hex.DecodeString(viper.GetString(flagContract))
			if err != nil {
				return err
			}
			ops, err := readOPsFromFlag(cdc)
			if err != nil {
				return err
			}
			st := types.NewStateTransition(srcc, signers, contract, ops)
			bz, err := cdc.MarshalBinaryLengthPrefixed(st)
			if err != nil {
				return err
			}
			if err := ioutil.WriteFile(dest, bz, 0600); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringSlice(flagSigners, nil, "")
	cmd.Flags().StringSlice(flagOPs, nil, "")
	cmd.Flags().String(flagContract, "", "")
	return cmd
}

func GetMergeStateTransitionFilesCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use: "merge-st [dest-file] [[src-file]...]",
		RunE: func(cmd *cobra.Command, args []string) error {
			dest := args[0]
			sts, err := concatBytesFromFiles(cdc, args[1:]...)
			if err != nil {
				return err
			}
			bz, err := cdc.MarshalBinaryLengthPrefixed(sts)
			if err != nil {
				return err
			}
			if err := ioutil.WriteFile(dest, bz, 0600); err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}

func concatBytesFromFiles(cdc *codec.Codec, srcs ...string) (types.StateTransitions, error) {
	var sts types.StateTransitions
	for _, src := range srcs {
		st, err := readStateTransitionFromFile(cdc, src)
		if err != nil {
			return nil, err
		}
		sts = append(sts, *st)
	}
	return sts, nil
}

func readStateTransitionFromFile(cdc *codec.Codec, path string) (*types.StateTransition, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var st types.StateTransition
	err = cdc.UnmarshalBinaryLengthPrefixed(b, &st)
	return &st, err
}

func readStateTransitionsFile(cdc *codec.Codec, path string) (types.StateTransitions, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var sts types.StateTransitions
	err = cdc.UnmarshalBinaryLengthPrefixed(b, &sts)
	return sts, err
}

func readSignersFromFlag() ([]sdk.AccAddress, error) {
	signerStrs := viper.GetStringSlice(flagSigners)
	var addrs []sdk.AccAddress
	for _, s := range signerStrs {
		addr, err := sdk.AccAddressFromHex(s)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, addr)
	}
	return addrs, nil
}

func readOPsFromFlag(cdc *codec.Codec) (types.OPs, error) {
	opStrs := viper.GetStringSlice(flagOPs)
	var ops types.OPs
	for _, s := range opStrs {
		b, err := hex.DecodeString(s)
		if err != nil {
			return nil, err
		}
		var op types.OP
		if err := cdc.UnmarshalBinaryLengthPrefixed(b, &op); err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}
	return ops, nil
}
