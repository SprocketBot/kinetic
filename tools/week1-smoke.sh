#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PG_NAME="kinetic-v3-pg-smoke"
PG_PORT="${PG_PORT:-55432}"
DB_URL="postgres://postgres:postgres@localhost:${PG_PORT}/kinetic?sslmode=disable"

cleanup() {
  docker rm -f "${PG_NAME}" >/dev/null 2>&1 || true
}
trap cleanup EXIT

if docker ps -a --format '{{.Names}}' | grep -q "^${PG_NAME}$"; then
  docker rm -f "${PG_NAME}" >/dev/null
fi

docker run --name "${PG_NAME}" \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=kinetic \
  -p "${PG_PORT}:5432" \
  -d postgres:16 >/dev/null

for i in $(seq 1 40); do
  if docker exec "${PG_NAME}" pg_isready -U postgres -d kinetic >/dev/null 2>&1; then
    break
  fi
  sleep 1
  if [[ "$i" -eq 40 ]]; then
    echo "Postgres did not become ready in time" >&2
    exit 1
  fi
done

cd "${ROOT_DIR}"

echo "Running migrations (first pass)"
DATABASE_URL="${DB_URL}" MIGRATIONS_DIR="./migrations" go run ./cmd/migrate

echo "Running migrations (second pass, should apply 0)"
DATABASE_URL="${DB_URL}" MIGRATIONS_DIR="./migrations" go run ./cmd/migrate

echo "Running tests"
go test ./...

echo "Week 1 smoke passed."

