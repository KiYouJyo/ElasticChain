#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="${ELASTICCHAIN_LOCALNET_DIR:-$ROOT_DIR/.localnet-smoke}"
export ELASTICCHAIN_LOCALNET_DIR="$OUT"
SUCCESS=0

cleanup() {
  if [[ "$SUCCESS" -ne 1 && -d "$OUT/logs" ]]; then
    for log in "$OUT"/logs/node*.log; do
      [[ -f "$log" ]] || continue
      echo "===== ${log#$ROOT_DIR/} (tail) =====" >&2
      tail -n 120 "$log" >&2 || true
    done
  fi
  bash "$ROOT_DIR/scripts/localnet-stop.sh" || true
  if [[ "${KEEP_LOCALNET:-0}" != "1" ]]; then
    rm -rf "$OUT"
  fi
}
trap cleanup EXIT

bash "$ROOT_DIR/scripts/localnet-init.sh"
bash "$ROOT_DIR/scripts/localnet-start.sh"

python3 - <<'PY'
import json
import time
import urllib.error
import urllib.request

ports = [26657, 26658, 26659, 26660]
deadline = time.time() + 90
last_error = None

while time.time() < deadline:
    try:
        statuses = []
        for port in ports:
            with urllib.request.urlopen(f"http://127.0.0.1:{port}/status", timeout=2) as response:
                payload = json.load(response)
            result = payload["result"]
            height = int(result["sync_info"]["latest_block_height"])
            chain_id = result["node_info"]["network"]
            statuses.append((port, chain_id, height))

        if all(chain_id == "elastic-local-1" for _, chain_id, _ in statuses) and min(height for _, _, height in statuses) >= 2:
            break
    except (OSError, KeyError, ValueError, urllib.error.URLError) as exc:
        last_error = exc
    time.sleep(1)
else:
    raise SystemExit(f"localnet did not reach height 2 on all validators: {last_error}")

for port, chain_id, height in statuses:
    print(f"rpc={port} chain={chain_id} height={height}")

with urllib.request.urlopen("http://127.0.0.1:26657/validators?per_page=100", timeout=3) as response:
    validators = json.load(response)["result"]["validators"]
if len(validators) != 4:
    raise SystemExit(f"expected 4 validators, got {len(validators)}")
print("validator set size=4")
PY

SUCCESS=1
echo "four-validator PoS/BFT localnet smoke test passed"
