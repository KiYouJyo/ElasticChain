# ElasticChain

ElasticChain is an experimental modular Proof-of-Stake blockchain focused on **digital assets, programmable finance, and elastic horizontal scaling with shared security**.

The product goal is to support useful financial and digital-asset workloads — native assets, fungible tokens, NFTs, marketplaces and programmable contracts — while the protocol research asks:

> Can blockchain execution capacity expand and contract with demand without requiring every validator to execute every transaction?

ElasticChain is intentionally **not** specialized for planning/architecture workloads; domain-specific chains can be built separately. This repository stays focused on general digital-asset and financial execution.

## Architecture

ElasticChain separates the network into four responsibilities:

1. **Settlement & consensus** — a small, deterministic PoS/BFT layer owns validator state, finality, protocol parameters and execution-domain commitments.
2. **Elastic execution domains** — transactions execute in parallel domains that can split or merge at deterministic epoch boundaries.
3. **Cross-domain messaging** — domains communicate asynchronously through finalized settlement-layer messages rather than synchronous global state access.
4. **Proof & data availability** — the prototype begins with explicit commitments and reproducible execution, then evolves toward validity proofs, proof aggregation, erasure coding and data-availability sampling.

The implementation prioritizes correctness and inspectability over headline TPS. Scaling decisions must be reproducible from finalized on-chain state; local mempool observations or machine-specific timing are never consensus inputs.

## Application priorities

1. **Fungible assets** — native asset plus standard token issuance and transfer.
2. **NFTs / multi-token assets** — ERC-721, ERC-1155, ownership, metadata and batch operations.
3. **Financial/market workloads** — escrow, auctions, swaps, AMM-style pools and marketplace settlement.
4. **Developer compatibility** — Ethereum JSON-RPC, MetaMask, Foundry/Hardhat and established contract standards.
5. **Elastic scaling** — use those real workloads to prove safe parallel and cross-domain execution.

Issuing a token is therefore a first-class use case, but not the only benchmark. NFT ownership, marketplace contention and financial contract state must scale without supply drift, duplication or replay.

## Current status — v0.1 settlement foundation in progress

ElasticChain now has two working layers.

### Deterministic elastic reference core

- binary-prefix execution-domain topology and routing;
- deterministic split/merge planning with hysteresis;
- integer basis-point utilization thresholds;
- finalized cross-domain message lifecycle;
- destination binding and exactly-once consumption;
- per-source-domain nonce uniqueness;
- versioned deterministic snapshot/restore with fail-closed validation;
- race-tested unit coverage and scaling demo.

### Real PoS/BFT settlement bootstrap

`elasticchaind` is based on Cosmos SDK v0.54.4 / CometBFT and can launch a four-validator local network using the test denomination `uelastic`.

The v0.1 application wrapper stores ElasticChain topology/message snapshots in an isolated binary namespace inside a Cosmos IAVL store that exists from genesis. This state participates in the committed AppHash and is validated again during normal block production. The namespace is deliberately transitional: it proves restart and genesis export/import semantics without late-mounting new Store-v2 trees. A later owned application wiring will migrate the same logical state into dedicated `x/elastic` and `x/xmsg` stores.

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

The localnet smoke test builds the daemon, generates four validator homes, starts all validators, checks shared finality/AppHash, restarts the same node homes, exports the chain state to a portable genesis, wipes application/block databases, reboots from that export, and verifies continued shared finality.

See `docs/LOCALNET.md` for manual local-network commands.

## Repository map

```text
cmd/elasticchaind/          settlement daemon + research commands
internal/elastic/           deterministic protocol reference algorithms
internal/settlementapp/     ElasticChain-owned settlement application wiring
scripts/                    four-validator localnet tooling
config/                     protocol-parameter examples
docs/ARCHITECTURE.md        system decomposition and scaling model
docs/PROTOCOL.md            normative protocol invariants
docs/ASSETS.md              FT/NFT/market asset semantics and invariants
docs/ROADMAP.md             milestone plan and acceptance gates
docs/SECURITY.md            threat model and security gates
docs/DEPENDENCIES.md        framework/version baseline
docs/LOCALNET.md            local PoS/BFT network instructions
.github/workflows/ci.yml    core checks + four-validator persistence/export smoke test
```

## Planned milestones

- **v0.1** — settlement foundation, native asset, staking/validators and persistent ElasticChain consensus state.
- **v0.2** — digital asset runtime: Cosmos EVM + ERC-20 + ERC-721 + ERC-1155 + standard wallet/tooling compatibility.
- **v0.3** — financial and marketplace primitives: escrow, NFT sales/auctions, token swaps and AMM reference workloads.
- **v0.4** — deterministic parallel execution using token/NFT/market workloads.
- **v0.5** — two execution domains with finalized asynchronous cross-domain asset/state movement.
- **v0.6** — deterministic split/merge and elastic execution-domain scaling.
- **v0.7** — validity proofs and dedicated data-availability scaling.
- **v0.8+** — rotating shared-security committees, adversarial testnet, audits and economics.

See `docs/ROADMAP.md` for completion criteria. A milestone is not complete merely because code exists; its acceptance tests must pass.

## Framework direction

The v0.1 settlement layer is pinned to Cosmos SDK v0.54.4 and its compatible CometBFT line. v0.2 uses Cosmos EVM rather than inventing a new VM so established asset standards and tooling arrive before protocol-specific application abstractions.

## Project maturity

ElasticChain is research software. It is **not production-ready**, has not undergone a security audit, and must not be used to secure real financial value at this stage.
