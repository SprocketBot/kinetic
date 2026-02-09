#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PG_NAME="sprocket-v3-pg-week5-smoke"
PG_PORT="${PG_PORT:-55432}"
DB_URL=""
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

for offset in 0 1 2 3 4; do
  candidate_port=$((PG_PORT + offset))
  if docker run --name "$PG_NAME" \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_PASSWORD=postgres \
    -e POSTGRES_DB=sprocket \
    -p "${candidate_port}:5432" \
    -d postgres:16 >/dev/null 2>/tmp/sprocket_v3_week5_docker_err.log; then
    PG_PORT="${candidate_port}"
    DB_URL="postgres://postgres:postgres@localhost:${PG_PORT}/sprocket?sslmode=disable"
    break
  fi
done

if [[ -z "${DB_URL}" ]]; then
  echo "Unable to start Postgres container on ports ${PG_PORT}-${PG_PORT}+4" >&2
  cat /tmp/sprocket_v3_week5_docker_err.log >&2 || true
  exit 1
fi

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
DATABASE_URL="$DB_URL" MIGRATIONS_DIR="./migrations" go run ./cmd/migrate >/tmp/sprocket_v3_week5_migrate.log

echo "Starting API in DB-required mode"
DATABASE_URL="$DB_URL" REQUIRE_DATABASE=true go run ./cmd/api >/tmp/sprocket_v3_week5_api.log 2>&1 &
API_PID=$!
sleep 1

suffix="$(date +%s)"
league_slug="week5-league-${suffix}"
franchise_slug="week5-franchise-${suffix}"
club_slug="week5-club-${suffix}"
team_slug="week5-team-${suffix}"
player_slug="week5-player-${suffix}"
player_two_slug="week5-player-two-${suffix}"

assert_code() {
  local expected="$1"
  local actual="$2"
  local label="$3"
  if [[ "$actual" != "$expected" ]]; then
    echo "${label}: expected HTTP ${expected}, got ${actual}" >&2
    exit 1
  fi
}

league_code=$(curl -s -o /tmp/week5_create_league.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/leagues \
  -H 'content-type: application/json' \
  -d "{\"name\":\"Week5 League ${suffix}\",\"slug\":\"${league_slug}\"}")
assert_code 201 "$league_code" "create league"
league_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week5_create_league.json)

franchise_code=$(curl -s -o /tmp/week5_create_franchise.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/franchises \
  -H 'content-type: application/json' \
  -d "{\"leagueId\":${league_id},\"name\":\"Week5 Franchise ${suffix}\",\"slug\":\"${franchise_slug}\"}")
assert_code 201 "$franchise_code" "create franchise"
franchise_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week5_create_franchise.json)

club_code=$(curl -s -o /tmp/week5_create_club.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/clubs \
  -H 'content-type: application/json' \
  -d "{\"franchiseId\":${franchise_id},\"name\":\"Week5 Club ${suffix}\",\"slug\":\"${club_slug}\"}")
assert_code 201 "$club_code" "create club"
club_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week5_create_club.json)

team_code=$(curl -s -o /tmp/week5_create_team.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/teams \
  -H 'content-type: application/json' \
  -d "{\"clubId\":${club_id},\"name\":\"Week5 Team ${suffix}\",\"slug\":\"${team_slug}\"}")
assert_code 201 "$team_code" "create team"
team_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week5_create_team.json)

player_code=$(curl -s -o /tmp/week5_create_player.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/players \
  -H 'content-type: application/json' \
  -d "{\"teamId\":${team_id},\"displayName\":\"Week5 Player ${suffix}\",\"slug\":\"${player_slug}\"}")
assert_code 201 "$player_code" "create player"

player_two_code=$(curl -s -o /tmp/week5_create_player_two.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/players \
  -H 'content-type: application/json' \
  -d "{\"teamId\":${team_id},\"displayName\":\"Week5 Player Two ${suffix}\",\"slug\":\"${player_two_slug}\"}")
assert_code 201 "$player_two_code" "create second player"
player_two_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week5_create_player_two.json)

membership_code=$(curl -s -o /tmp/week5_create_membership.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/roster-memberships \
  -H 'content-type: application/json' \
  -d "{\"playerId\":${player_two_id},\"teamId\":${team_id}}")
assert_code 201 "$membership_code" "create roster membership"

dup_membership_code=$(curl -s -o /tmp/week5_duplicate_membership.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/roster-memberships \
  -H 'content-type: application/json' \
  -d "{\"playerId\":${player_two_id},\"teamId\":${team_id}}")
assert_code 409 "$dup_membership_code" "duplicate active roster membership"

bad_membership_fk_code=$(curl -s -o /tmp/week5_bad_membership_fk.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/roster-memberships \
  -H 'content-type: application/json' \
  -d "{\"playerId\":9999999,\"teamId\":${team_id}}")
assert_code 409 "$bad_membership_fk_code" "bad roster membership FK"

bad_membership_validation_code=$(curl -s -o /tmp/week5_bad_membership_validation.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/roster-memberships \
  -H 'content-type: application/json' \
  -d '{"playerId":0,"teamId":1}')
assert_code 400 "$bad_membership_validation_code" "bad roster membership payload"

list_memberships_code=$(curl -s -o /tmp/week5_list_memberships.json -w '%{http_code}' http://localhost:8080/v1/roster-memberships)
assert_code 200 "$list_memberships_code" "list roster memberships"

echo "Week 5 smoke passed."
