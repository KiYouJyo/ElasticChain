#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="${ELASTICCHAIN_LOCALNET_DIR:-$ROOT_DIR/.localnet}"
PID_FILE="$OUT/pids"

if [[ ! -f "$PID_FILE" ]]; then
  exit 0
fi

while IFS= read -r pid; do
  [[ -z "$pid" ]] && continue
  if kill -0 "$pid" 2>/dev/null; then
    kill "$pid" 2>/dev/null || true
  fi
done < "$PID_FILE"

for _ in $(seq 1 20); do
  alive=0
  while IFS= read -r pid; do
    [[ -z "$pid" ]] && continue
    if kill -0 "$pid" 2>/dev/null; then
      alive=1
      break
    fi
  done < "$PID_FILE"
  [[ "$alive" -eq 0 ]] && break
  sleep 0.25
done

while IFS= read -r pid; do
  [[ -z "$pid" ]] && continue
  if kill -0 "$pid" 2>/dev/null; then
    kill -9 "$pid" 2>/dev/null || true
  fi
done < "$PID_FILE"

rm -f "$PID_FILE"
