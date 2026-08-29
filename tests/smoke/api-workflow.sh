#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
PG_NAME="kinetic-v3-pg-kinetic_api_smoke-smoke"
PG_PORT="${PG_PORT:-55432}"
DB_URL=""
API_PID=""
ADMIN_COOKIE="/tmp/kinetic_v3_kinetic_api_smoke_admin.cookies"
OPPOSING_ADMIN_COOKIE="/tmp/kinetic_v3_kinetic_api_smoke_opposing_admin.cookies"

cleanup() {
  if [[ -n "$API_PID" ]]; then
    kill "$API_PID" >/dev/null 2>&1 || true
    wait "$API_PID" 2>/dev/null || true
  fi
  docker rm -f "$PG_NAME" >/dev/null 2>&1 || true
  rm -f "$ADMIN_COOKIE" "$OPPOSING_ADMIN_COOKIE"
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
    -d postgres:16 >/dev/null 2>/tmp/kinetic_v3_kinetic_api_smoke_docker_err.log; then
    PG_PORT="${candidate_port}"
    DB_URL="postgres://postgres:postgres@localhost:${PG_PORT}/kinetic?sslmode=disable"
    break
  fi
done

if [[ -z "${DB_URL}" ]]; then
  echo "Unable to start Postgres container on ports ${PG_PORT}-${PG_PORT}+4" >&2
  cat /tmp/kinetic_v3_kinetic_api_smoke_docker_err.log >&2 || true
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
DATABASE_URL="$DB_URL" MIGRATIONS_DIR="./migrations" go run ./cmd/migrate >/tmp/kinetic_v3_kinetic_api_smoke_migrate.log

echo "Starting API in DB-required mode"
DATABASE_URL="$DB_URL" REQUIRE_DATABASE=true AUTH_LOCAL_LOGIN_ENABLED=true go run ./cmd/api >/tmp/kinetic_v3_kinetic_api_smoke_api.log 2>&1 &
API_PID=$!

for i in $(seq 1 40); do
  if curl -fsS http://localhost:8080/readyz >/dev/null 2>&1; then
    break
  fi
  sleep 1
  if [[ "$i" -eq 40 ]]; then
    echo "API did not become ready in time" >&2
    cat /tmp/kinetic_v3_kinetic_api_smoke_api.log >&2 || true
    exit 1
  fi
done

rm -f "$ADMIN_COOKIE" "$OPPOSING_ADMIN_COOKIE"
auth_code=$(curl -s -o /tmp/kinetic_api_smoke_auth_admin.json -w '%{http_code}' \
  -c "$ADMIN_COOKIE" \
  "http://localhost:8080/v1/auth/callback?subject=kinetic_api_smoke-admin&displayName=Kinetic%20Admin&roles=admin&redirect=http://localhost:5173/app/admin")
if [[ "$auth_code" != "302" ]]; then
  echo "create admin smoke session: expected HTTP 302, got ${auth_code}" >&2
  cat /tmp/kinetic_api_smoke_auth_admin.json >&2 || true
  exit 1
fi
opposing_auth_code=$(curl -s -o /tmp/kinetic_api_smoke_auth_opposing_admin.json -w '%{http_code}' \
  -c "$OPPOSING_ADMIN_COOKIE" \
  "http://localhost:8080/v1/auth/callback?subject=kinetic_api_smoke-opposing-admin&displayName=Kinetic%20Opposing%20Admin&roles=admin&redirect=http://localhost:5173/app/admin")
if [[ "$opposing_auth_code" != "302" ]]; then
  echo "create opposing admin smoke session: expected HTTP 302, got ${opposing_auth_code}" >&2
  cat /tmp/kinetic_api_smoke_auth_opposing_admin.json >&2 || true
  exit 1
fi

suffix="$(date +%s)"
league_slug="kinetic_api_smoke-league-${suffix}"
franchise_slug="kinetic_api_smoke-franchise-${suffix}"
club_a_slug="kinetic_api_smoke-club-a-${suffix}"
club_b_slug="kinetic_api_smoke-club-b-${suffix}"
team_a_slug="kinetic_api_smoke-team-a-${suffix}"
team_b_slug="kinetic_api_smoke-team-b-${suffix}"
queue_slug="kinetic_api_smoke-queue-${suffix}"

assert_code() {
  local expected="$1"
  local actual="$2"
  local label="$3"
  if [[ "$actual" != "$expected" ]]; then
    echo "${label}: expected HTTP ${expected}, got ${actual}" >&2
    exit 1
  fi
}

league_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_create_league.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/leagues \
  -H 'content-type: application/json' \
  -d "{\"name\":\"Kinetic League ${suffix}\",\"slug\":\"${league_slug}\"}")
