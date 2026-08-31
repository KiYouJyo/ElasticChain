package settlementapp

import (
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log/v2"
	"cosmossdk.io/simapp"

	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/KiYouJyo/ElasticChain/internal/elastic"
)

const (
	GenesisKey     = "elasticchain"
	genesisVersion = uint32(1)
)

// v0.1 deliberately stores ElasticChain-owned state under a collision-resistant
// namespace inside an already versioned Cosmos SDK IAVL store. The previous
// prototype mounted new stores after SimApp construction; that could commit the
// first blocks but Store v2 could not reload those late-added stores after a
// process restart. Using a genesis-mounted host store keeps ElasticChain state
// in AppHash while we prove lifecycle/export/import semantics. A dedicated
// x/elastic and x/xmsg store migration belongs to the later owned app wiring.
var (
	stateNamespace = []byte{0xff, 'e', 'l', 'a', 's', 't', 'i', 'c', 'c', 'h', 'a', 'i', 'n', 0x00}
	topologyKey    = appendNamespacedKey("topology/v1")
	messageKey     = appendNamespacedKey("messages/v1")
)

// GenesisState is the portable JSON boundary for ElasticChain-owned consensus
// state. Standard Cosmos modules remain in their normal genesis keys.
type GenesisState struct {
	Version  uint32                       `json:"version"`
	Topology elastic.TopologySnapshot     `json:"topology"`
	Messages elastic.MessageQueueSnapshot `json:"messages"`
}

func DefaultGenesisState() GenesisState {
	return GenesisState{
		Version:  genesisVersion,
		Topology: elastic.NewTopology().Snapshot(),
		Messages: elastic.NewMessageQueue().Snapshot(),
	}
}

func (g GenesisState) Validate() error {
	if g.Version != genesisVersion {
		return fmt.Errorf("unsupported ElasticChain genesis version %d", g.Version)
	}
	if err := g.Topology.Validate(); err != nil {
		return fmt.Errorf("invalid topology: %w", err)
	}
	if err := g.Messages.Validate(); err != nil {
		return fmt.Errorf("invalid message queue: %w", err)
	}
	return nil
}

// App composes the Cosmos SDK reference settlement application with
// ElasticChain-owned consensus state. Standard account, bank, staking and
// slashing responsibilities remain upstream while topology and cross-domain
// messaging are stored under an isolated namespace in committed application
// state and therefore participate in the same AppHash/finality.
type App struct {
	*simapp.SimApp
	stateKey *storetypes.KVStoreKey
}

// New constructs the upstream application without loading the database, swaps
// in ElasticChain lifecycle hooks, and only then loads the latest committed
// version. No late-mounted IAVL stores are introduced in v0.1.
func New(
	logger log.Logger,
	db dbm.DB,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	upstream := simapp.NewSimApp(logger, db, false, appOpts, baseAppOptions...)

	app := &App{
		SimApp:   upstream,
		stateKey: upstream.GetKey(upgradetypes.StoreKey),
	}
	if app.stateKey == nil {
		panic("Cosmos SDK upgrade store is unavailable for ElasticChain v0.1 state namespace")
	}
	upstream.SetInitChainer(app.InitChainer)
	upstream.SetBeginBlocker(app.BeginBlocker)

	if loadLatest {
		if err := upstream.LoadLatestVersion(); err != nil {
			panic(fmt.Errorf("load ElasticChain application state: %w", err))
		}
	}
	return app
}

// InitChainer imports ElasticChain-owned state when genesis contains it,
// otherwise it creates the canonical default root topology and empty queue.
func (app *App) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	genesis, found, err := parseGenesis(req.AppStateBytes)
	if err != nil {
		return nil, err
	}
	if !found {
		genesis = DefaultGenesisState()
	}
	if err := genesis.Validate(); err != nil {
		return nil, fmt.Errorf("validate ElasticChain genesis: %w", err)
	}

	resp, err := app.SimApp.InitChainer(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := app.writeElasticState(ctx, genesis); err != nil {
		return nil, err
	}
	return resp, nil
}

