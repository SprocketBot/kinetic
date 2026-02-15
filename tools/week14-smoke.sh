#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

./tools/ci-verify.sh
./tools/week12-smoke.sh

echo "Week 14 smoke passed."
