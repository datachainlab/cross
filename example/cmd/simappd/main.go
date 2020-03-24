package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/cosmos/cosmos-sdk/x/staking"

	codecstd "github.com/cosmos/cosmos-sdk/codec/std"
	app "github.com/datachainlab/cross/example/simapp"
	appcontract "github.com/datachainlab/cross/example/simapp/contract"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

const (
	flagInvCheckPeriod = "inv-check-period"
	flagContractMode   = "contractmode"
)

var invCheckPeriod uint

func main() {
	cdc := codecstd.MakeCodec(app.ModuleBasics)
	appCodec := codecstd.NewAppCodec(cdc)

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
	config.Seal()

	ctx := server.NewDefaultContext()
	cobra.EnableCommandSorting = false
	rootCmd := &cobra.Command{
		Use:               "simappd",
		Short:             "Simapp Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	rootCmd.AddCommand(genutilcli.InitCmd(ctx, cdc, app.ModuleBasics, app.DefaultNodeHome))
	rootCmd.AddCommand(genutilcli.CollectGenTxsCmd(ctx, cdc, bank.GenesisBalancesIterator{}, app.DefaultNodeHome))
	rootCmd.AddCommand(genutilcli.MigrateGenesisCmd(ctx, cdc))
	rootCmd.AddCommand(
		genutilcli.GenTxCmd(
			ctx, cdc, app.ModuleBasics, staking.AppModuleBasic{},
			bank.GenesisBalancesIterator{}, app.DefaultNodeHome, app.DefaultCLIHome,
		),
	)
	rootCmd.AddCommand(genutilcli.ValidateGenesisCmd(ctx, cdc, app.ModuleBasics))
	rootCmd.AddCommand(AddGenesisAccountCmd(ctx, cdc, appCodec, app.DefaultNodeHome, app.DefaultCLIHome))
	rootCmd.AddCommand(flags.NewCompletionCmd(rootCmd, true))
	rootCmd.AddCommand(testnetCmd(ctx, cdc, app.ModuleBasics, bank.GenesisBalancesIterator{}))
	rootCmd.AddCommand(replayCmd())
	rootCmd.AddCommand(debug.Cmd(cdc))

	server.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "GA", app.DefaultNodeHome)
	rootCmd.PersistentFlags().UintVar(&invCheckPeriod, flagInvCheckPeriod,
		0, "Assert registered invariants every N blocks")
	rootCmd.PersistentFlags().String(flagContractMode, "", "Use specified contract handler")
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	var cache sdk.MultiStorePersistentCache

	if viper.GetBool(server.FlagInterBlockCache) {
		cache = store.NewCommitKVStoreCacheManager()
	}

	skipUpgradeHeights := make(map[int64]bool)
	for _, h := range viper.GetIntSlice(server.FlagUnsafeSkipUpgrades) {
		skipUpgradeHeights[int64(h)] = true
	}

	return app.NewSimApp(
		logger, db, traceStore, true, skipUpgradeHeights, viper.GetString(cli.HomeFlag), invCheckPeriod,
		getContractHandler(viper.GetString(flagContractMode)),
		app.DefaultAnteHandlerProvider,
		baseapp.SetPruning(store.NewPruningOptionsFromString(viper.GetString("pruning"))),
		baseapp.SetMinGasPrices(viper.GetString(server.FlagMinGasPrices)),
		baseapp.SetHaltHeight(viper.GetUint64(server.FlagHaltHeight)),
		baseapp.SetHaltTime(viper.GetUint64(server.FlagHaltTime)),
		baseapp.SetInterBlockCache(cache),
	)
}

func exportAppStateAndTMValidators(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
) (json.RawMessage, []tmtypes.GenesisValidator, error) {

	if height != -1 {
		gapp := app.NewSimApp(logger, db, traceStore, false, map[int64]bool{}, viper.GetString(cli.HomeFlag), uint(1), getContractHandler(viper.GetString(flagContractMode)), app.DefaultAnteHandlerProvider)
		err := gapp.LoadHeight(height)
		if err != nil {
			return nil, nil, err
		}
		return gapp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
	}

	gapp := app.NewSimApp(logger, db, traceStore, true, map[int64]bool{}, viper.GetString(cli.HomeFlag), uint(1), app.DefaultContractHandlerProvider, app.DefaultAnteHandlerProvider)
	return gapp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
}

func getContractHandler(mode string) app.ContractHandlerProvider {
	switch mode {
	case "":
		return app.DefaultContractHandlerProvider
	case "train":
		return appcontract.TrainReservationContractHandler
	case "hotel":
		return appcontract.HotelReservationContractHandler
	default:
		panic(fmt.Sprintf("unknown contract mode: %v", mode))
	}
}