assert_code 201 "$league_code" "create league"
league_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/kinetic_api_smoke_create_league.json)

franchise_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_create_franchise.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/franchises \
  -H 'content-type: application/json' \
  -d "{\"leagueId\":${league_id},\"name\":\"Kinetic Franchise ${suffix}\",\"slug\":\"${franchise_slug}\"}")
assert_code 201 "$franchise_code" "create franchise"
franchise_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/kinetic_api_smoke_create_franchise.json)

club_a_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_create_club_a.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/clubs \
  -H 'content-type: application/json' \
  -d "{\"franchiseId\":${franchise_id},\"name\":\"Kinetic Club A ${suffix}\",\"slug\":\"${club_a_slug}\"}")
assert_code 201 "$club_a_code" "create club A"
club_a_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/kinetic_api_smoke_create_club_a.json)

club_b_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_create_club_b.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/clubs \
  -H 'content-type: application/json' \
  -d "{\"franchiseId\":${franchise_id},\"name\":\"Kinetic Club B ${suffix}\",\"slug\":\"${club_b_slug}\"}")
assert_code 201 "$club_b_code" "create club B"
club_b_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/kinetic_api_smoke_create_club_b.json)

team_a_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_create_team_a.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/teams \
  -H 'content-type: application/json' \
  -d "{\"clubId\":${club_a_id},\"name\":\"Kinetic Team A ${suffix}\",\"slug\":\"${team_a_slug}\"}")
assert_code 201 "$team_a_code" "create team A"
team_a_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/kinetic_api_smoke_create_team_a.json)

team_b_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_create_team_b.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/teams \
  -H 'content-type: application/json' \
  -d "{\"clubId\":${club_b_id},\"name\":\"Kinetic Team B ${suffix}\",\"slug\":\"${team_b_slug}\"}")
assert_code 201 "$team_b_code" "create team B"
team_b_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/kinetic_api_smoke_create_team_b.json)

season_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_create_season.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/seasons \
  -H 'content-type: application/json' \
  -d "{\"name\":\"Kinetic Season ${suffix}\",\"slug\":\"kinetic_api_smoke-season-${suffix}\"}")
assert_code 201 "$season_code" "create season"
season_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/kinetic_api_smoke_create_season.json)

schedule_group_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_create_schedule_group.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/schedule-groups \
  -H 'content-type: application/json' \
  -d "{\"seasonId\":${season_id},\"name\":\"Week 1\",\"sequence\":1}")
assert_code 201 "$schedule_group_code" "create schedule group"
schedule_group_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/kinetic_api_smoke_create_schedule_group.json)

fixture_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_create_fixture.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/fixtures \
  -H 'content-type: application/json' \
  -d "{\"scheduleGroupId\":${schedule_group_id},\"homeClubId\":${club_a_id},\"awayClubId\":${club_b_id}}")
assert_code 201 "$fixture_code" "create fixture"
fixture_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/kinetic_api_smoke_create_fixture.json)

match_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_create_match.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/matches \
  -H 'content-type: application/json' \
  -d "{\"fixtureId\":${fixture_id},\"homeTeamId\":${team_a_id},\"awayTeamId\":${team_b_id},\"state\":\"planned\"}")
assert_code 201 "$match_code" "create match"
match_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/kinetic_api_smoke_create_match.json)

queue_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_create_queue.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/queues \
  -H 'content-type: application/json' \
  -d "{\"name\":\"Kinetic Queue ${suffix}\",\"slug\":\"${queue_slug}\"}")
assert_code 201 "$queue_code" "create queue"
queue_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/kinetic_api_smoke_create_queue.json)

enqueue_a_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_enqueue_a.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/queue-entries \
  -H 'content-type: application/json' \
  -d "{\"queueId\":${queue_id},\"teamId\":${team_a_id}}")
assert_code 201 "$enqueue_a_code" "enqueue team A"

enqueue_b_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_enqueue_b.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/queue-entries \
  -H 'content-type: application/json' \
  -d "{\"queueId\":${queue_id},\"teamId\":${team_b_id}}")
assert_code 201 "$enqueue_b_code" "enqueue team B"

process_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_process_ok.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/scrim-promotions/process \
  -H 'content-type: application/json' \
  -d "{\"queueId\":${queue_id}}")
assert_code 200 "$process_code" "process queue promotions"

