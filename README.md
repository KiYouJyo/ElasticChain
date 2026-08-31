# ElasticChain

ElasticChain is an experimental modular Proof-of-Stake blockchain architecture focused on **elastic horizontal scaling with shared security**.

The project explores one central question:

> Can blockchain execution capacity expand and contract with demand without requiring every validator to execute every transaction?

## Design direction

ElasticChain separates the network into four responsibilities:

1. **Settlement & consensus** — a small, deterministic PoS/BFT layer owns validator state, finality, protocol parameters and execution-domain commitments.
2. **Elastic execution domains** — transactions execute in parallel domains that can split or merge at deterministic epoch boundaries.
3. **Cross-domain messaging** — domains communicate asynchronously through finalized settlement-layer messages rather than synchronous global state access.
4. **Proof & data availability** — the prototype begins with explicit commitments and reproducible execution, then evolves toward validity proofs, proof aggregation, erasure coding and data-availability sampling.

The initial implementation deliberately prioritizes correctness and inspectability over headline TPS. Scaling decisions must be reproducible from finalized on-chain state; local mempool observations or machine-specific timing are never consensus inputs.

## Status

Repository bootstrap in progress. The first milestone is a deterministic local prototype and protocol specification, followed by a Cosmos SDK / CometBFT settlement chain.

## Planned milestones

- **v0.1** — settlement-chain foundation, native token, staking, validators, transfers and slashing.
- **v0.2** — EVM compatibility and standard Ethereum JSON-RPC tooling.
- **v0.3** — conflict-aware parallel execution within one execution domain.
- **v0.4** — two execution domains with finalized asynchronous cross-domain messages.
- **v0.5** — deterministic split/merge and elastic execution-domain scaling.
- **v0.6** — validity proofs, recursive aggregation and dedicated data-availability scaling.

See `docs/ARCHITECTURE.md`, `docs/PROTOCOL.md`, and `docs/ROADMAP.md` for the normative project direction once the bootstrap PR lands.

## Project maturity

ElasticChain is research software. It is **not production-ready**, has not undergone a security audit, and must not be used to secure real financial value at this stage.