// BeginBlocker fails closed if committed ElasticChain state cannot be restored.
// This continuously validates the persistence boundary, including after a node
// process restart and database reload.
func (app *App) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	if _, err := app.readElasticState(ctx); err != nil {
		return sdk.BeginBlock{}, err
	}
	return app.SimApp.BeginBlocker(ctx)
}

// ExportAppStateAndValidators extends the standard Cosmos module export with
// ElasticChain's own topology/message state so a fresh network can reconstruct
// the same consensus state from the exported genesis document.
func (app *App) ExportAppStateAndValidators(
	forZeroHeight bool,
	jailAllowedAddrs,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	exported, err := app.SimApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	ctx := app.NewContextLegacy(true, cmtproto.Header{Height: app.LastBlockHeight()})
	elasticGenesis, err := app.readElasticState(ctx)
	if err != nil {
		return servertypes.ExportedApp{}, fmt.Errorf("export ElasticChain state: %w", err)
	}

	var appState map[string]json.RawMessage
	if err := json.Unmarshal(exported.AppState, &appState); err != nil {
		return servertypes.ExportedApp{}, fmt.Errorf("decode upstream app state: %w", err)
	}
	bz, err := json.Marshal(elasticGenesis)
	if err != nil {
		return servertypes.ExportedApp{}, fmt.Errorf("encode ElasticChain genesis: %w", err)
	}
	appState[GenesisKey] = bz
	exported.AppState, err = json.MarshalIndent(appState, "", "  ")
	if err != nil {
		return servertypes.ExportedApp{}, fmt.Errorf("encode exported app state: %w", err)
	}
	return exported, nil
}

func (app *App) writeElasticState(ctx sdk.Context, state GenesisState) error {
	if err := state.Validate(); err != nil {
		return err
	}
	store := ctx.KVStore(app.stateKey)
	if err := setJSON(store, topologyKey, state.Topology); err != nil {
		return fmt.Errorf("write elastic topology: %w", err)
	}
	if err := setJSON(store, messageKey, state.Messages); err != nil {
		return fmt.Errorf("write xmsg queue: %w", err)
	}
	return nil
}

func (app *App) readElasticState(ctx sdk.Context) (GenesisState, error) {
	store := ctx.KVStore(app.stateKey)

	var topology elastic.TopologySnapshot
	if err := getJSON(store, topologyKey, &topology); err != nil {
		return GenesisState{}, fmt.Errorf("load elastic topology: %w", err)
	}
	if _, err := elastic.RestoreTopology(topology); err != nil {
		return GenesisState{}, fmt.Errorf("restore elastic topology: %w", err)
	}

	var messages elastic.MessageQueueSnapshot
	if err := getJSON(store, messageKey, &messages); err != nil {
		return GenesisState{}, fmt.Errorf("load xmsg queue: %w", err)
	}
	if _, err := elastic.RestoreMessageQueue(messages); err != nil {
		return GenesisState{}, fmt.Errorf("restore xmsg queue: %w", err)
	}

	state := GenesisState{Version: genesisVersion, Topology: topology, Messages: messages}
	if err := state.Validate(); err != nil {
		return GenesisState{}, err
	}
	return state, nil
}

func parseGenesis(appStateBytes []byte) (GenesisState, bool, error) {
	var appState map[string]json.RawMessage
	if err := json.Unmarshal(appStateBytes, &appState); err != nil {
		return GenesisState{}, false, fmt.Errorf("decode application genesis: %w", err)
	}
	raw, found := appState[GenesisKey]
	if !found {
		return GenesisState{}, false, nil
	}
	var state GenesisState
	if err := json.Unmarshal(raw, &state); err != nil {
		return GenesisState{}, false, fmt.Errorf("decode ElasticChain genesis: %w", err)
	}
	return state, true, nil
}

func appendNamespacedKey(suffix string) []byte {
	key := make([]byte, 0, len(stateNamespace)+len(suffix))
	key = append(key, stateNamespace...)
	key = append(key, suffix...)
	return key
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
		return fmt.Errorf("state key %x is missing", key)
	}
	if err := json.Unmarshal(bz, target); err != nil {
		return err
	}
	return nil
}
