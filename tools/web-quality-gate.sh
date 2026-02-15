#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR/web/client"

echo "[web-quality] npm ci"
npm ci

echo "[web-quality] lint"
npm run lint

echo "[web-quality] typecheck"
npm run typecheck

echo "[web-quality] unit tests"
npm run test

echo "[web-quality] build"
npm run build

echo "Web quality gate passed."
