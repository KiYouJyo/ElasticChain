# ElasticChain Roadmap

The roadmap is ordered to prove safety properties before adding cryptographic and scaling complexity. A milestone is complete only when its acceptance criteria pass; version numbers describe protocol capability, not marketing releases.

## v0.0.1 — Architecture bootstrap

**Goal:** turn the design into testable protocol primitives without pulling in a blockchain framework.

Deliverables:

- deterministic routing-prefix topology;
- deterministic split/merge planner;
- integer-only utilization policy;
- finalized cross-domain message lifecycle;
- unit tests for routing, hysteresis and exactly-once delivery;
- architecture, protocol and security specifications;
- CI enforcing formatting, vetting and tests.

Acceptance:

- `go test -race ./...` passes;
- `go vet ./...` passes;
- `gofmt -l .` returns no files;
- `go run ./cmd/elasticchaind demo-scaling` demonstrates 1 → 2 → 1 active domains;
- no consensus decision depends on local-only telemetry.

## v0.1 — Settlement-chain foundation

**Goal:** run a real local PoS/BFT settlement network.

Baseline:

- Cosmos SDK v0.54.x;
- CometBFT version pinned by the selected Cosmos SDK release;
- Cosmos SDK runtime/app wiring where appropriate;
- Go toolchain required by the selected SDK release.

Modules/capabilities:

- auth/accounts;
- bank/native transfers;
- staking/delegation;
- distribution/rewards;
- slashing/evidence;
- governance/upgrade only where required for testnet administration;
- custom `x/elastic` module owning topology, scaling policy and transition state;
- custom `x/xmsg` module owning canonical cross-domain messages.

Local network target:

- 4 validators;
- deterministic genesis;
- validator restart/recovery test;
- double-sign/slashing integration test where practical;
- exported state and reproducible local bootstrap scripts.

Acceptance:

- 4-node local network reaches finality continuously;
- native token can transfer and stake;
- validator rewards and penalties are observable;
- `x/elastic` state survives restart/export/import;
- malformed topology transition is rejected by all validators;
- CI executes unit and integration tests.

## v0.2 — EVM compatibility

**Goal:** make the chain usable by existing Ethereum application tooling without inventing a VM.

Target:

- Cosmos EVM integration;
- Ethereum JSON-RPC;
- MetaMask-compatible accounts and transaction flow;
- Solidity deployment through Foundry/Hardhat;
- ERC-20 smoke test;
- clear denomination conversion rules between native Cosmos accounting and EVM units.

Acceptance:

- deploy and call a Solidity contract locally;
- create/transfer a standard ERC-20 test token;
- replay protection and chain ID tests pass;
- EVM state participates in deterministic export/import.

## v0.3 — Parallel execution in one domain

**Goal:** prove deterministic parallelism before cross-domain state exists.

Work:

- transaction access-set declaration/derivation experiment;
- conflict graph construction;
- deterministic batch scheduler;
- parallel execution of disjoint state accesses;
- serial fallback for conflicts;
- benchmark harness across 1/2/4/8/16 worker configurations.

Critical rule:

The committed result must be identical regardless of available CPU worker count.

Acceptance:

- randomized differential test: serial result == parallel result;
- fuzz conflict scheduler;
- reproducible throughput/latency benchmark report;
- no consensus-visible goroutine scheduling dependence.

## v0.4 — Two execution domains

**Goal:** prove the first genuine horizontal partition.

Work:

- activate prefix-routed Domain A and Domain B;
- domain state commitments on settlement layer;
- asynchronous source → settlement → destination messaging;
- exactly-once destination consumption;
- state migration prototype for one root split;
- per-domain local fee accounting experiment;
- failure tests with one execution domain unavailable.

Acceptance:

- same-domain transactions continue while the unrelated domain is intentionally stalled, where settlement rules permit;
- cross-domain message cannot be consumed early, twice or by wrong destination;
- split migration preserves supply and application state;
- independent replay reconstructs both domain commitments.

## v0.5 — Elastic execution

**Goal:** make topology scale based on finalized demand.

Work:

- persist pressure counters in settlement state;
- epoch-delayed topology transition lifecycle;
- deterministic split action scheduling;
- deterministic sibling merge scheduling;
- routing-state migration;
- queue draining rules;
- anti-oscillation hysteresis;
- per-domain fee market;
- benchmark harness for 1/2/4/8/16 execution domains.

Test loads:

- uniform traffic;
- one hot-key/application domain;
- burst traffic;
- cross-domain-heavy traffic;
- alternating hot/cold adversarial load;
- validator/executor churn.

Acceptance:

- adding execution resources increases measured aggregate capacity under partitionable workloads;
- unrelated domains do not inherit a hot domain's fee spike by default;
- no split/merge causes double-spend, message replay or supply drift;
- topology state is deterministic across all validators;
- scale-in does not oscillate under threshold noise.

## v0.6 — Proof and data-availability layer

**Goal:** remove settlement-layer dependence on re-executing all domain work and prepare for large data throughput.

Work streams:

### Validity proofs

- select proof system using transparent criteria: proof size, prover cost, verifier cost, recursion support, audit maturity and licensing;
- prove domain state transition;
- settlement verifier;
- recursive aggregation experiment;
- adversarial invalid-proof suite.

### Data availability

- data commitments;
- erasure coding;
- sampling protocol;
- reconstruction tests;
- withholding simulation;
- light-node sampling benchmark.

Acceptance:

- invalid domain transition cannot settle with an invalid proof;
- independently sampled/reconstructed data is sufficient to replay required state;
- throughput claims include prover and DA costs, not execution alone.

## v0.7 — Shared execution committees

**Goal:** decentralize execution-domain operation without fragmenting security.

Work:

- validator-derived committee selection;
- unpredictable/rotating assignments;
- committee liveness fallback;
- domain operator slashing evidence;
- concentration and correlated-failure simulation.

Acceptance:

- no domain relies on a fixed privileged operator set;
- committee capture probability is quantitatively documented for configured adversarial stake fractions;
- missed/invalid domain work has a deterministic recovery path.

## v0.8 — Public adversarial testnet

**Goal:** expose the protocol to real network failure modes with no meaningful financial value.

Requirements before launch:

- threat model reviewed;
- reproducible binaries;
- deterministic genesis and upgrade procedure;
- telemetry that does not affect consensus;
- faucet-only test asset;
- documented incident response;
- chaos tests and state-sync tests;
- no unaudited bridge carrying real assets.

## v0.9 — Audit and economics

**Goal:** only after protocol behavior stabilizes, design economics around measured security needs.

Work:

- formal supply/issuance proposal;
- validator reward model;
- fee burn/distribution decision;
- attack-cost model;
- external security review;
- protocol invariants/property tests;
- governance-minimization proposal.

## v1.0 — Conditional production candidate

There is deliberately no target date.

A production candidate is justified only if:

- the elastic architecture demonstrates real horizontal scaling;
- security assumptions are explicit and tested;
- independent reviewers can reproduce the network;
- critical cryptography and consensus changes are audited;
- the project has more than one implementation/operator dependency where practical;
- operating the chain with real value has a defensible reason beyond issuing a token.
