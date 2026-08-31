package settlementapp

import (
	"encoding/json"
	"testing"
)

func TestDefaultGenesisStateValid(t *testing.T) {
	state := DefaultGenesisState()
	if err := state.Validate(); err != nil {
		t.Fatalf("default genesis must validate: %v", err)
	}
}

func TestParseGenesisRoundTrip(t *testing.T) {
	state := DefaultGenesisState()
	raw, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	appState, err := json.Marshal(map[string]json.RawMessage{GenesisKey: raw})
	if err != nil {
		t.Fatal(err)
	}

	got, found, err := parseGenesis(appState)
	if err != nil {
		t.Fatalf("parse genesis: %v", err)
	}
	if !found {
		t.Fatal("expected ElasticChain genesis to be found")
	}
	if err := got.Validate(); err != nil {
		t.Fatalf("round-tripped genesis invalid: %v", err)
	}
	if got.Topology.NextDomainID != state.Topology.NextDomainID || len(got.Topology.Domains) != len(state.Topology.Domains) {
		t.Fatalf("topology changed during round trip: got %+v want %+v", got.Topology, state.Topology)
	}
}

func TestParseGenesisAllowsLegacyGenesisWithoutElasticState(t *testing.T) {
	state, found, err := parseGenesis([]byte(`{"bank":{}}`))
	if err != nil {
		t.Fatalf("parse legacy genesis: %v", err)
	}
	if found {
		t.Fatalf("unexpected ElasticChain genesis: %+v", state)
	}
}

func TestGenesisRejectsUnknownVersion(t *testing.T) {
	state := DefaultGenesisState()
	state.Version++
	if err := state.Validate(); err == nil {
		t.Fatal("expected unknown genesis version to be rejected")
	}
}
