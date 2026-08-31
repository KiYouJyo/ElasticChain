package main

import (
	"fmt"
	"os"
	"path/filepath"

	"cosmossdk.io/simapp"
	simdcmd "cosmossdk.io/simapp/simd/cmd"

	"github.com/cosmos/cosmos-sdk/server"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
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

	// Keep the mature upstream account/genesis/staking CLI, but replace commands
	// that instantiate the application. Both running and exporting must use the
	// ElasticChain wrapper or the elastic/xmsg stores would be omitted.
	for _, child := range rootCmd.Commands() {
		if child.Name() == "start" || child.Name() == "export" {
			rootCmd.RemoveCommand(child)
		}
	}
	rootCmd.AddCommand(
		server.StartCmdWithOptions(newSettlementApp, simapp.DefaultNodeHome, server.StartCmdOptions{}),
		server.ExportCmd(exportSettlementApp, simapp.DefaultNodeHome),
	)

	if err := svrcmd.Execute(rootCmd, "ELASTICCHAIN", simapp.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
