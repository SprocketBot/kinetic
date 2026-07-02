#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PG_NAME="kinetic-v3-pg-week9-smoke"
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
    -d postgres:16 >/dev/null 2>/tmp/kinetic_v3_week9_docker_err.log; then
    PG_PORT="${candidate_port}"
    DB_URL="postgres://postgres:postgres@localhost:${PG_PORT}/kinetic?sslmode=disable"
    break
  fi
done

if [[ -z "${DB_URL}" ]]; then
  echo "Unable to start Postgres container on ports ${PG_PORT}-${PG_PORT}+4" >&2
  cat /tmp/kinetic_v3_week9_docker_err.log >&2 || true
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
DATABASE_URL="$DB_URL" MIGRATIONS_DIR="./migrations" go run ./cmd/migrate >/tmp/kinetic_v3_week9_migrate.log

echo "Starting API in DB-required mode"
DATABASE_URL="$DB_URL" REQUIRE_DATABASE=true go run ./cmd/api >/tmp/kinetic_v3_week9_api.log 2>&1 &
API_PID=$!
sleep 1

suffix="$(date +%s)"
league_slug="week9-league-${suffix}"
franchise_slug="week9-franchise-${suffix}"
club_a_slug="week9-club-a-${suffix}"
club_b_slug="week9-club-b-${suffix}"
team_a_slug="week9-team-a-${suffix}"
team_b_slug="week9-team-b-${suffix}"
queue_slug="week9-queue-${suffix}"

assert_code() {
  local expected="$1"
  local actual="$2"
  local label="$3"
  if [[ "$actual" != "$expected" ]]; then
    echo "${label}: expected HTTP ${expected}, got ${actual}" >&2
    exit 1
  fi
}

league_code=$(curl -s -o /tmp/week9_create_league.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/leagues \
  -H 'content-type: application/json' \
  -d "{\"name\":\"Week9 League ${suffix}\",\"slug\":\"${league_slug}\"}")
assert_code 201 "$league_code" "create league"
league_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week9_create_league.json)

franchise_code=$(curl -s -o /tmp/week9_create_franchise.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/franchises \
  -H 'content-type: application/json' \
  -d "{\"leagueId\":${league_id},\"name\":\"Week9 Franchise ${suffix}\",\"slug\":\"${franchise_slug}\"}")
assert_code 201 "$franchise_code" "create franchise"
franchise_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week9_create_franchise.json)

club_a_code=$(curl -s -o /tmp/week9_create_club_a.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/clubs \
  -H 'content-type: application/json' \
  -d "{\"franchiseId\":${franchise_id},\"name\":\"Week9 Club A ${suffix}\",\"slug\":\"${club_a_slug}\"}")
assert_code 201 "$club_a_code" "create club A"
club_a_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week9_create_club_a.json)

club_b_code=$(curl -s -o /tmp/week9_create_club_b.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/clubs \
  -H 'content-type: application/json' \
  -d "{\"franchiseId\":${franchise_id},\"name\":\"Week9 Club B ${suffix}\",\"slug\":\"${club_b_slug}\"}")
assert_code 201 "$club_b_code" "create club B"
club_b_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week9_create_club_b.json)

team_a_code=$(curl -s -o /tmp/week9_create_team_a.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/teams \
  -H 'content-type: application/json' \
  -d "{\"clubId\":${club_a_id},\"name\":\"Week9 Team A ${suffix}\",\"slug\":\"${team_a_slug}\"}")
assert_code 201 "$team_a_code" "create team A"
team_a_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week9_create_team_a.json)

team_b_code=$(curl -s -o /tmp/week9_create_team_b.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/teams \
  -H 'content-type: application/json' \
  -d "{\"clubId\":${club_b_id},\"name\":\"Week9 Team B ${suffix}\",\"slug\":\"${team_b_slug}\"}")
assert_code 201 "$team_b_code" "create team B"
team_b_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week9_create_team_b.json)

queue_code=$(curl -s -o /tmp/week9_create_queue.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/queues \
  -H 'content-type: application/json' \
  -d "{\"name\":\"Week9 Queue ${suffix}\",\"slug\":\"${queue_slug}\"}")
assert_code 201 "$queue_code" "create queue"
queue_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week9_create_queue.json)

enqueue_a_code=$(curl -s -o /tmp/week9_enqueue_a.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/queue-entries \
  -H 'content-type: application/json' \
  -d "{\"queueId\":${queue_id},\"teamId\":${team_a_id}}")
assert_code 201 "$enqueue_a_code" "enqueue team A"

enqueue_b_code=$(curl -s -o /tmp/week9_enqueue_b.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/queue-entries \
  -H 'content-type: application/json' \
  -d "{\"queueId\":${queue_id},\"teamId\":${team_b_id}}")
assert_code 201 "$enqueue_b_code" "enqueue team B"

promote_code=$(curl -s -o /tmp/week9_promote_ok.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/scrim-promotions \
  -H 'content-type: application/json' \
  -d "{\"queueId\":${queue_id}}")
assert_code 201 "$promote_code" "promote queue to scrim"
scrim_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week9_promote_ok.json)

in_progress_code=$(curl -s -o /tmp/week9_scrim_in_progress.json -w '%{http_code}' \
  -X PATCH http://localhost:8080/v1/scrims \
  -H 'content-type: application/json' \
  -d "{\"scrimId\":${scrim_id},\"state\":\"in_progress\"}")
assert_code 200 "$in_progress_code" "transition scrim to in_progress"

closed_code=$(curl -s -o /tmp/week9_scrim_closed.json -w '%{http_code}' \
  -X PATCH http://localhost:8080/v1/scrims \
  -H 'content-type: application/json' \
  -d "{\"scrimId\":${scrim_id},\"state\":\"closed\"}")
assert_code 200 "$closed_code" "transition scrim to closed"

invalid_transition_code=$(curl -s -o /tmp/week9_scrim_invalid_transition.json -w '%{http_code}' \
  -X PATCH http://localhost:8080/v1/scrims \
  -H 'content-type: application/json' \
  -d "{\"scrimId\":${scrim_id},\"state\":\"in_progress\"}")
assert_code 409 "$invalid_transition_code" "reject closed -> in_progress"

list_decisions_code=$(curl -s -o /tmp/week9_list_decisions.json -w '%{http_code}' http://localhost:8080/v1/matchmaking-decisions)
assert_code 200 "$list_decisions_code" "list matchmaking decisions"
if ! grep -q '"orderingStrategy"' /tmp/week9_list_decisions.json; then
  echo "orderingStrategy field not found in matchmaking decisions payload" >&2
  cat /tmp/week9_list_decisions.json >&2 || true
  exit 1
fi

echo "Week 9 smoke passed."
