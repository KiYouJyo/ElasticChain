package main

import (
	"fmt"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"

	"github.com/KiYouJyo/ElasticChain/internal/settlementapp"
)

func newSettlementApp(logger log.Logger, db dbm.DB, appOpts servertypes.AppOptions) servertypes.Application {
	return settlementapp.New(
		logger,
		db,
		true,
		appOpts,
		server.DefaultBaseappOptions(appOpts)...,
	)
}

func exportSettlementApp(
	logger log.Logger,
	db dbm.DB,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	app := settlementapp.New(
		logger,
		db,
		false,
		appOpts,
		server.DefaultBaseappOptions(appOpts)...,
	)

	if height == -1 {
		if err := app.LoadLatestVersion(); err != nil {
			return servertypes.ExportedApp{}, fmt.Errorf("load latest ElasticChain state: %w", err)
		}
	} else if err := app.LoadHeight(height); err != nil {
		return servertypes.ExportedApp{}, fmt.Errorf("load ElasticChain height %d: %w", height, err)
	}

	return app.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}
