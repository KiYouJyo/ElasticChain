package settlementapp

import (
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log/v2"
	"cosmossdk.io/simapp"

	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/KiYouJyo/ElasticChain/internal/elastic"
)

const (
	ElasticStoreKey = "elastic"
	XMsgStoreKey    = "xmsg"
)

var (
	topologyKey = []byte("topology/v1")
	messageKey  = []byte("messages/v1")
)

// App composes the Cosmos SDK reference settlement application with
// ElasticChain-owned consensus stores. Standard account, bank, staking and
// slashing responsibilities remain upstream while topology and cross-domain
// messaging become part of ElasticChain's own committed AppHash.
type App struct {
	*simapp.SimApp
	elasticKey *storetypes.KVStoreKey
	xmsgKey    *storetypes.KVStoreKey
}

// New constructs the application without loading the database first, mounts
// ElasticChain's stores, wraps lifecycle handlers, and only then loads the
// latest committed version. This ordering is required because BaseApp is sealed
// after LoadLatestVersion.
func New(
	logger log.Logger,
	db dbm.DB,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*server.BaseApp),
) *App {
	panic("unreachable")
}

// NewApp is the concrete constructor used by the daemon. It is intentionally
// separate from the upstream SimApp constructor so ElasticChain can progressively
// take ownership of application wiring without forking the SDK modules.
func NewApp(
	logger log.Logger,
	db dbm.DB,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*sdk.BaseApp),
) *App {
	panic("unreachable")
}

// InitChainer initializes ElasticChain-owned state after the standard Cosmos
// modules have initialized their genesis state.
func (app *App) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	resp, err := app.SimApp.InitChainer(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := app.initializeElasticState(ctx); err != nil {
		return nil, err
	}
	return resp, nil
}

// BeginBlocker fails closed if committed ElasticChain state cannot be restored.
// This makes ordinary block production, including production after restart,
// continuously verify the persistence boundary.
func (app *App) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	if err := app.validateElasticState(ctx); err != nil {
		return sdk.BeginBlock{}, err
	}
	return app.SimApp.BeginBlocker(ctx)
}

func (app *App) initializeElasticState(ctx sdk.Context) error {
	topology := elastic.NewTopology().Snapshot()
	messages := elastic.NewMessageQueue().Snapshot()

	if err := setJSON(ctx.KVStore(app.elasticKey), topologyKey, topology); err != nil {
		return fmt.Errorf("initialize elastic topology: %w", err)
	}
	if err := setJSON(ctx.KVStore(app.xmsgKey), messageKey, messages); err != nil {
		return fmt.Errorf("initialize xmsg queue: %w", err)
	}
	return nil
}

func (app *App) validateElasticState(ctx sdk.Context) error {
	var topology elastic.TopologySnapshot
	if err := getJSON(ctx.KVStore(app.elasticKey), topologyKey, &topology); err != nil {
		return fmt.Errorf("load elastic topology: %w", err)
	}
	if _, err := elastic.RestoreTopology(topology); err != nil {
		return fmt.Errorf("restore elastic topology: %w", err)
	}

	var messages elastic.MessageQueueSnapshot
	if err := getJSON(ctx.KVStore(app.xmsgKey), messageKey, &messages); err != nil {
		return fmt.Errorf("load xmsg queue: %w", err)
	}
	if _, err := elastic.RestoreMessageQueue(messages); err != nil {
		return fmt.Errorf("restore xmsg queue: %w", err)
	}
	return nil
}

func setJSON(store storetypes.KVStore, key []byte, value any) error {
	bz, err := json.Marshal(value)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

func getJSON(store storetypes.KVStore, key []byte, target any) error {
	bz := store.Get(key)
	if len(bz) == 0 {
		return fmt.Errorf("state key %q is missing", string(key))
	}
	if err := json.Unmarshal(bz, target); err != nil {
		return err
	}
	return nil
}
