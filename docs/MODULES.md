# ElasticChain settlement modules

This document fixes the state boundaries for the first ElasticChain-owned Cosmos SDK modules. The deterministic algorithms remain framework-independent in `internal/elastic`; Cosmos keepers must persist and reconstruct that state rather than reimplementing protocol decisions.

## `x/elastic`

### Responsibility

`x/elastic` owns consensus state that defines the execution topology and deterministic scaling policy:

- current execution-domain prefix tree;
- next domain identifier;
- scaling policy parameters;
- finalized per-domain pressure counters;
- scheduled topology transitions;
- transition/cooldown metadata;
- later, per-domain state commitments.

It does **not** execute application transactions and must never observe process-local mempool size, CPU load, wall-clock latency or local network measurements as consensus inputs.

### Canonical logical keys

The first keeper should expose a stable logical schema equivalent to:

```text
elastic/params                         -> ScalingPolicy
elastic/topology/next_id               -> DomainID
elastic/topology/domain/<id>           -> Domain
elastic/pressure/<domain_id>           -> finalized pressure counters
elastic/transition/pending             -> optional scheduled transition
```

Physical key encoding may use Cosmos Collections, but changing encoding requires an explicit state migration.

### Import/export boundary

`internal/elastic.TopologySnapshot` is the reference export representation. A keeper export must be semantically equivalent to `Topology.Snapshot()` and import must pass `TopologySnapshot.Validate()` before committing state.

No partial topology import is allowed.

## `x/xmsg`

### Responsibility

`x/xmsg` owns canonical asynchronous cross-domain message state:

- message ID and payload commitment;
- source and destination domains;
- source nonce;
- settlement height;
- lifecycle status (`PENDING`, `FINALIZED`, `CONSUMED`);
- replay-protection index for `(source_domain, nonce)`.

### Canonical logical keys

```text
xmsg/message/<message_id>              -> CrossDomainMessage
xmsg/source_nonce/<source>/<nonce>     -> message_id
```

The source-nonce index is derived data. Import/export should persist canonical messages and deterministically rebuild the index, matching `RestoreMessageQueue`, so two mutable copies cannot silently disagree.

### State transitions

```text
submit   : absent -> PENDING
finalize : PENDING -> FINALIZED
consume  : FINALIZED -> CONSUMED
```

`CONSUMED` is terminal. Consumption must validate destination binding and settlement finality before changing state.

## Transaction boundary

For v0.1 the custom modules should expose only protocol-administration/testnet messages required to exercise state safely. Automatic scaling itself eventually executes at deterministic epoch boundaries from finalized counters, not from arbitrary user transactions.

Candidate v0.1 messages:

- `MsgSetScalingPolicy` — authority/governance gated;
- `MsgRecordDomainMetrics` — disabled for public submission in production; test harness or consensus hook only;
- `MsgSubmitCrossDomainMessage` — source-domain authenticated;
- `MsgFinalizeCrossDomainMessage` — settlement hook only;
- `MsgConsumeCrossDomainMessage` — destination-domain execution hook only.

A public testnet command must not be mistaken for the final authority model.

## Genesis

Genesis starts with exactly one active root domain:

```text
DomainID:      0
Prefix:        0
PrefixBits:    0
Active:        true
NextDomainID:  1
```

The default scaling policy comes from `internal/elastic.DefaultScalingPolicy()` unless an explicit, validated genesis policy is supplied.

Cross-domain message state is empty at genesis.

## Determinism rules

1. Consensus-path arithmetic uses integers only.
2. Iteration over maps is never serialized directly; canonical exports sort domain IDs and message IDs.
3. Imported state is validated before keeper writes occur.
4. Keeper transactions must be atomic: a failed transition leaves both primary and derived indexes unchanged.
5. Any malformed state discovered during export/import or invariant checks is fatal rather than repaired heuristically.

## v0.1 acceptance extension

Before `x/elastic` and `x/xmsg` are considered integrated:

- fresh genesis must reproduce the one-domain reference topology;
- keeper state must survive process restart;
- export -> fresh init -> import must reproduce the same snapshots;
- malformed topology import must fail;
- duplicate source nonce import must fail;
- consumed messages must remain consumed after restart/import;
- all four validators must compute identical state roots after the same transition sequence.
