#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COVERAGE_MIN="${COVERAGE_MIN:-23.5}"
COVERAGE_FILE="$(mktemp)"
trap 'rm -f "$COVERAGE_FILE"' EXIT

cd "$ROOT_DIR"

echo "[go-test] unit tests and coverage"
go test ./cmd/... ./internal/... -coverprofile="$COVERAGE_FILE"

coverage="$(go tool cover -func="$COVERAGE_FILE" | awk '/^total:/ {sub(/%/, "", $3); print $3}')"
if [[ -z "$coverage" ]]; then
  echo "Unable to read total Go coverage" >&2
  exit 1
fi

printf '[go-test] total coverage: %s%% (minimum %s%%)\n' "$coverage" "$COVERAGE_MIN"
awk -v actual="$coverage" -v minimum="$COVERAGE_MIN" 'BEGIN { exit !(actual + 0 >= minimum + 0) }' || {
  echo "Go coverage is below the required minimum" >&2
  exit 1
}

