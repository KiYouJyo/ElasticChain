# ElasticChain Roadmap

ElasticChain is an experimental elastic financial and digital-asset blockchain. Its first application priority is useful asset issuance and exchange — fungible tokens, NFTs and programmable financial workloads — while its protocol priority remains deterministic safety and horizontal scaling.

The roadmap is ordered so that the chain first becomes useful for real asset workloads, then proves that those workloads can scale safely. A milestone is complete only when its acceptance criteria pass; version numbers describe protocol capability, not marketing releases.

## Product priorities

1. **Digital assets:** native asset, fungible tokens, NFTs and multi-token assets.
2. **Financial execution:** swaps, pools, escrow, auctions and other auditable financial contracts.
3. **Interoperable developer experience:** Ethereum-compatible tooling and widely used token standards before inventing new application standards.
4. **Elastic execution:** capacity should grow horizontally with partitionable demand without weakening settlement safety.
5. **Proof/data availability:** reduce settlement verification and data replication cost only after execution semantics are stable.

Token issuance is important, but the protocol must not optimize only for creating arbitrary coins. NFT ownership, marketplace workloads, high-frequency transfers and contract state are first-class scaling workloads.

## v0.0.1 — Architecture bootstrap

**Goal:** turn the elastic-chain design into testable protocol primitives without pulling in a blockchain framework.

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

**Goal:** run a real local PoS/BFT settlement network and commit ElasticChain-owned consensus state.

Baseline:

- Cosmos SDK v0.54.x;
- CometBFT version pinned by the selected Cosmos SDK release;
- Go toolchain required by the selected SDK release.

Modules/capabilities:

- auth/accounts;
- bank/native transfers;
- staking/delegation;
- distribution/rewards;
- slashing/evidence;
- governance/upgrade only where required for testnet administration;
- ElasticChain-owned `elastic` state for topology, scaling policy and transition state;
- ElasticChain-owned `xmsg` state for canonical cross-domain messages.

Local network target:

- 4 validators;
- deterministic genesis;
- validator restart/recovery test;
- shared AppHash validation;
- exported state and reproducible local bootstrap scripts.

Acceptance:

- 4-node local network reaches finality continuously;
- native token can transfer and stake;
- validator rewards and penalties are observable;
- ElasticChain topology/message state is part of committed state and survives restart;
- malformed imported topology/message state fails closed;
- CI executes unit and integration tests.

## v0.2 — Digital asset runtime: FT + NFT

**Goal:** make ElasticChain immediately useful for common digital-asset workloads through existing Ethereum standards.

Runtime:

- Cosmos EVM integration;
- Ethereum JSON-RPC;
- MetaMask-compatible accounts and transactions;
- Foundry/Hardhat deployment flow;
- deterministic EVM export/import;
- explicit native denomination ↔ EVM unit conversion.

Asset standards — highest application priority:

- ERC-20 fungible token deployment, mint, burn, approve and transfer tests;
- ERC-721 NFT mint, transfer, approval and metadata tests;
- ERC-1155 multi-token/NFT tests;
- contract ownership and mint-authority patterns;
- immutable and mutable metadata examples with content-hash commitments;
- batch mint/transfer tests;
- royalty-interface compatibility where applicable, without protocol-enforced guaranteed royalties;
- reference contracts kept deliberately small and auditable.

Acceptance:

- deploy/call Solidity contracts locally;
- issue and transfer an ERC-20;
- mint and transfer ERC-721 and ERC-1155 assets;
- wallets can discover balances/ownership through standard interfaces;
- replay protection and chain-ID tests pass;
- EVM/NFT state survives validator restart and deterministic export/import.

## v0.3 — Financial and marketplace primitives

**Goal:** support meaningful financial activity rather than asset issuance alone.

Reference workloads:

- escrow contract;
- fixed-price NFT sale;
- English/Dutch auction experiments;
- constant-product AMM reference pool;
- fungible-token swap;
- liquidity add/remove;
- signed order / permit-style authorization where supported;
- fee accounting and event indexing;
- safe pause/admin patterns for application contracts, clearly separate from protocol authority.

Non-goals for this stage:

