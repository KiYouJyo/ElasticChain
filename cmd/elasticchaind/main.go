package main

import (
	"fmt"
	"os"
	"path/filepath"

	"cosmossdk.io/simapp"
	simdcmd "cosmossdk.io/simapp/simd/cmd"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
)

const version = "0.1.0-dev"

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "resolve home directory:", err)
		os.Exit(1)
	}

	// v0.1 intentionally reuses the Cosmos SDK reference settlement application
	// while ElasticChain's deterministic topology and messaging logic remains in
	// internal/elastic. The next step is replacing the reference application
	// wiring with ElasticChain-owned x/elastic and x/xmsg modules.
	simapp.DefaultNodeHome = filepath.Join(home, ".elasticchain")

	rootCmd := simdcmd.NewRootCmd()
	rootCmd.Use = "elasticchaind"
	rootCmd.Short = "ElasticChain experimental PoS settlement daemon"
	rootCmd.Long = "ElasticChain experimental PoS/BFT settlement daemon. Research software; do not secure real value with this build."
	rootCmd.Version = version
	rootCmd.AddCommand(newDemoScalingCmd())

	if err := svrcmd.Execute(rootCmd, "ELASTICCHAIN", simapp.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