list_scrims_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_list_scrims.json -w '%{http_code}' http://localhost:8080/v1/scrims)
assert_code 200 "$list_scrims_code" "list scrims"
scrim_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/kinetic_api_smoke_list_scrims.json | tail -n1)

in_progress_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_scrim_in_progress.json -w '%{http_code}' \
  -X PATCH http://localhost:8080/v1/scrims \
  -H 'content-type: application/json' \
  -d "{\"scrimId\":${scrim_id},\"state\":\"in_progress\"}")
assert_code 200 "$in_progress_code" "transition scrim to in_progress"

closed_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_scrim_closed.json -w '%{http_code}' \
  -X PATCH http://localhost:8080/v1/scrims \
  -H 'content-type: application/json' \
  -d "{\"scrimId\":${scrim_id},\"state\":\"closed\"}")
assert_code 200 "$closed_code" "transition scrim to closed"

invalid_transition_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_scrim_invalid_transition.json -w '%{http_code}' \
  -X PATCH http://localhost:8080/v1/scrims \
  -H 'content-type: application/json' \
  -d "{\"scrimId\":${scrim_id},\"state\":\"in_progress\"}")
assert_code 409 "$invalid_transition_code" "reject closed -> in_progress"

create_submission_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_create_submission_1.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/result-submissions \
  -H 'content-type: application/json' \
  -d "{\"contextType\":\"scrim\",\"contextId\":${scrim_id},\"submittedByTeamId\":${team_a_id},\"winningTeamId\":${team_a_id},\"losingTeamId\":${team_b_id},\"payloadJson\":{\"score\":\"3-1\",\"source\":\"kinetic_api_smoke-smoke\"}}")
assert_code 201 "$create_submission_code" "create submission 1"
submission_1_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/kinetic_api_smoke_create_submission_1.json)

ratify_1_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_ratify_submission_1_team_a.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/result-submission-ratifications \
  -H 'content-type: application/json' \
  -d "{\"submissionId\":${submission_1_id},\"teamId\":${team_a_id}}")
assert_code 409 "$ratify_1_code" "reject duplicate submitter attestation"

ratify_2_code=$(curl -s -b "$OPPOSING_ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_ratify_submission_1_team_b.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/result-submission-ratifications \
  -H 'content-type: application/json' \
  -d "{\"submissionId\":${submission_1_id},\"teamId\":${team_b_id}}")
assert_code 200 "$ratify_2_code" "ratify submission 1 by team B"
if ! grep -q '"state":"ratified"' /tmp/kinetic_api_smoke_ratify_submission_1_team_b.json; then
  echo "submission 1 did not transition to ratified" >&2
  cat /tmp/kinetic_api_smoke_ratify_submission_1_team_b.json >&2 || true
  exit 1
fi

create_submission_2_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_create_submission_2.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/result-submissions \
  -H 'content-type: application/json' \
  -d "{\"contextType\":\"scrim\",\"contextId\":${scrim_id},\"submittedByTeamId\":${team_b_id},\"winningTeamId\":${team_b_id},\"losingTeamId\":${team_a_id},\"payloadJson\":{\"score\":\"1-2\",\"source\":\"kinetic_api_smoke-smoke\"}}")
assert_code 201 "$create_submission_2_code" "create submission 2"
submission_2_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/kinetic_api_smoke_create_submission_2.json)

duplicate_pending_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_create_submission_duplicate_pending.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/result-submissions \
  -H 'content-type: application/json' \
  -d "{\"contextType\":\"scrim\",\"contextId\":${scrim_id},\"submittedByTeamId\":${team_a_id},\"winningTeamId\":${team_a_id},\"losingTeamId\":${team_b_id},\"payloadJson\":{\"score\":\"1-0\",\"source\":\"kinetic_api_smoke-smoke\"}}")
assert_code 409 "$duplicate_pending_code" "reject duplicate pending submission"

