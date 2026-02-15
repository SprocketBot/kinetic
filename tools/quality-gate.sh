#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

echo "[quality] gofmt check"
unformatted="$(gofmt -l .)"
if [[ -n "$unformatted" ]]; then
  echo "The following files are not gofmt-formatted:"
  echo "$unformatted"
  exit 1
fi

echo "[quality] go vet"
go vet ./...

echo "[quality] go test"
go test ./...

echo "[quality] go build"
go build ./cmd/api
go build ./cmd/migrate

echo "Quality gate passed."
