#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

./tools/quality-gate.sh

echo "[ci] staticcheck"
go install honnef.co/go/tools/cmd/staticcheck@latest
"$(go env GOPATH)/bin/staticcheck" ./...

echo "[ci] go test -race"
go test -race ./...

echo "CI verification passed."