- no algorithmic stablecoin;
- no leveraged lending protocol;
- no bridge carrying real external assets;
- no promise that reference contracts are production-audited financial products.

Acceptance:

- end-to-end local flows for token swap, NFT sale and auction pass;
- invariant/property tests cover AMM conservation and authorization rules;
- contracts cannot mint or transfer assets without the permissions specified by their standard;
- workload generator can produce repeatable token/NFT/market traffic for later scaling benchmarks.

## v0.4 — Deterministic parallel execution in one domain

**Goal:** prove deterministic parallelism using the financial workloads from v0.2–v0.3 before cross-domain state exists.

Work:

- transaction access-set declaration/derivation experiment;
- conflict graph construction;
- deterministic batch scheduler;
- parallel execution of disjoint state accesses;
- serial fallback for conflicts;
- benchmark harness across 1/2/4/8/16 worker configurations;
- workload profiles for token transfers, independent NFT mints, marketplace hot spots and AMM contention.

Critical rule:

The committed result must be identical regardless of available CPU worker count.

Acceptance:

- randomized differential test: serial result == parallel result;
- fuzz conflict scheduler;
- reproducible throughput/latency benchmark report;
- no consensus-visible goroutine scheduling dependence.

## v0.5 — Two execution domains

**Goal:** prove the first genuine horizontal partition under real digital-asset workloads.

Work:

- activate prefix-routed Domain A and Domain B;
- domain state commitments on settlement layer;
- asynchronous source → settlement → destination messaging;
- exactly-once destination consumption;
- state migration prototype for one root split;
- per-domain local fee accounting experiment;
- NFT/token ownership and supply preservation during migration;
- failure tests with one execution domain unavailable.

Acceptance:

- same-domain transactions continue while the unrelated domain is intentionally stalled, where settlement rules permit;
- cross-domain message cannot be consumed early, twice or by wrong destination;
- split migration preserves native supply, ERC-20 supply, NFT ownership and application state;
- independent replay reconstructs both domain commitments.

## v0.6 — Elastic execution

**Goal:** make topology scale based on finalized financial/application demand.

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

- uniform token transfers;
- NFT mint bursts;
- one hot collection/marketplace;
- AMM hot-pool contention;
- cross-domain-heavy traffic;
- alternating hot/cold adversarial load;
- validator/executor churn.

Acceptance:

- adding execution resources increases measured aggregate capacity under partitionable workloads;
- unrelated domains do not inherit a hot market's fee spike by default;
- no split/merge causes double-spend, NFT duplication, message replay or supply drift;
- topology state is deterministic across all validators;
- scale-in does not oscillate under threshold noise.

## v0.7 — Proof and data-availability layer

**Goal:** remove settlement-layer dependence on re-executing all domain work and prepare for large asset/contract data throughput.

### Validity proofs

- select proof system using proof size, prover cost, verifier cost, recursion support, audit maturity and licensing;
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

## v0.8 — Shared execution committees

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

## v0.9 — Public adversarial testnet

**Goal:** expose token, NFT, market and scaling workloads to real network failure modes with no meaningful financial value.

Requirements before launch:

- threat model reviewed;
- reproducible binaries;
- deterministic genesis and upgrade procedure;
- telemetry that does not affect consensus;
- faucet-only test assets and NFT collections;
- documented incident response;
- chaos tests and state-sync tests;
- no unaudited bridge carrying real assets.

## v0.10 — Audit and economics

**Goal:** only after protocol behavior stabilizes, design economics around measured security needs.

Work:

- formal supply/issuance proposal for the native asset;
- validator reward model;
- fee burn/distribution decision;
- attack-cost model;
- external security review;
- protocol invariants/property tests;
- governance-minimization proposal;
- economic-spam analysis for cheap token/NFT issuance.

## v1.0 — Conditional production candidate

There is deliberately no target date.

A production candidate is justified only if:

- token, NFT and financial-contract workloads are demonstrably useful and reproducible;
- the elastic architecture demonstrates real horizontal scaling;
- security assumptions are explicit and tested;
- independent reviewers can reproduce the network;
- critical cryptography and consensus changes are audited;
- operating the chain with real value has a defensible reason beyond merely creating another token.
