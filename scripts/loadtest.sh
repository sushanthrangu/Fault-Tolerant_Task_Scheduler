#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8086}"
N="${N:-50}"

echo "Load test: creating ${N} jobs against ${BASE_URL}"

for i in $(seq 1 "$N"); do
  curl -sS -X POST "${BASE_URL}/jobs" \
    -H "Content-Type: application/json" \
    -H "Idempotency-Key: load-${i}" \
    -d "{\"type\":\"demo\",\"payload\":{\"i\":${i}},\"max_attempts\":3}" >/dev/null &
done

wait
echo "done"
