package main

import (
	"fmt"
	"os"
	"path/filepath"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log/v2"
	"cosmossdk.io/simapp"
	simdcmd "cosmossdk.io/simapp/simd/cmd"

	"github.com/cosmos/cosmos-sdk/server"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"

	"github.com/KiYouJyo/ElasticChain/internal/settlementapp"
)

const version = "0.1.0-dev"

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "resolve home directory:", err)
		os.Exit(1)
	}

	simapp.DefaultNodeHome = filepath.Join(home, ".elasticchain")

	rootCmd := simdcmd.NewRootCmd()
	rootCmd.Use = "elasticchaind"
	rootCmd.Short = "ElasticChain experimental PoS settlement daemon"
	rootCmd.Long = "ElasticChain experimental PoS/BFT settlement daemon. Research software; do not secure real value with this build."
	rootCmd.Version = version
	rootCmd.AddCommand(newDemoScalingCmd())

	// The upstream CLI is retained for mature account/genesis/staking commands,
	// but node execution must use ElasticChain's application wrapper so the
	// elastic and xmsg stores are mounted before the database is loaded.
	for _, child := range rootCmd.Commands() {
		if child.Name() == "start" {
			rootCmd.RemoveCommand(child)
			break
		}
	}
	rootCmd.AddCommand(server.StartCmdWithOptions(
		func(logger log.Logger, db dbm.DB, appOpts servertypes.AppOptions) servertypes.Application {
			return settlementapp.New(
				logger,
				db,
				true,
				appOpts,
				server.DefaultBaseappOptions(appOpts)...,
			)
		},
		simapp.DefaultNodeHome,
		server.StartCmdOptions{},
	))

	if err := svrcmd.Execute(rootCmd, "ELASTICCHAIN", simapp.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
