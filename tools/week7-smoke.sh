#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PG_NAME="kinetic-v3-pg-week7-smoke"
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
    -e POSTGRES_DB=kinetic \
    -p "${candidate_port}:5432" \
    -d postgres:16 >/dev/null 2>/tmp/kinetic_v3_week7_docker_err.log; then
    PG_PORT="${candidate_port}"
    DB_URL="postgres://postgres:postgres@localhost:${PG_PORT}/kinetic?sslmode=disable"
    break
  fi
done

if [[ -z "${DB_URL}" ]]; then
  echo "Unable to start Postgres container on ports ${PG_PORT}-${PG_PORT}+4" >&2
  cat /tmp/kinetic_v3_week7_docker_err.log >&2 || true
  exit 1
fi

for i in $(seq 1 40); do
  if docker exec "$PG_NAME" pg_isready -U postgres -d kinetic >/dev/null 2>&1; then
    break
  fi
  sleep 1
  if [[ "$i" -eq 40 ]]; then
    echo "Postgres did not become ready in time" >&2
    exit 1
  fi
done

echo "Running migrations"
DATABASE_URL="$DB_URL" MIGRATIONS_DIR="./migrations" go run ./cmd/migrate >/tmp/kinetic_v3_week7_migrate.log

echo "Starting API in DB-required mode"
DATABASE_URL="$DB_URL" REQUIRE_DATABASE=true go run ./cmd/api >/tmp/kinetic_v3_week7_api.log 2>&1 &
API_PID=$!
sleep 1

suffix="$(date +%s)"
league_slug="week7-league-${suffix}"
franchise_slug="week7-franchise-${suffix}"
club_a_slug="week7-club-a-${suffix}"
club_b_slug="week7-club-b-${suffix}"
team_a_slug="week7-team-a-${suffix}"
team_b_slug="week7-team-b-${suffix}"
season_slug="week7-season-${suffix}"

assert_code() {
  local expected="$1"
  local actual="$2"
  local label="$3"
  if [[ "$actual" != "$expected" ]]; then
    echo "${label}: expected HTTP ${expected}, got ${actual}" >&2
    exit 1
  fi
}

league_code=$(curl -s -o /tmp/week7_create_league.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/leagues \
  -H 'content-type: application/json' \
  -d "{\"name\":\"Week7 League ${suffix}\",\"slug\":\"${league_slug}\"}")
assert_code 201 "$league_code" "create league"
league_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week7_create_league.json)

franchise_code=$(curl -s -o /tmp/week7_create_franchise.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/franchises \
  -H 'content-type: application/json' \
  -d "{\"leagueId\":${league_id},\"name\":\"Week7 Franchise ${suffix}\",\"slug\":\"${franchise_slug}\"}")
assert_code 201 "$franchise_code" "create franchise"
franchise_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week7_create_franchise.json)

club_a_code=$(curl -s -o /tmp/week7_create_club_a.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/clubs \
  -H 'content-type: application/json' \
  -d "{\"franchiseId\":${franchise_id},\"name\":\"Week7 Club A ${suffix}\",\"slug\":\"${club_a_slug}\"}")
assert_code 201 "$club_a_code" "create club A"
club_a_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week7_create_club_a.json)

club_b_code=$(curl -s -o /tmp/week7_create_club_b.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/clubs \
  -H 'content-type: application/json' \
  -d "{\"franchiseId\":${franchise_id},\"name\":\"Week7 Club B ${suffix}\",\"slug\":\"${club_b_slug}\"}")
assert_code 201 "$club_b_code" "create club B"
club_b_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week7_create_club_b.json)

team_a_code=$(curl -s -o /tmp/week7_create_team_a.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/teams \
  -H 'content-type: application/json' \
  -d "{\"clubId\":${club_a_id},\"name\":\"Week7 Team A ${suffix}\",\"slug\":\"${team_a_slug}\"}")
assert_code 201 "$team_a_code" "create team A"
team_a_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week7_create_team_a.json)