reject_submission_2_code=$(curl -s -b "$OPPOSING_ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_reject_submission_2.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/result-submission-rejections \
  -H 'content-type: application/json' \
  -d "{\"submissionId\":${submission_2_id},\"teamId\":${team_a_id},\"reason\":\"replay mismatch\"}")
assert_code 200 "$reject_submission_2_code" "reject submission 2"
if ! grep -q '"state":"rejected"' /tmp/kinetic_api_smoke_reject_submission_2.json; then
  echo "submission 2 did not transition to rejected" >&2
  cat /tmp/kinetic_api_smoke_reject_submission_2.json >&2 || true
  exit 1
fi

list_submissions_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_list_result_submissions.json -w '%{http_code}' \
  http://localhost:8080/v1/result-submissions)
assert_code 200 "$list_submissions_code" "list result submissions"
if ! grep -q '"contextType":"scrim"' /tmp/kinetic_api_smoke_list_result_submissions.json; then
  echo "result submissions payload missing expected contextType" >&2
  cat /tmp/kinetic_api_smoke_list_result_submissions.json >&2 || true
  exit 1
fi

ingest_replay_1_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_ingest_replay_1.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/replay-evidence \
  -H 'content-type: application/json' \
  -d "{\"contextType\":\"scrim\",\"contextId\":${scrim_id},\"submittedByTeamId\":${team_a_id},\"replayBody\":\"kinetic_api_smoke-smoke-replay-body\",\"parserName\":\"rl-replay\",\"parserVersion\":\"v0.1.0\",\"parserConfigDigest\":\"cfg-kinetic_api_smoke-smoke\",\"resultSubmissionId\":${submission_1_id},\"parseOutputJson\":{\"goals\":4}}")
assert_code 201 "$ingest_replay_1_code" "ingest replay evidence first attempt"
if ! grep -q '"duplicate":false' /tmp/kinetic_api_smoke_ingest_replay_1.json; then
  echo "expected duplicate=false for first replay ingest" >&2
  cat /tmp/kinetic_api_smoke_ingest_replay_1.json >&2 || true
  exit 1
fi

ingest_replay_2_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_ingest_replay_2.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/replay-evidence \
  -H 'content-type: application/json' \
  -d "{\"contextType\":\"scrim\",\"contextId\":${scrim_id},\"submittedByTeamId\":${team_a_id},\"replayBody\":\"kinetic_api_smoke-smoke-replay-body\",\"parserName\":\"rl-replay\",\"parserVersion\":\"v0.1.0\",\"parserConfigDigest\":\"cfg-kinetic_api_smoke-smoke\",\"resultSubmissionId\":${submission_1_id},\"parseOutputJson\":{\"goals\":4}}")
assert_code 200 "$ingest_replay_2_code" "ingest replay evidence duplicate attempt"
if ! grep -q '"duplicate":true' /tmp/kinetic_api_smoke_ingest_replay_2.json; then
  echo "expected duplicate=true for second replay ingest" >&2
  cat /tmp/kinetic_api_smoke_ingest_replay_2.json >&2 || true
  exit 1
fi

list_replay_evidence_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_list_replay_evidence.json -w '%{http_code}' \
  http://localhost:8080/v1/replay-evidence)
assert_code 200 "$list_replay_evidence_code" "list replay evidence"
if ! grep -q '"replaySha256"' /tmp/kinetic_api_smoke_list_replay_evidence.json; then
  echo "replay evidence payload missing replaySha256" >&2
  cat /tmp/kinetic_api_smoke_list_replay_evidence.json >&2 || true
  exit 1
fi

list_parse_runs_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_list_replay_parse_runs.json -w '%{http_code}' \
  http://localhost:8080/v1/replay-parse-runs)
assert_code 200 "$list_parse_runs_code" "list replay parse runs"
if ! grep -q '"parserVersion"' /tmp/kinetic_api_smoke_list_replay_parse_runs.json; then
  echo "replay parse runs payload missing parserVersion" >&2
  cat /tmp/kinetic_api_smoke_list_replay_parse_runs.json >&2 || true
  exit 1
fi

list_replay_links_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_list_result_submission_replay_links.json -w '%{http_code}' \
  http://localhost:8080/v1/result-submission-replay-links)
assert_code 200 "$list_replay_links_code" "list result submission replay links"
if ! grep -q '"resultSubmissionId"' /tmp/kinetic_api_smoke_list_result_submission_replay_links.json; then
  echo "result submission replay links payload missing resultSubmissionId" >&2
  cat /tmp/kinetic_api_smoke_list_result_submission_replay_links.json >&2 || true
  exit 1
fi

report_exception_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_report_exception.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/exceptions/report \
  -H 'content-type: application/json' \
  -d "{\"category\":\"scheduling_conflict\",\"contextType\":\"match\",\"contextId\":${match_id},\"reasonCode\":\"time_unavailable\",\"severity\":3,\"suggestedAction\":\"propose_reschedule\",\"detailsJson\":{\"source\":\"kinetic_api_smoke-smoke\"}}")
assert_code 201 "$report_exception_code" "report exception"
exception_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/kinetic_api_smoke_report_exception.json)

