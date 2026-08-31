# ElasticChain Digital Asset Model

Status: design target for v0.2+; not yet production implementation.

ElasticChain treats digital assets and financial contracts as first-class application workloads. The protocol should scale them without changing their ownership, authorization or supply semantics.

## 1. Asset layers

ElasticChain separates three concepts:

1. **Native settlement asset** — the chain's protocol denomination used for fees, staking and validator economics.
2. **Fungible application assets** — initially exposed through standard EVM/ERC-20 contracts.
3. **Non-fungible / multi-token assets** — initially exposed through ERC-721 and ERC-1155 contracts.

The chain should not invent a proprietary token standard until established interfaces have been implemented, benchmarked and shown to be insufficient.

## 2. Native asset invariants

At every finalized state transition:

- native balance changes must conserve supply except for explicitly defined protocol issuance/burn paths;
- fees, validator rewards, slashing and any burn must be separately attributable;
- a topology split/merge must not mint, destroy or duplicate native value;
- cross-domain movement must use an exactly-once settlement path;
- the same spend cannot be finalized in two domains.

Native monetary policy is deliberately deferred until protocol behavior is stable. Testnet `uelastic` has no implied economic value.

## 3. Fungible tokens

v0.2 targets ERC-20 compatibility including:

- `totalSupply`;
- balances;
- transfers;
- allowances/approvals;
- mint/burn where explicitly authorized by the contract;
- standard events;
- deterministic replay/export/import.

Reference contracts should include at least:

- fixed-supply token;
- owner/minter-controlled token;
- capped mintable token;
- burnable token.

### Fungible invariants

For every token contract used in protocol tests:

- total supply must equal the contract-defined aggregate supply after every finalized block;
- migration between execution domains must preserve total supply exactly;
- balances must not exist simultaneously as independently spendable copies in source and destination domains;
- allowance state must migrate consistently with balances;
- failed/reverted transactions must not mutate balances, allowance or supply.

## 4. NFTs and multi-token assets

v0.2 targets:

- ERC-721 ownership, mint, burn, transfer and approval;
- ERC-1155 fungible/non-fungible IDs and batch operations;
- collection-level and item-level metadata references;
- standard wallet/indexer discovery paths.

### NFT identity

An NFT identity is defined by its contract/collection plus token ID. Elastic scaling must never turn a topology migration into a second valid copy of the same identity.

Required invariants:

- one ERC-721 token ID has at most one current owner;
- burned token state cannot reappear through stale cross-domain messages;
- transfer authorization is checked against finalized ownership/approval state;
- cross-domain delivery is exactly once;
- a domain split/merge preserves token IDs, ownership and approval state;
- ERC-1155 balances preserve per-ID supply.

## 5. Metadata and media

Large media should not be stored directly in settlement state by default.

Reference NFT patterns should support:

- URI metadata;
- content-addressed metadata/media such as IPFS-compatible URIs;
- an optional on-chain content hash that commits to off-chain bytes;
- small fully on-chain metadata/SVG examples for testing;
- explicit authority rules for mutable metadata.

ElasticChain must distinguish **ownership finality** from **off-chain media availability**. A valid NFT does not automatically guarantee that a referenced external file remains hosted forever.

## 6. Royalties

Royalty interfaces may be supported for ecosystem compatibility, but the base protocol must not claim that creator royalties are universally enforceable. Marketplace contracts decide whether and how royalty information is honored unless a future protocol change explicitly states otherwise.

## 7. Financial / marketplace workloads

v0.3 reference contracts should include:

- escrow;
- fixed-price sale;
- NFT auction;
- constant-product AMM;
- fungible-token swap;
- liquidity provision/removal.

These contracts exist first as workload and correctness fixtures. They are not considered audited production financial products.

### Marketplace invariants

- a seller cannot transfer an asset they do not own/control;
- a sale cannot settle twice;
- payment and asset state must either both finalize according to contract semantics or revert consistently;
- an auction cannot accept a post-finalization winner change;
- escrow release/refund paths must be mutually consistent;
- AMM accounting must satisfy its documented invariant subject to fees and integer rounding.

## 8. Cross-domain asset movement

Elastic execution creates a special asset-safety requirement. A cross-domain move should conceptually follow:

```text
source lock/burn/update
        ↓
finalized settlement message
        ↓
destination consume/update
```

The concrete mechanism may vary by asset/runtime, but it must prevent a source asset and destination representation from being simultaneously spendable when the protocol intends a move rather than a bridge/wrapped asset.

The settlement layer must bind a message to:

- source domain;
- destination domain;
- source nonce;
- payload/content commitment;
- settlement/finality state;
- consumed state.

## 9. Scaling benchmark classes

Asset support is part of the scaling test plan. Required workload families include:

- independent ERC-20 transfers;
- one hot fungible token;
- independent ERC-721 mints;
- one hot NFT collection;
- ERC-1155 batch mint/transfer;
- fixed-price marketplace sales;
- auction bursts;
- one hot AMM pool;
- many independent AMM pools;
- cross-domain asset transfers.

Throughput results are only valid if post-run invariant checks confirm correct balances, supply, NFT ownership, message consumption and contract state.

## 10. Explicit non-goals before audit maturity

Before independent security review, ElasticChain should not present the following as production-ready primitives:

- algorithmic stablecoins;
- leveraged lending;
- derivatives/perpetuals;
- real-value cross-chain bridges;
- privacy-preserving asset systems using unaudited custom cryptography;
- protocol-enforced claims about legal ownership of off-chain property.

The immediate target is narrower: make standard token/NFT/market workloads correct, inspectable and elastically scalable first.
