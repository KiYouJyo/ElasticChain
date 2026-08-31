# ElasticChain Security Model

ElasticChain is research software. No current milestone is suitable for securing real financial value.

This document defines the security properties the architecture is attempting to preserve as execution becomes more parallel and elastic.

## 1. Security objective

Elastic scaling is useful only if adding execution capacity does not silently weaken the security budget of each execution partition.

The desired property is:

> More execution domains increase available execution capacity while canonical settlement, validator stake and finality remain shared.

## 2. Trust assumptions by milestone

### v0.0.x

- local deterministic prototype only;
- no networking or adversarial consensus assumption;
- validates algorithms and state-machine invariants, not blockchain security.

### v0.1–v0.3

- settlement safety/liveness relies on Cosmos SDK + CometBFT assumptions and correct application wiring;
- execution remains settlement-local or directly reproducible;
- no claim of independent ElasticChain consensus innovation.

### v0.4–v0.5

Additional assumptions appear:

- execution-domain data remains available;
- domain commitments can be independently reconstructed/validated under the active milestone rules;
- state migration is correct;
- cross-domain messages are finalized and replay-safe;
- topology changes are deterministic.

### v0.6+

Additional cryptographic assumptions may include:

- soundness of the selected validity-proof system;
- correctness of verifier implementation;
- security of any setup ceremony, if the proof system requires one;
- correctness of erasure coding and data-availability sampling parameters.

These assumptions must be documented before code is treated as production-capable.

## 3. Consensus invariants

The following are release-blocking invariants:

1. **Deterministic state transition** — honest validators given identical finalized state and input produce identical next state.
2. **No floating-point consensus arithmetic** — protocol ratios use deterministic integer/fixed-point forms.
3. **No local telemetry in consensus** — mempool size, CPU load and wall-clock latency cannot directly change canonical state.
4. **Canonical topology** — one routing key maps to exactly one active execution domain.
5. **No overlapping active prefixes** — split/merge must preserve a valid prefix partition.
6. **Supply conservation** — topology changes and cross-domain movement cannot create or destroy native units except explicit issuance/burn rules.
7. **Finality before consumption** — cross-domain messages cannot execute at the destination before required settlement finality.
8. **Exactly-once delivery** — a finalized message cannot be consumed twice.
9. **Destination binding** — a message cannot be redirected to another domain.
10. **Versioned rules** — consensus-critical encoding and arithmetic cannot change silently between software releases.

## 4. Threats and mitigations

### 4.1 Byzantine settlement validators

Mitigation baseline:

- use established CometBFT consensus rather than inventing a new BFT protocol in early releases;
- preserve evidence/slashing integration;
- test restart, equivocation and malformed application transition handling.

### 4.2 Domain capture

Threat:

If each execution domain has a tiny fixed validator set, increasing domain count can make individual domains cheaper to corrupt.

Required direction:

- one shared stake pool;
- rotating validator-derived committees;
- unpredictable assignment;
- settlement-verifiable domain commitments/proofs;
- quantitative committee-capture analysis before public testnet.

### 4.3 Split/merge oscillation

Threat:

An attacker creates alternating load near a threshold, causing constant migration and expensive topology churn.

Mitigation:

- distinct scale-out and scale-in thresholds;
- long consecutive-epoch windows;
- minimum topology lifetime / cooldown in the settlement implementation;
- explicit migration cost accounting;
- adversarial oscillation benchmarks.

### 4.4 Hot-key attack

Threat:

A single account/contract remains a sequential bottleneck even when the domain around it is split.

Mitigation direction:

- recognize that partitioning does not make one mutable state object parallel;
- isolate local fee spikes;
- permit application-level partitioning;
- investigate state-object sharding only with explicit semantics;
- report benchmark limits honestly.

### 4.5 Cross-domain replay

Mitigation:

- deterministic IDs including source, destination, nonce and payload commitment;
- settlement lifecycle state;
- consumed-message registry or equivalent accumulator;
- destination verification;
- property/fuzz tests.

### 4.6 Data withholding

Threat:

An executor publishes a state commitment/proof but withholds data required by users or future state reconstruction.

Mitigation direction:

- separate execution validity from availability;
- data commitments;
- erasure coding;
- sampling;
- withholding simulation;
- no final acceptance of a domain transition without the required availability condition.

### 4.7 Nondeterministic parallel execution

Threat:

Different goroutine/thread schedules commit different transaction order or state.

Mitigation:

- deterministic conflict graph;
- canonical batch/order rules;
- serial reference implementation;
- differential testing across worker counts;
- randomized/fuzz tests.

### 4.8 Governance or upgrade capture

Mitigation direction:

- minimize mutable consensus parameters;
- delayed upgrades;
- transparent on-chain version transitions;
- reproducible binaries;
- avoid emergency administrator keys as a permanent architecture feature.

## 5. Key management

Before public testnet:

- validator consensus keys must be separated from ordinary wallet keys;
- documentation must cover backup and rotation;
- development keys must never ship as production defaults;
- CI and repository history must contain no secrets;
- faucet/testnet keys must be treated as disposable and clearly labeled.

## 6. Cryptography policy

Early ElasticChain versions MUST NOT invent custom cryptographic primitives.

When ZK/DA work begins:

- choose maintained libraries with public security review;
- pin versions;
- document curves/hash functions/transcript construction;
- add test vectors;
- isolate cryptographic adapters behind small interfaces;
- treat verifier bugs as consensus-critical.

## 7. Benchmark integrity

Security and performance claims must include the full system cost.

A TPS number is incomplete unless the benchmark documents:

- transaction workload;
- state contention;
- number of execution domains;
- validator/executor hardware;
- network topology;
- settlement cost;
- cross-domain ratio;
- proof generation/verification cost where applicable;
- DA bandwidth/cost where applicable;
- finality definition.

## 8. Release gates

No milestone may be called complete while its required tests are failing.

Before any public value-bearing network, require at minimum:

- external protocol/security review;
- dependency vulnerability review;
- property/fuzz tests for consensus-critical state machines;
- reproducible testnet bootstrap;
- documented incident response;
- key compromise procedure;
- upgrade rollback/abort procedure;
- proof that state export/import preserves canonical state;
- economic attack-cost analysis;
- explicit statement of remaining centralization points.

## 9. Responsible disclosure

A formal security contact and disclosure policy should be added before the first public adversarial testnet. Until then, security-sensitive issues should not be publicly demonstrated against any network carrying real assets because no such use is supported.
