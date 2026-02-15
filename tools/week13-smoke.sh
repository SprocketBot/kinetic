#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

./tools/week12-smoke.sh
./tools/quality-gate.sh

echo "Week 13 smoke passed."
