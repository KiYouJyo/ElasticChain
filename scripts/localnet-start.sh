#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${ELASTICCHAIN_BIN:-$ROOT_DIR/build/elasticchaind}"
OUT="${ELASTICCHAIN_LOCALNET_DIR:-$ROOT_DIR/.localnet}"
LOG_DIR="$OUT/logs"
PID_FILE="$OUT/pids"

if [[ ! -x "$BIN" ]]; then
  echo "elasticchaind binary not found at $BIN; run scripts/localnet-init.sh first" >&2
  exit 1
fi

mkdir -p "$LOG_DIR"
: > "$PID_FILE"

for i in 0 1 2 3; do
  home="$OUT/node${i}/elasticchaind"
  if [[ ! -f "$home/config/genesis.json" ]]; then
    echo "missing validator home $home; run scripts/localnet-init.sh first" >&2
    exit 1
  fi

  "$BIN" start --home "$home" >"$LOG_DIR/node${i}.log" 2>&1 &
  pid=$!
  echo "$pid" >> "$PID_FILE"
  echo "started node${i} pid=$pid rpc=$((26657 + i))"
done

echo "PID file: $PID_FILE"
