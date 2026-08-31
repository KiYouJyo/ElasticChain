# Local settlement network

ElasticChain v0.1 is bringing up a real PoS/BFT settlement network before replacing every framework component with ElasticChain-owned modules.

## Current bootstrap boundary

The daemon currently reuses the Cosmos SDK `simapp` application wiring from the security-fixed `v0.54.4` line for the standard settlement responsibilities:

- accounts and bank transfers;
- staking and validator-set updates;
- mint/distribution;
- slashing/evidence;
- governance/upgrade plumbing;
- CometBFT networking and finality;
- CLI, keyring, RPC, gRPC and genesis tooling.

ElasticChain-specific deterministic execution-domain topology, split/merge planning and cross-domain message rules remain implemented in `internal/elastic`.

This reuse is intentionally temporary. The next v0.1 step replaces the reference application wiring with ElasticChain-owned `x/elastic` and `x/xmsg` modules while retaining the audited/maintained Cosmos SDK consensus and staking substrate.

## Native test denomination

The local network uses:

- chain ID: `elastic-local-1`;
- staking/native test denomination: `uelastic`;
- validators: 4;
- commit timeout: 1 second.

`uelastic` is a test denomination only. This repository does not define a public token sale, monetary value, or production mainnet distribution.

## Build

```bash
make build
```

The daemon is written to `build/elasticchaind`.

## Initialize four validators

```bash
make localnet-init
```

The generated validator homes live under `.localnet/node0` through `.localnet/node3`.

The upstream testnet generator currently leaves some module genesis defaults at the literal denomination `stake` even when gentxs use a custom staking denomination. `scripts/localnet-init.sh` therefore performs a deterministic exact-string genesis rewrite to `uelastic` and verifies that all four canonical genesis files are byte-identical before startup.

## Start and stop

```bash
make localnet-start
make localnet-stop
```

On a single host the RPC endpoints are:

- node0: `127.0.0.1:26657`
- node1: `127.0.0.1:26658`
- node2: `127.0.0.1:26659`
- node3: `127.0.0.1:26660`

Logs are stored under `.localnet/logs`.

## Automated acceptance smoke test

```bash
make localnet-smoke
```

The smoke test:

1. builds `elasticchaind`;
2. deterministically generates four validator homes;
3. canonicalizes the genesis denomination to `uelastic`;
4. starts all four validator processes;
5. waits until every RPC reports chain `elastic-local-1` at height >= 2;
6. checks that the active validator set contains four validators;
7. terminates all node processes.

This proves that the v0.1 bootstrap is a real multi-validator BFT chain rather than an in-memory scaling simulation.

## What this does not prove yet

Passing the localnet test does **not** yet prove the ElasticChain scaling design. In particular, the current settlement application does not persist `internal/elastic` topology or cross-domain message state. Those become consensus state when `x/elastic` and `x/xmsg` land.

Do not use this build to secure real financial value.
