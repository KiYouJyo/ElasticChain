# ElasticChain

ElasticChain is an experimental modular Proof-of-Stake blockchain architecture focused on **elastic horizontal scaling with shared security**.

The project explores one central question:

> Can blockchain execution capacity expand and contract with demand without requiring every validator to execute every transaction?

## Architecture

ElasticChain separates the network into four responsibilities:

1. **Settlement & consensus** — a small, deterministic PoS/BFT layer owns validator state, finality, protocol parameters and execution-domain commitments.
2. **Elastic execution domains** — transactions execute in parallel domains that can split or merge at deterministic epoch boundaries.
3. **Cross-domain messaging** — domains communicate asynchronously through finalized settlement-layer messages rather than synchronous global state access.
4. **Proof & data availability** — the prototype begins with explicit commitments and reproducible execution, then evolves toward validity proofs, proof aggregation, erasure coding and data-availability sampling.

The initial implementation deliberately prioritizes correctness and inspectability over headline TPS. Scaling decisions must be reproducible from finalized on-chain state; local mempool observations or machine-specific timing are never consensus inputs.

## Current status — v0.0.1 bootstrap

The repository now contains a dependency-free Go reference prototype for the consensus-sensitive elastic primitives:

- binary-prefix execution-domain topology;
- deterministic routing;
- split/merge planning with hysteresis;
- integer basis-point utilization thresholds;
- finalized cross-domain message lifecycle;
- exactly-once message consumption;
- unit tests and CI.

This is **not yet a networked blockchain node**. The next milestone, v0.1, wraps these invariants in a real Cosmos SDK / CometBFT settlement chain.

## Quick start

Requires the Go version declared in `go.mod`.

```bash
make check
make demo
```

Or directly:

```bash
go test -race ./...
go run ./cmd/elasticchaind demo-scaling
```

The scaling demo intentionally uses very short pressure windows and shows the reference topology moving:

```text
1 active domain → split → 2 active domains → merge → 1 active domain
```

Production-shaped defaults remain conservative and are recorded in `config/protocol.example.json`.

## Repository map

```text
cmd/elasticchaind/          local research CLI
internal/elastic/           deterministic reference algorithms
config/                     protocol-parameter examples
docs/ARCHITECTURE.md        system decomposition and scaling model
docs/PROTOCOL.md            normative protocol invariants
docs/ROADMAP.md             milestone plan and acceptance gates
docs/SECURITY.md            threat model and security gates
docs/DEPENDENCIES.md        framework/version baseline
.github/workflows/ci.yml    formatting, vet and race-test checks
```

## Planned milestones

- **v0.1** — settlement-chain foundation, native token, staking, validators, transfers and slashing.
- **v0.2** — EVM compatibility and standard Ethereum JSON-RPC tooling.
- **v0.3** — conflict-aware parallel execution within one execution domain.
- **v0.4** — two execution domains with finalized asynchronous cross-domain messages.
- **v0.5** — deterministic split/merge and elastic execution-domain scaling.
- **v0.6** — validity proofs, recursive aggregation and dedicated data-availability scaling.
- **v0.7+** — rotating shared-security execution committees, adversarial testnet, audits and economics.

See `docs/ROADMAP.md` for completion criteria. A milestone is not complete merely because code exists; its acceptance tests must pass.

## Framework direction

The v0.1 settlement layer targets the Cosmos SDK v0.54.x line. The project will follow the CometBFT version pinned by the selected Cosmos SDK release rather than independently upgrading the consensus engine. EVM support is deferred to v0.2 through Cosmos EVM instead of inventing a new VM.

## Project maturity

ElasticChain is research software. It is **not production-ready**, has not undergone a security audit, and must not be used to secure real financial value at this stage.
