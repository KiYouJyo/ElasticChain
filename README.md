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

## Current status — v0.1 settlement foundation in progress

ElasticChain now has two working layers.

### Deterministic elastic reference core

- binary-prefix execution-domain topology and routing;
- deterministic split/merge planning with hysteresis;
- integer basis-point utilization thresholds;
- finalized cross-domain message lifecycle;
- destination binding and exactly-once consumption;
- per-source-domain nonce uniqueness;
- race-tested unit coverage and scaling demo.

### Real PoS/BFT settlement bootstrap

`elasticchaind` now wraps the Cosmos SDK v0.54.4 settlement stack and can launch a four-validator local CometBFT network using the test denomination `uelastic`.

CI verifies that all four validator RPCs reach the same `elastic-local-1` chain at height >= 2 and that the active validator set contains four validators.

The current daemon deliberately reuses Cosmos SDK `simapp` for standard accounts, staking, distribution, slashing and network plumbing. This is a bootstrap boundary, not the final application architecture. The next v0.1 work is to replace that reference wiring with ElasticChain-owned `x/elastic` and `x/xmsg` consensus state while retaining Cosmos SDK / CometBFT as the maintained consensus substrate.

## Quick start

Requires the Go version declared in `go.mod`.

```bash
make check
make demo
make localnet-smoke
```

The deterministic scaling demo uses short pressure windows and shows:

```text
1 active domain → split → 2 active domains → merge → 1 active domain
```

The localnet smoke test builds the daemon, generates four validator homes, starts all four validators, waits for shared finality and checks the validator set.

See `docs/LOCALNET.md` for manual local-network commands.

## Repository map

```text
cmd/elasticchaind/          settlement daemon + research commands
internal/elastic/           deterministic protocol reference algorithms
scripts/                    four-validator localnet tooling
config/                     protocol-parameter examples
docs/ARCHITECTURE.md        system decomposition and scaling model
docs/PROTOCOL.md            normative protocol invariants
docs/ROADMAP.md             milestone plan and acceptance gates
docs/SECURITY.md            threat model and security gates
docs/DEPENDENCIES.md        framework/version baseline
docs/LOCALNET.md            local PoS/BFT network instructions
.github/workflows/ci.yml    core checks + four-validator finality smoke test
```

## Planned milestones

- **v0.1** — settlement-chain foundation, native token, staking, validators, transfers, slashing and persistent ElasticChain protocol state.
- **v0.2** — EVM compatibility and standard Ethereum JSON-RPC tooling.
- **v0.3** — conflict-aware parallel execution within one execution domain.
- **v0.4** — two execution domains with finalized asynchronous cross-domain messages.
- **v0.5** — deterministic split/merge and elastic execution-domain scaling.
- **v0.6** — validity proofs, recursive aggregation and dedicated data-availability scaling.
- **v0.7+** — rotating shared-security execution committees, adversarial testnet, audits and economics.

See `docs/ROADMAP.md` for completion criteria. A milestone is not complete merely because code exists; its acceptance tests must pass.

## Framework direction

The v0.1 settlement layer is pinned to Cosmos SDK v0.54.4 and the compatible CometBFT v0.39.4 line selected by that release. EVM support is deferred to v0.2 through Cosmos EVM instead of inventing a new VM.

## Project maturity

ElasticChain is research software. It is **not production-ready**, has not undergone a security audit, and must not be used to secure real financial value at this stage.
