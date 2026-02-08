#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PG_NAME="sprocket-v3-pg-week3-smoke"
PG_PORT="${PG_PORT:-55432}"
DB_URL="postgres://postgres:postgres@localhost:${PG_PORT}/sprocket?sslmode=disable"
API_PID=""

cleanup() {
  if [[ -n "$API_PID" ]]; then
    kill "$API_PID" >/dev/null 2>&1 || true
    wait "$API_PID" 2>/dev/null || true
  fi
  docker rm -f "$PG_NAME" >/dev/null 2>&1 || true
}
trap cleanup EXIT

if docker ps -a --format '{{.Names}}' | grep -q "^${PG_NAME}$"; then
  docker rm -f "$PG_NAME" >/dev/null
fi

cd "$ROOT_DIR"

docker run --name "$PG_NAME" \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=sprocket \
  -p "${PG_PORT}:5432" \
  -d postgres:16 >/dev/null

for i in $(seq 1 40); do
  if docker exec "$PG_NAME" pg_isready -U postgres -d sprocket >/dev/null 2>&1; then
    break
  fi
  sleep 1
  if [[ "$i" -eq 40 ]]; then
    echo "Postgres did not become ready in time" >&2
    exit 1
  fi
done

echo "Running migrations"
DATABASE_URL="$DB_URL" MIGRATIONS_DIR="./migrations" go run ./cmd/migrate >/tmp/sprocket_v3_week3_migrate.log

echo "Starting API in DB-required mode"
DATABASE_URL="$DB_URL" REQUIRE_DATABASE=true go run ./cmd/api >/tmp/sprocket_v3_week3_api.log 2>&1 &
API_PID=$!
sleep 1

suffix="$(date +%s)"
league_slug="week3-league-${suffix}"
franchise_slug="week3-franchise-${suffix}"
club_slug="week3-club-${suffix}"

assert_code() {
  local expected="$1"
  local actual="$2"
  local label="$3"
  if [[ "$actual" != "$expected" ]]; then
    echo "${label}: expected HTTP ${expected}, got ${actual}" >&2
    exit 1
  fi
}

league_code=$(curl -s -o /tmp/week3_create_league.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/leagues \
  -H 'content-type: application/json' \
  -d "{\"name\":\"Week3 League ${suffix}\",\"slug\":\"${league_slug}\"}")
assert_code 201 "$league_code" "create league"

league_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week3_create_league.json)
if [[ -z "$league_id" ]]; then
  echo "failed to parse league id" >&2
  exit 1
fi

dupe_code=$(curl -s -o /tmp/week3_dupe_league.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/leagues \
  -H 'content-type: application/json' \
  -d "{\"name\":\"Week3 League Duplicate ${suffix}\",\"slug\":\"${league_slug}\"}")
assert_code 409 "$dupe_code" "duplicate league slug"

bad_fk_code=$(curl -s -o /tmp/week3_bad_fk_franchise.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/franchises \
  -H 'content-type: application/json' \
  -d "{\"leagueId\":9999999,\"name\":\"Bad FK\",\"slug\":\"bad-fk-${suffix}\"}")
assert_code 409 "$bad_fk_code" "franchise bad FK"

franchise_code=$(curl -s -o /tmp/week3_create_franchise.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/franchises \
  -H 'content-type: application/json' \
  -d "{\"leagueId\":${league_id},\"name\":\"Week3 Franchise ${suffix}\",\"slug\":\"${franchise_slug}\"}")
assert_code 201 "$franchise_code" "create franchise"

franchise_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week3_create_franchise.json)
if [[ -z "$franchise_id" ]]; then
  echo "failed to parse franchise id" >&2
  exit 1
fi

club_code=$(curl -s -o /tmp/week3_create_club.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/clubs \
  -H 'content-type: application/json' \
  -d "{\"franchiseId\":${franchise_id},\"name\":\"Week3 Club ${suffix}\",\"slug\":\"${club_slug}\"}")
assert_code 201 "$club_code" "create club"

list_code=$(curl -s -o /tmp/week3_list_leagues.json -w '%{http_code}' http://localhost:8080/v1/leagues)
assert_code 200 "$list_code" "list leagues"

echo "Week 3 smoke passed."
