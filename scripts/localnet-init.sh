#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${ELASTICCHAIN_BIN:-$ROOT_DIR/build/elasticchaind}"
OUT="${ELASTICCHAIN_LOCALNET_DIR:-$ROOT_DIR/.localnet}"
CHAIN_ID="${ELASTICCHAIN_CHAIN_ID:-elastic-local-1}"

mkdir -p "$(dirname "$BIN")"
if [[ ! -x "$BIN" ]]; then
  (cd "$ROOT_DIR" && go build -o "$BIN" ./cmd/elasticchaind)
fi

rm -rf "$OUT"
"$BIN" testnet init-files \
  --validator-count 4 \
  --output-dir "$OUT" \
  --single-host \
  --staking-denom uelastic \
  --minimum-gas-prices 0uelastic \
  --node-daemon-home elasticchaind \
  --chain-id "$CHAIN_ID" \
  --keyring-backend test \
  --commit-timeout 1s

# The upstream testnet helper lets gentxs use a custom staking denomination but
# leaves several default genesis parameters at "stake". Canonicalize every
# exact denom string before starting validators so the generated testnet is a
# genuine uelastic PoS network.
python3 - "$OUT" <<'PY'
import glob
import json
import pathlib
import sys

root = pathlib.Path(sys.argv[1])
paths = sorted(glob.glob(str(root / "node*" / "elasticchaind" / "config" / "genesis.json")))
if len(paths) != 4:
    raise SystemExit(f"expected 4 genesis files, found {len(paths)}")


def rewrite(value):
    if isinstance(value, dict):
        return {key: rewrite(item) for key, item in value.items()}
    if isinstance(value, list):
        return [rewrite(item) for item in value]
    if value == "stake":
        return "uelastic"
    return value

canonical = None
for path in paths:
    with open(path, "r", encoding="utf-8") as handle:
        genesis = rewrite(json.load(handle))

    encoded = json.dumps(genesis, indent=2, sort_keys=True) + "\n"
    with open(path, "w", encoding="utf-8") as handle:
        handle.write(encoded)

    if '"stake"' in encoded:
        raise SystemExit(f"legacy stake denomination remains in {path}")
    if canonical is None:
        canonical = encoded
    elif encoded != canonical:
        raise SystemExit("validator genesis files are not byte-identical after canonicalization")

print(f"canonicalized {len(paths)} validator genesis files to uelastic")
PY

printf 'ElasticChain localnet initialized:\n  chain-id: %s\n  validators: 4\n  denom: uelastic\n  directory: %s\n' "$CHAIN_ID" "$OUT"
