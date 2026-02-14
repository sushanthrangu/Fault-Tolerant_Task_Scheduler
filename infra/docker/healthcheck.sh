#!/usr/bin/env sh
set -eu

HOST="${HOST:-localhost}"
PORT="${PORT:-8085}"

curl -fsS "http://${HOST}:${PORT}/healthz" >/dev/null
echo "ok"
