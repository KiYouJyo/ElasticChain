#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${ELASTICCHAIN_BIN:-$ROOT_DIR/build/elasticchaind}"
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
            statuses.append((port, result["node_info"]["network"], int(result["sync_info"]["latest_block_height"])))
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

app_hashes = []
for port in ports:
    with urllib.request.urlopen(f"http://127.0.0.1:{port}/commit?height=2", timeout=3) as response:
        commit = json.load(response)["result"]
    app_hashes.append(commit["signed_header"]["header"]["app_hash"])
if len(set(app_hashes)) != 1:
    raise SystemExit(f"validators disagree on height-2 AppHash: {app_hashes}")
print(f"height=2 shared_app_hash={app_hashes[0]}")
PY

# Restart the exact same node homes. InitChain does not run again; successful
# block production proves that elastic/xmsg state was loaded from committed
# stores and passed BeginBlock restore/validation.
bash "$ROOT_DIR/scripts/localnet-stop.sh"
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
        heights = []
        for port in ports:
            with urllib.request.urlopen(f"http://127.0.0.1:{port}/status", timeout=2) as response:
                heights.append(int(json.load(response)["result"]["sync_info"]["latest_block_height"]))
        if min(heights) >= 4:
            break
    except (OSError, KeyError, ValueError, urllib.error.URLError) as exc:
        last_error = exc
    time.sleep(1)
else:
    raise SystemExit(f"restarted localnet did not reach height 4: {last_error}")

app_hashes = []
for port in ports:
    with urllib.request.urlopen(f"http://127.0.0.1:{port}/commit?height=3", timeout=3) as response:
        commit = json.load(response)["result"]
    app_hashes.append(commit["signed_header"]["header"]["app_hash"])
if len(set(app_hashes)) != 1:
    raise SystemExit(f"validators disagree after restart at height 3: {app_hashes}")
print(f"restart persistence verified; height=3 shared_app_hash={app_hashes[0]}")
PY

# Stop the network before opening the application DB for export.
bash "$ROOT_DIR/scripts/localnet-stop.sh"
EXPORT_GENESIS="$OUT/exported-genesis.json"
"$BIN" export \
  --home "$OUT/node0/elasticchaind" \
  --output-document "$EXPORT_GENESIS"

# Prove the portable genesis actually contains ElasticChain-owned consensus
# state rather than only the upstream Cosmos modules.
EXPORT_INITIAL_HEIGHT="$(python3 - "$EXPORT_GENESIS" <<'PY'
import json
import sys

path = sys.argv[1]
with open(path, "r", encoding="utf-8") as handle:
    genesis = json.load(handle)

elastic = genesis.get("app_state", {}).get("elasticchain")
if not isinstance(elastic, dict):
    raise SystemExit("exported genesis is missing app_state.elasticchain")
if elastic.get("version") != 1:
    raise SystemExit(f"unexpected ElasticChain genesis version: {elastic.get('version')}")

topology = elastic.get("topology", {})
domains = topology.get("Domains") or topology.get("domains")
if not isinstance(domains, list) or len(domains) != 1:
    raise SystemExit(f"expected one root execution domain in export, got {domains}")

messages = elastic.get("messages", {})
message_list = messages.get("Messages") if isinstance(messages, dict) else None
if message_list is None and isinstance(messages, dict):
    message_list = messages.get("messages")
if not isinstance(message_list, list) or message_list:
    raise SystemExit(f"expected empty exported message queue, got {message_list}")

initial_height = int(genesis.get("initial_height", "0"))
if initial_height <= 0:
    raise SystemExit(f"invalid exported initial height {initial_height}")
print(initial_height)
PY
)"
echo "exported portable ElasticChain genesis initial_height=$EXPORT_INITIAL_HEIGHT"

# Re-bootstrap all four validators from the exported genesis. Preserve only
# node identity, validator keys/configuration, and each validator's anti-double-
# sign state; delete application/block databases so InitChain must reconstruct
# state from the exported document.
for i in 0 1 2 3; do
  home="$OUT/node${i}/elasticchaind"
  signer_state="$OUT/node${i}/priv_validator_state.json"
  cp "$home/data/priv_validator_state.json" "$signer_state"
  rm -rf "$home/data"
  mkdir -p "$home/data"
  cp "$signer_state" "$home/data/priv_validator_state.json"
  rm -f "$signer_state"
  cp "$EXPORT_GENESIS" "$home/config/genesis.json"
done

bash "$ROOT_DIR/scripts/localnet-start.sh"

python3 - "$EXPORT_INITIAL_HEIGHT" <<'PY'
import json
import sys
import time
import urllib.error
import urllib.request

initial_height = int(sys.argv[1])
ports = [26657, 26658, 26659, 26660]
deadline = time.time() + 90
last_error = None

while time.time() < deadline:
    try:
        statuses = []
        for port in ports:
            with urllib.request.urlopen(f"http://127.0.0.1:{port}/status", timeout=2) as response:
                result = json.load(response)["result"]
            statuses.append((result["node_info"]["network"], int(result["sync_info"]["latest_block_height"])))
        if all(chain == "elastic-local-1" for chain, _ in statuses) and min(height for _, height in statuses) >= initial_height:
            break
    except (OSError, KeyError, ValueError, urllib.error.URLError) as exc:
        last_error = exc
    time.sleep(1)
else:
    raise SystemExit(f"export/import network did not resume at height {initial_height}: {last_error}")

app_hashes = []
for port in ports:
    with urllib.request.urlopen(f"http://127.0.0.1:{port}/commit?height={initial_height}", timeout=3) as response:
        result = json.load(response)["result"]
    app_hashes.append(result["signed_header"]["header"]["app_hash"])
if len(set(app_hashes)) != 1:
    raise SystemExit(f"validators disagree after export/import: {app_hashes}")
print(f"export/import rebootstrap verified; height={initial_height} shared_app_hash={app_hashes[0]}")
PY

SUCCESS=1
echo "four-validator ElasticApp restart/export/import/finality smoke test passed"
