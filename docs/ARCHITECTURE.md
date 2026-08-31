# ElasticChain Architecture

## 1. Objective

ElasticChain is an experimental modular PoS blockchain designed around one primary property:

> **Execution capacity should scale horizontally with demand while security remains anchored in one shared settlement and validator system.**

The architecture intentionally avoids the traditional single-chain assumption that every full validator must execute every user transaction.

The design is organized into four logical layers:

```text
Users / wallets / applications
            │
            ▼
┌──────────────────────────────┐
│ Elastic execution domains   │
│ E0  E1  E2  ...  En         │
│ split / merge by epoch      │
└──────────────┬───────────────┘
               │ commitments + messages
               ▼
┌──────────────────────────────┐
│ Settlement layer            │
│ finality / staking /        │
│ topology / protocol params  │
└──────────────┬───────────────┘
               │
               ▼
┌──────────────────────────────┐
│ Shared PoS/BFT security     │
│ validator set + slashing   │
└──────────────────────────────┘

Future proof/data path:
execution → validity proofs → proof aggregation → settlement
execution data → erasure coding / DAS → availability confirmation
```

## 2. Design principles

### 2.1 Shared security first

Execution domains are not independent sovereign chains with separate security budgets. They ultimately inherit security from one settlement-layer validator set.

A growing number of domains must not fragment stake into increasingly weak independent validator pools.

### 2.2 Determinism over adaptive cleverness

Any automatic protocol transition must be independently reproducible by every honest validator.

Consensus-triggering scaling signals therefore **must be finalized on-chain measurements**. The following are explicitly forbidden as direct consensus inputs:

- local mempool size;
- local CPU or RAM utilization;
- local wall-clock request latency;
- network round-trip measurements seen by only one node;
- external monitoring APIs.

Suitable protocol inputs include finalized block gas utilization, committed DA byte utilization and finalized cross-domain queue depth.

### 2.3 Integer consensus arithmetic

Consensus-sensitive ratios use integer basis points or fixed-point representations. Floating-point arithmetic is excluded from state-transition decisions.

### 2.4 Asynchronous cross-domain composition

The first sharded design does not promise synchronous atomic calls across domains.

Cross-domain execution follows:

```text
source execution
    ↓
source commitment
    ↓
settlement finality
    ↓
canonical cross-domain message
    ↓
destination consumption exactly once
```

This sacrifices some composability in exchange for tractable correctness and horizontal scaling.

### 2.5 Scale-out before scale-up

Higher hardware requirements may improve a single node, but they do not solve the architectural bottleneck that every validator executes all traffic. ElasticChain treats stronger hardware as an optimization, not the primary scaling model.

## 3. Settlement layer

### Responsibilities

The settlement chain owns:

- validator registration and stake;
- validator rewards and slashing;
- finality;
- protocol parameter governance;
- native token accounting;
- execution-domain topology;
- domain state commitments;
- scaling-pressure counters;
- finalized cross-domain messages;
- future proof verification and DA commitments.

### Initial implementation target

The first production-shaped settlement implementation targets **Cosmos SDK v0.54.x** with the CometBFT version pinned by that SDK release rather than independently selecting a consensus-engine version.

The bootstrap prototype remains dependency-free Go so the elastic algorithms can be tested independently from framework wiring.

## 4. Execution domains

An execution domain is a deterministic partition of routing-key space.

The prototype models routing as a binary prefix tree over the high bits of a cryptographic routing hash:

```text
root (*)
   ├── 0*
   │   ├── 00*
   │   └── 01*
   └── 1*
       ├── 10*
       └── 11*
```

A domain split retires one active leaf and activates two child leaves. A merge retires two sibling leaves and reactivates their parent.

This gives deterministic routing without maintaining arbitrary mutable key ranges.

### Routing

For an account, contract or application partition key:

1. derive a cryptographic routing hash;
2. inspect its prefix;
3. select the deepest active domain whose prefix matches.

Routing changes only when a split/merge is finalized at an epoch boundary.

## 5. Elastic scaling

Elastic scaling is intentionally conservative.

### Scale-out pressure

A domain becomes hot when any finalized signal exceeds the configured threshold, for example:

- gas utilization >= 80%;
- DA utilization >= 80%;
- finalized outbound cross-domain queue >= configured threshold.

A split is proposed only after the pressure remains high for a configured number of consecutive epochs.

### Scale-in pressure

A domain becomes cold only when all configured utilization metrics remain low and its cross-domain queue is empty.

Only two active leaf siblings may merge, and only after both remain cold for the configured consolidation period.

### Priority

Scale-out takes priority over scale-in within the same planning epoch. The protocol should not consolidate capacity while another domain is already overloaded.

### Epoch boundaries

Topology changes occur only at deterministic boundaries:

```text
Epoch N finalized metrics
        ↓
Settlement computes action set
        ↓
Action set committed
        ↓
Drain / migration window
        ↓
Topology transition activates at Epoch N+k
```

The current in-memory prototype applies immediately; the Cosmos settlement module must introduce the explicit committed transition window before public testnet use.

## 6. Cross-domain messaging

A cross-domain message contains at minimum:

- source domain;
- destination domain;
- source nonce;
- payload commitment;
- settlement inclusion/finality reference;
- status.

Protocol invariants:

1. message IDs are deterministic;
2. source and destination are bound into the ID;
3. a message cannot be consumed before settlement finality;
4. only the designated destination can consume it;
5. successful consumption is exactly-once;
6. message data required for replay must remain available.

The prototype implements the state machine but not cryptographic inclusion proofs yet.

## 7. Parallel execution

Before introducing multiple execution domains, v0.3 introduces conflict-aware parallel execution inside one domain.

Transactions should declare or derive an access set. Non-conflicting transactions may execute concurrently:

```text
TX1: A → B       ┐
                  ├─ parallel
TX2: C → D       ┘

TX3: A → C  conflicts with TX1 → ordered execution
```

The deterministic scheduler must produce the same committed result regardless of host CPU count.

## 8. Proof roadmap

The project deliberately does **not** require ZK proving for the first multi-domain prototype.

Evolution:

1. deterministic state commitments and settlement-layer validation;
2. fraud-detectable/replayable domain execution;
3. validity-proof prototype;
4. recursive aggregation of multiple domain proofs;
5. proof-market / prover parallelism if justified by benchmarks.

The settlement layer should eventually verify a compact proof of a domain state transition instead of re-executing the domain workload.

## 9. Data availability roadmap

Data availability is treated separately from execution correctness.

The first testnet may publish all required domain data through the settlement path for simplicity. Later versions introduce:

- erasure coding;
- data commitments;
- sampling;
- availability attestations;
- explicit withholding tests.

No throughput claim is considered meaningful if transaction data cannot be independently reconstructed.

## 10. Threat model summary

ElasticChain must defend against:

- Byzantine validators;
- equivocation and double-signing;
- malicious execution-domain operators;
- forged or replayed cross-domain messages;
- data withholding;
- adversarial split/merge oscillation;
- stake concentration;
- committee capture;
- invalid state commitments;
- nondeterministic execution;
- denial-of-service on a single hot domain.

See `SECURITY.md` for security invariants and milestone gates.

## 11. Non-goals for early milestones

The following are intentionally out of scope until the base protocol is demonstrably correct:

- production financial value;
- permissionless bridge TVL;
- anonymous validator participation at internet scale;
- synchronous cross-domain smart-contract calls;
- custom cryptography;
- a new virtual machine invented from scratch;
- marketing TPS numbers without reproducible benchmarks.
