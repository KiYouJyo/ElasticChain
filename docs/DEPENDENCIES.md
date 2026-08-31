# Dependency Baseline

Baseline researched: **2026-08-31**.

This file records the intended framework line for the first settlement-chain implementation. It is deliberately separate from the dependency-free v0.0.x prototype.

## Settlement chain

### Cosmos SDK

Target line: **v0.54.x**.

At bootstrap time the current published release is v0.54.2. Its upstream `go.mod` declares:

- Go `1.25.9`;
- `github.com/cometbft/cometbft v0.39.1`.

Rule: ElasticChain should pin the Cosmos SDK release and accept the CometBFT line selected by that SDK unless an explicit compatibility upgrade is tested. Do **not** independently bump CometBFT merely because a newer CometBFT release exists.

### Cosmos SDK application wiring

Prefer the current runtime/app-wiring model where it remains supported by the selected SDK version:

- `runtime.App` / `AppBuilder`;
- dependency injection/app configuration for standard modules;
- explicit manual wiring only where custom ElasticChain modules require it.

Custom modules planned for v0.1:

- `x/elastic` — execution-domain topology, scaling policy, pressure counters and transition lifecycle;
- `x/xmsg` — canonical cross-domain message registry and lifecycle.

## EVM

Target for v0.2: **Cosmos EVM** rather than a custom EVM fork.

Goals:

- Ethereum bytecode compatibility;
- Ethereum JSON-RPC;
- MetaMask / Foundry / Hardhat compatibility;
- Cosmos-native staking/governance remain settlement-layer concerns.

The exact Cosmos EVM release is intentionally not pinned until v0.2 begins because its compatibility matrix must be checked against the v0.1 settlement dependency set at that time.

## Prototype dependency policy

`internal/elastic` currently uses only the Go standard library. This is intentional:

- elastic routing/scaling/message invariants can be tested without framework state;
- framework upgrades cannot silently change the reference algorithm;
- Cosmos modules can wrap a small deterministic core rather than duplicating logic.

## Upgrade policy

When updating a consensus/framework dependency:

1. read upstream breaking/state-breaking release notes;
2. compare the complete dependency delta;
3. run unit, race, integration and state export/import tests;
4. run a multi-validator upgrade simulation;
5. document any consensus-visible behavior change in an ADR;
6. never merge a consensus dependency bump solely to silence an automated version alert.
