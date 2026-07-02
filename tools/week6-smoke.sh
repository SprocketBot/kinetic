#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PG_NAME="kinetic-v3-pg-week6-smoke"
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
    -d postgres:16 >/dev/null 2>/tmp/kinetic_v3_week6_docker_err.log; then
    PG_PORT="${candidate_port}"
    DB_URL="postgres://postgres:postgres@localhost:${PG_PORT}/kinetic?sslmode=disable"
    break
  fi
done

if [[ -z "${DB_URL}" ]]; then
  echo "Unable to start Postgres container on ports ${PG_PORT}-${PG_PORT}+4" >&2
  cat /tmp/kinetic_v3_week6_docker_err.log >&2 || true
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
DATABASE_URL="$DB_URL" MIGRATIONS_DIR="./migrations" go run ./cmd/migrate >/tmp/kinetic_v3_week6_migrate.log

echo "Starting API in DB-required mode"
DATABASE_URL="$DB_URL" REQUIRE_DATABASE=true go run ./cmd/api >/tmp/kinetic_v3_week6_api.log 2>&1 &
API_PID=$!
sleep 1

suffix="$(date +%s)"
league_slug="week6-league-${suffix}"
franchise_slug="week6-franchise-${suffix}"
club_slug="week6-club-${suffix}"
team_slug="week6-team-${suffix}"
team_two_slug="week6-team-two-${suffix}"
queue_slug="week6-queue-${suffix}"

assert_code() {
  local expected="$1"
  local actual="$2"
  local label="$3"
  if [[ "$actual" != "$expected" ]]; then
    echo "${label}: expected HTTP ${expected}, got ${actual}" >&2
    exit 1
  fi
}

league_code=$(curl -s -o /tmp/week6_create_league.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/leagues \
  -H 'content-type: application/json' \
  -d "{\"name\":\"Week6 League ${suffix}\",\"slug\":\"${league_slug}\"}")
assert_code 201 "$league_code" "create league"
league_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week6_create_league.json)

franchise_code=$(curl -s -o /tmp/week6_create_franchise.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/franchises \
  -H 'content-type: application/json' \
  -d "{\"leagueId\":${league_id},\"name\":\"Week6 Franchise ${suffix}\",\"slug\":\"${franchise_slug}\"}")
assert_code 201 "$franchise_code" "create franchise"
franchise_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week6_create_franchise.json)

club_code=$(curl -s -o /tmp/week6_create_club.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/clubs \
  -H 'content-type: application/json' \
  -d "{\"franchiseId\":${franchise_id},\"name\":\"Week6 Club ${suffix}\",\"slug\":\"${club_slug}\"}")
assert_code 201 "$club_code" "create club"
club_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week6_create_club.json)

team_code=$(curl -s -o /tmp/week6_create_team.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/teams \
  -H 'content-type: application/json' \
  -d "{\"clubId\":${club_id},\"name\":\"Week6 Team ${suffix}\",\"slug\":\"${team_slug}\"}")
assert_code 201 "$team_code" "create team"
team_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week6_create_team.json)

team_two_code=$(curl -s -o /tmp/week6_create_team_two.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/teams \
  -H 'content-type: application/json' \
  -d "{\"clubId\":${club_id},\"name\":\"Week6 Team Two ${suffix}\",\"slug\":\"${team_two_slug}\"}")
assert_code 201 "$team_two_code" "create second team"
team_two_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week6_create_team_two.json)

queue_code=$(curl -s -o /tmp/week6_create_queue.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/queues \
  -H 'content-type: application/json' \
  -d "{\"name\":\"Week6 Queue ${suffix}\",\"slug\":\"${queue_slug}\"}")
assert_code 201 "$queue_code" "create queue"
queue_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week6_create_queue.json)

join_one_code=$(curl -s -o /tmp/week6_join_one.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/queue-entries \
  -H 'content-type: application/json' \
  -d "{\"queueId\":${queue_id},\"teamId\":${team_id}}")
assert_code 201 "$join_one_code" "join queue (team one)"

join_two_code=$(curl -s -o /tmp/week6_join_two.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/queue-entries \
  -H 'content-type: application/json' \
  -d "{\"queueId\":${queue_id},\"teamId\":${team_two_id}}")
assert_code 201 "$join_two_code" "join queue (team two)"

dup_join_code=$(curl -s -o /tmp/week6_dup_join.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/queue-entries \
  -H 'content-type: application/json' \
  -d "{\"queueId\":${queue_id},\"teamId\":${team_id}}")
assert_code 409 "$dup_join_code" "duplicate queue join"

bad_fk_join_code=$(curl -s -o /tmp/week6_bad_fk_join.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/queue-entries \
  -H 'content-type: application/json' \
  -d "{\"queueId\":${queue_id},\"teamId\":9999999}")
assert_code 409 "$bad_fk_join_code" "bad queue/team FK"

bad_validation_join_code=$(curl -s -o /tmp/week6_bad_validation_join.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/queue-entries \
  -H 'content-type: application/json' \
  -d '{"queueId":0,"teamId":1}')
assert_code 400 "$bad_validation_join_code" "bad queue join payload"

leave_code=$(curl -s -o /tmp/week6_leave.json -w '%{http_code}' \
  -X DELETE http://localhost:8080/v1/queue-entries \
  -H 'content-type: application/json' \
  -d "{\"queueId\":${queue_id},\"teamId\":${team_id}}")
assert_code 200 "$leave_code" "leave queue"

leave_again_code=$(curl -s -o /tmp/week6_leave_again.json -w '%{http_code}' \
  -X DELETE http://localhost:8080/v1/queue-entries \
  -H 'content-type: application/json' \
  -d "{\"queueId\":${queue_id},\"teamId\":${team_id}}")
assert_code 409 "$leave_again_code" "leave queue twice"

list_queues_code=$(curl -s -o /tmp/week6_list_queues.json -w '%{http_code}' http://localhost:8080/v1/queues)
assert_code 200 "$list_queues_code" "list queues"

list_entries_code=$(curl -s -o /tmp/week6_list_entries.json -w '%{http_code}' http://localhost:8080/v1/queue-entries)
assert_code 200 "$list_entries_code" "list active queue entries"

echo "Week 6 smoke passed."
