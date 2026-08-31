# ElasticChain Protocol Specification

Status: **Draft / research prototype**

This document defines the protocol-level invariants that implementations must preserve. It intentionally separates normative behavior from implementation-specific framework choices.

## 1. Terminology

- **Settlement layer**: canonical PoS/BFT chain that owns finality, staking, protocol parameters, topology and cross-domain commitments.
- **Execution domain**: an active leaf in the routing-prefix tree that executes a subset of transactions.
- **Epoch**: deterministic protocol interval used for collecting finalized utilization measurements and considering topology changes.
- **Domain commitment**: commitment to the post-execution state of an execution domain.
- **Cross-domain message**: canonical asynchronous message emitted by a source domain and consumed by a destination domain after settlement finality.
- **Split**: replacement of one active leaf domain by its two child prefixes.
- **Merge**: replacement of two active sibling leaf domains by their parent prefix.

## 2. Consensus determinism

Every consensus-sensitive state transition MUST be reproducible from finalized chain state.

Implementations MUST NOT use the following directly in consensus decisions:

- floating-point arithmetic;
- local mempool size;
- wall-clock processing latency;
- local CPU or memory utilization;
- local peer count;
- external HTTP/API observations;
- non-finalized measurements whose values can differ across validators.

Utilization ratios SHOULD use basis points (`0..10000`) or another explicitly specified fixed-point integer representation.

## 3. Canonical execution-domain topology

### 3.1 Root

Genesis begins with at least one active root execution domain covering the entire routing-key space.

### 3.2 Prefix routing

Each domain is represented by a binary prefix `(prefix, prefix_bits)`.

A routing key is first hashed with the protocol-selected cryptographic hash. The active domain with the longest matching prefix owns the routed state item.

### 3.3 Split

A domain may split only when:

1. it is active;
2. it is a leaf;
3. the active domain count is below `max_domains`;
4. the scale-out policy has been satisfied;
5. the transition is committed at an allowed topology-change boundary.

Splitting prefix `P/b` creates:

- left child `(P << 1)/(b+1)`;
- right child `((P << 1) | 1)/(b+1)`.

The parent becomes inactive.

### 3.4 Merge

A parent may be reactivated only when:

1. it is inactive;
2. it has exactly two active children;
3. both children are leaves;
4. both children satisfy the scale-in policy;
5. no pending migration or unconsumed protocol-critical message prevents consolidation;
6. the result would not reduce active domains below `min_domains`.

The children become inactive and the parent becomes active.

## 4. Elastic scaling policy

Protocol parameters include at minimum:

- `min_domains`;
- `max_domains`;
- `scale_out_utilization_bps`;
- `scale_in_utilization_bps`;
- `scale_out_queue_depth`;
- `scale_out_consecutive_epochs`;
- `scale_in_consecutive_epochs`.

### 4.1 Hot domain

A domain is hot for an epoch if **any** configured scale-out signal is satisfied.

The bootstrap policy uses:

- finalized gas utilization;
- finalized DA-byte utilization;
- finalized cross-domain queue depth.

### 4.2 Cold domain

A domain is cold only when **all** configured scale-in conditions are satisfied and the relevant cross-domain queue is empty.

### 4.3 Hysteresis

`scale_in_utilization_bps` MUST be lower than `scale_out_utilization_bps`.

Scale-out and scale-in require separate consecutive-epoch windows. This prevents topology oscillation under borderline load.

### 4.4 Ordering

When both split and merge candidates exist in the same epoch, scale-out actions take priority.

Within an action class, actions MUST use deterministic ordering by canonical domain identifier.

## 5. Cross-domain message protocol

### 5.1 Message identifier

A message ID commits to at least:

```text
source_domain || destination_domain || source_nonce || payload_hash
```

using the protocol-selected cryptographic hash.

### 5.2 Lifecycle

```text
PENDING → FINALIZED → CONSUMED
```

No reverse transition is permitted.

### 5.3 Submission

The source domain emits a message commitment as part of its canonical state transition.

A protocol implementation MUST prevent accidental or adversarial duplicate IDs.

### 5.4 Finalization

A message becomes finalized only after the settlement layer accepts the source commitment under its finality rules.

### 5.5 Consumption

The destination may consume a finalized message only when:

1. the destination matches the message destination;
2. the required settlement-finality condition is met;
3. the message has not previously been consumed;
4. inclusion and availability checks required by the active protocol version succeed.

Successful consumption MUST be exactly-once.

## 6. Execution commitments

Each domain transition eventually commits at minimum:

- previous state root;
- new state root;
- transaction/data commitment;
- outgoing message root;
- epoch/height reference.

The validation mechanism evolves by milestone:

- early prototype: deterministic replay/testing;
- multi-domain prototype: explicit settlement commitments and challenge tooling;
- later milestone: validity proof verification;
- later milestone: recursively aggregated proofs.

## 7. Native asset and accounting

The settlement layer is the canonical source of native-token issuance and staking accounting.

Execution domains MUST NOT independently mint settlement-native units except through an explicitly authorized protocol transition.

Cross-domain movement of the native asset MUST preserve global supply exactly.

## 8. Validator and committee model

The settlement layer maintains one canonical validator set and stake distribution.

Execution committees, once introduced, MUST be derived from that shared validator/security pool. A domain MUST NOT create an unrelated independent security budget by default.

Committee selection must be:

- deterministic from consensus-visible randomness/state;
- difficult to predict far enough in advance for cheap targeted corruption;
- regularly rotated;
- auditable from settlement state.

## 9. Upgrade model

Consensus-breaking upgrades require an explicit version transition coordinated through the settlement chain's upgrade mechanism.

Protocol parameters that affect consensus MUST be versioned and stored in canonical state.

No implementation may silently change split/merge arithmetic, routing, message-ID construction or finality rules without a protocol-version transition.

## 10. Safety invariants

At all supported milestones the following properties are mandatory:

1. one routing key maps to exactly one active domain;
2. active domain prefixes do not overlap;
3. scale decisions are deterministic;
4. no message is consumed more than once;
5. no message is consumed by the wrong destination;
6. no cross-domain message is consumed before required finality;
7. topology changes do not create or destroy native supply;
8. invalid domain commitments cannot become canonical merely because a domain operator asserts them;
9. validator-local observations cannot change consensus state;
10. protocol progress must not depend on a single privileged server.

## 11. Liveness goals

The protocol should continue making progress provided the underlying settlement BFT assumptions remain satisfied and sufficient execution/DA capacity remains available.

A hot or unavailable execution domain should degrade locally where practical instead of forcing unrelated domains to stop.

## 12. Deferred questions

The following remain research items and must be resolved before a public value-bearing network:

- exact validator/committee sampling algorithm;
- state migration protocol during split and merge;
- cryptographic proof system and trusted-setup policy, if any;
- data-availability coding and sampling parameters;
- fee-market coupling between settlement and domains;
- proposer/builder separation and MEV policy;
- execution-domain failure recovery;
- light-client proof format;
- governance minimization and emergency upgrade process.