team_b_code=$(curl -s -o /tmp/week7_create_team_b.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/teams \
  -H 'content-type: application/json' \
  -d "{\"clubId\":${club_b_id},\"name\":\"Week7 Team B ${suffix}\",\"slug\":\"${team_b_slug}\"}")
assert_code 201 "$team_b_code" "create team B"
team_b_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week7_create_team_b.json)

season_code=$(curl -s -o /tmp/week7_create_season.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/seasons \
  -H 'content-type: application/json' \
  -d "{\"name\":\"Week7 Season ${suffix}\",\"slug\":\"${season_slug}\"}")
assert_code 201 "$season_code" "create season"
season_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week7_create_season.json)

group_code=$(curl -s -o /tmp/week7_create_group.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/schedule-groups \
  -H 'content-type: application/json' \
  -d "{\"seasonId\":${season_id},\"name\":\"Week7 Group ${suffix}\",\"sequence\":1}")
assert_code 201 "$group_code" "create schedule group"
group_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week7_create_group.json)

bad_fixture_code=$(curl -s -o /tmp/week7_bad_fixture.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/fixtures \
  -H 'content-type: application/json' \
  -d "{\"scheduleGroupId\":${group_id},\"homeClubId\":${club_a_id},\"awayClubId\":${club_a_id}}")
assert_code 400 "$bad_fixture_code" "invalid fixture clubs"

fixture_code=$(curl -s -o /tmp/week7_create_fixture.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/fixtures \
  -H 'content-type: application/json' \
  -d "{\"scheduleGroupId\":${group_id},\"homeClubId\":${club_a_id},\"awayClubId\":${club_b_id}}")
assert_code 201 "$fixture_code" "create fixture"
fixture_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week7_create_fixture.json)

planned_match_code=$(curl -s -o /tmp/week7_match_planned.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/matches \
  -H 'content-type: application/json' \
  -d "{\"fixtureId\":${fixture_id},\"homeTeamId\":${team_a_id},\"awayTeamId\":${team_b_id},\"state\":\"planned\"}")
assert_code 201 "$planned_match_code" "create planned match"

bad_ready_match_code=$(curl -s -o /tmp/week7_match_bad_ready.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/matches \
  -H 'content-type: application/json' \
  -d "{\"fixtureId\":${fixture_id},\"homeTeamId\":${team_a_id},\"awayTeamId\":${team_b_id},\"state\":\"ready\"}")
assert_code 400 "$bad_ready_match_code" "ready requires ratified schedule"

ready_match_code=$(curl -s -o /tmp/week7_match_ready.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/matches \
  -H 'content-type: application/json' \
  -d "{\"fixtureId\":${fixture_id},\"homeTeamId\":${team_a_id},\"awayTeamId\":${team_b_id},\"state\":\"ready\",\"scheduledFor\":\"2030-01-01T10:00:00Z\",\"homeTimeRatifiedAt\":\"2030-01-01T08:00:00Z\",\"awayTimeRatifiedAt\":\"2030-01-01T09:00:00Z\"}")
assert_code 201 "$ready_match_code" "create ready match"

list_seasons_code=$(curl -s -o /tmp/week7_list_seasons.json -w '%{http_code}' http://localhost:8080/v1/seasons)
assert_code 200 "$list_seasons_code" "list seasons"

list_groups_code=$(curl -s -o /tmp/week7_list_groups.json -w '%{http_code}' http://localhost:8080/v1/schedule-groups)
assert_code 200 "$list_groups_code" "list schedule groups"

list_fixtures_code=$(curl -s -o /tmp/week7_list_fixtures.json -w '%{http_code}' http://localhost:8080/v1/fixtures)
assert_code 200 "$list_fixtures_code" "list fixtures"

list_matches_code=$(curl -s -o /tmp/week7_list_matches.json -w '%{http_code}' http://localhost:8080/v1/matches)
assert_code 200 "$list_matches_code" "list matches"

echo "Week 7 smoke passed."