triage_exception_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_triage_exception.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/operator-inbox/triage \
  -H 'content-type: application/json' \
  -d "{\"ticketId\":${exception_id},\"actor\":\"ops-user\",\"reasonCode\":\"captain_conflict\",\"severity\":2,\"suggestedAction\":\"offer_slots\",\"minutesSpent\":5}")
assert_code 200 "$triage_exception_code" "triage exception"

resolve_exception_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_resolve_exception.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/operator-inbox/resolve \
  -H 'content-type: application/json' \
  -d "{\"ticketId\":${exception_id},\"actor\":\"ops-user\",\"resolutionCode\":\"rescheduled\",\"notes\":\"captains agreed\",\"automated\":false,\"minutesSpent\":10}")
assert_code 200 "$resolve_exception_code" "resolve exception"

automation_schedule_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_exception_automation_schedule.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/exception-automations/scheduling \
  -H 'content-type: application/json' \
  -d "{\"matchId\":${match_id},\"conflictCode\":\"captain_conflict\",\"homeConfirmed\":false,\"awayConfirmed\":false,\"actor\":\"ops-bot\"}")
assert_code 200 "$automation_schedule_code" "evaluate scheduling exception"

automation_noshow_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_exception_automation_no_show.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/exception-automations/no-show \
  -H 'content-type: application/json' \
  -d "{\"matchId\":${match_id},\"homeCheckedIn\":true,\"awayCheckedIn\":false,\"graceMinutes\":20,\"actor\":\"ops-bot\"}")
assert_code 200 "$automation_noshow_code" "evaluate no-show exception"

automation_replay_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_exception_automation_replay.json -w '%{http_code}' \
  -X POST http://localhost:8080/v1/exception-automations/replay-dispute \
  -H 'content-type: application/json' \
  -d "{\"resultSubmissionId\":${submission_1_id},\"parseStatus\":\"parsed\",\"identityStatus\":\"resolved\",\"disputeRaised\":false,\"actor\":\"ops-bot\"}")
assert_code 200 "$automation_replay_code" "evaluate replay dispute exception"

list_inbox_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_list_operator_inbox.json -w '%{http_code}' http://localhost:8080/v1/operator-inbox)
assert_code 200 "$list_inbox_code" "list operator inbox"
if ! grep -q '"category"' /tmp/kinetic_api_smoke_list_operator_inbox.json; then
  echo "operator inbox payload missing category" >&2
  cat /tmp/kinetic_api_smoke_list_operator_inbox.json >&2 || true
  exit 1
fi

list_exception_actions_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_list_exception_actions.json -w '%{http_code}' http://localhost:8080/v1/exception-actions)
assert_code 200 "$list_exception_actions_code" "list exception actions"
if ! grep -q '"actionType"' /tmp/kinetic_api_smoke_list_exception_actions.json; then
  echo "exception actions payload missing actionType" >&2
  cat /tmp/kinetic_api_smoke_list_exception_actions.json >&2 || true
  exit 1
fi

exception_metrics_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_exception_metrics.json -w '%{http_code}' http://localhost:8080/v1/exception-metrics)
assert_code 200 "$exception_metrics_code" "get exception metrics"
if ! grep -q '"adminHoursPerWeek"' /tmp/kinetic_api_smoke_exception_metrics.json; then
  echo "exception metrics payload missing adminHoursPerWeek" >&2
  cat /tmp/kinetic_api_smoke_exception_metrics.json >&2 || true
  exit 1
fi

list_decisions_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_list_decisions.json -w '%{http_code}' http://localhost:8080/v1/matchmaking-decisions)
assert_code 200 "$list_decisions_code" "list matchmaking decisions"
if ! grep -q '"orderingStrategy"' /tmp/kinetic_api_smoke_list_decisions.json; then
  echo "orderingStrategy field not found in matchmaking decisions payload" >&2
  cat /tmp/kinetic_api_smoke_list_decisions.json >&2 || true
  exit 1
fi

list_runs_code=$(curl -s -b "$ADMIN_COOKIE" -o /tmp/kinetic_api_smoke_list_processing_runs.json -w '%{http_code}' http://localhost:8080/v1/promotion-processing-runs)
assert_code 200 "$list_runs_code" "list promotion processing runs"
if ! grep -q '"processedQueues"' /tmp/kinetic_api_smoke_list_processing_runs.json; then
  echo "processedQueues field not found in promotion processing runs payload" >&2
  cat /tmp/kinetic_api_smoke_list_processing_runs.json >&2 || true
  exit 1
fi

echo "API workflow smoke test passed."
