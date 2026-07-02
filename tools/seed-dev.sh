#!/usr/bin/env bash
set -euo pipefail

API_BASE="${API_BASE:-http://localhost:8080}"
TMP_DIR="${TMP_DIR:-/tmp}"

assert_code() {
  local expected="$1"
  local actual="$2"
  local label="$3"
  if [[ "$actual" != "$expected" ]]; then
    echo "${label}: expected HTTP ${expected}, got ${actual}" >&2
    exit 1
  fi
}

extract_id() {
  sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' "$1" | tail -n1
}

echo "Seeding Kinetic v3 dev data"
echo "Waiting for API to be ready at ${API_BASE}/healthz"
for i in $(seq 1 30); do
  if curl -s -f "${API_BASE}/healthz" >/dev/null; then
    break
  fi
  sleep 1
  if [[ "$i" -eq 30 ]]; then
    echo "API did not become ready in time." >&2
    exit 1
  fi
done

suffix="$(date +%s)"
league_slug="dev-league-${suffix}"
franchise_slug="dev-franchise-${suffix}"
club_a_slug="dev-club-a-${suffix}"
club_b_slug="dev-club-b-${suffix}"
team_a_slug="dev-team-a-${suffix}"
team_b_slug="dev-team-b-${suffix}"
season_slug="dev-season-${suffix}"
queue_slug="dev-queue-${suffix}"

league_code=$(curl -s -o "${TMP_DIR}/seed_dev_league.json" -w '%{http_code}' \
  -X POST "${API_BASE}/v1/leagues" \
  -H 'content-type: application/json' \
  -d "{\"name\":\"MLE Dev League\",\"slug\":\"${league_slug}\"}")
assert_code 201 "$league_code" "create league"
league_id="$(extract_id "${TMP_DIR}/seed_dev_league.json")"

franchise_code=$(curl -s -o "${TMP_DIR}/seed_dev_franchise.json" -w '%{http_code}' \
  -X POST "${API_BASE}/v1/franchises" \
  -H 'content-type: application/json' \
  -d "{\"leagueId\":${league_id},\"name\":\"Dev Franchise\",\"slug\":\"${franchise_slug}\"}")
assert_code 201 "$franchise_code" "create franchise"
franchise_id="$(extract_id "${TMP_DIR}/seed_dev_franchise.json")"

club_a_code=$(curl -s -o "${TMP_DIR}/seed_dev_club_a.json" -w '%{http_code}' \
  -X POST "${API_BASE}/v1/clubs" \
  -H 'content-type: application/json' \
  -d "{\"franchiseId\":${franchise_id},\"name\":\"Dev Club A\",\"slug\":\"${club_a_slug}\"}")
assert_code 201 "$club_a_code" "create club A"
club_a_id="$(extract_id "${TMP_DIR}/seed_dev_club_a.json")"

club_b_code=$(curl -s -o "${TMP_DIR}/seed_dev_club_b.json" -w '%{http_code}' \
  -X POST "${API_BASE}/v1/clubs" \
  -H 'content-type: application/json' \
  -d "{\"franchiseId\":${franchise_id},\"name\":\"Dev Club B\",\"slug\":\"${club_b_slug}\"}")
assert_code 201 "$club_b_code" "create club B"
club_b_id="$(extract_id "${TMP_DIR}/seed_dev_club_b.json")"

team_a_code=$(curl -s -o "${TMP_DIR}/seed_dev_team_a.json" -w '%{http_code}' \
  -X POST "${API_BASE}/v1/teams" \
  -H 'content-type: application/json' \
  -d "{\"clubId\":${club_a_id},\"name\":\"Dev Team A\",\"slug\":\"${team_a_slug}\"}")
assert_code 201 "$team_a_code" "create team A"
team_a_id="$(extract_id "${TMP_DIR}/seed_dev_team_a.json")"

team_b_code=$(curl -s -o "${TMP_DIR}/seed_dev_team_b.json" -w '%{http_code}' \
  -X POST "${API_BASE}/v1/teams" \
  -H 'content-type: application/json' \
  -d "{\"clubId\":${club_b_id},\"name\":\"Dev Team B\",\"slug\":\"${team_b_slug}\"}")
assert_code 201 "$team_b_code" "create team B"
team_b_id="$(extract_id "${TMP_DIR}/seed_dev_team_b.json")"

season_code=$(curl -s -o "${TMP_DIR}/seed_dev_season.json" -w '%{http_code}' \
  -X POST "${API_BASE}/v1/seasons" \
  -H 'content-type: application/json' \
  -d "{\"name\":\"Dev Season\",\"slug\":\"${season_slug}\"}")
assert_code 201 "$season_code" "create season"
season_id="$(extract_id "${TMP_DIR}/seed_dev_season.json")"

schedule_group_code=$(curl -s -o "${TMP_DIR}/seed_dev_schedule_group.json" -w '%{http_code}' \
  -X POST "${API_BASE}/v1/schedule-groups" \
  -H 'content-type: application/json' \
  -d "{\"seasonId\":${season_id},\"name\":\"Week 1\",\"sequence\":1}")
assert_code 201 "$schedule_group_code" "create schedule group"
schedule_group_id="$(extract_id "${TMP_DIR}/seed_dev_schedule_group.json")"

fixture_code=$(curl -s -o "${TMP_DIR}/seed_dev_fixture.json" -w '%{http_code}' \
  -X POST "${API_BASE}/v1/fixtures" \
  -H 'content-type: application/json' \
  -d "{\"scheduleGroupId\":${schedule_group_id},\"homeClubId\":${club_a_id},\"awayClubId\":${club_b_id}}")
assert_code 201 "$fixture_code" "create fixture"
fixture_id="$(extract_id "${TMP_DIR}/seed_dev_fixture.json")"

match_code=$(curl -s -o "${TMP_DIR}/seed_dev_match.json" -w '%{http_code}' \
  -X POST "${API_BASE}/v1/matches" \
  -H 'content-type: application/json' \
  -d "{\"fixtureId\":${fixture_id},\"homeTeamId\":${team_a_id},\"awayTeamId\":${team_b_id},\"state\":\"planned\"}")
assert_code 201 "$match_code" "create match"
match_id="$(extract_id "${TMP_DIR}/seed_dev_match.json")"

queue_code=$(curl -s -o "${TMP_DIR}/seed_dev_queue.json" -w '%{http_code}' \
  -X POST "${API_BASE}/v1/queues" \
  -H 'content-type: application/json' \
  -d "{\"name\":\"Dev Scrim Queue\",\"slug\":\"${queue_slug}\"}")
assert_code 201 "$queue_code" "create queue"
queue_id="$(extract_id "${TMP_DIR}/seed_dev_queue.json")"

enqueue_a_code=$(curl -s -o "${TMP_DIR}/seed_dev_enqueue_a.json" -w '%{http_code}' \
  -X POST "${API_BASE}/v1/queue-entries" \
  -H 'content-type: application/json' \
  -d "{\"queueId\":${queue_id},\"teamId\":${team_a_id}}")
assert_code 201 "$enqueue_a_code" "enqueue team A"

enqueue_b_code=$(curl -s -o "${TMP_DIR}/seed_dev_enqueue_b.json" -w '%{http_code}' \
  -X POST "${API_BASE}/v1/queue-entries" \
  -H 'content-type: application/json' \
  -d "{\"queueId\":${queue_id},\"teamId\":${team_b_id}}")
assert_code 201 "$enqueue_b_code" "enqueue team B"

process_code=$(curl -s -o "${TMP_DIR}/seed_dev_process.json" -w '%{http_code}' \
  -X POST "${API_BASE}/v1/scrim-promotions/process" \
  -H 'content-type: application/json' \
  -d "{\"queueId\":${queue_id}}")
assert_code 200 "$process_code" "process queue promotions"

list_scrims_code=$(curl -s -o "${TMP_DIR}/seed_dev_scrims.json" -w '%{http_code}' \
  "${API_BASE}/v1/scrims")
assert_code 200 "$list_scrims_code" "list scrims"
scrim_id="$(extract_id "${TMP_DIR}/seed_dev_scrims.json")"

submission_code=$(curl -s -o "${TMP_DIR}/seed_dev_submission.json" -w '%{http_code}' \
  -X POST "${API_BASE}/v1/result-submissions" \
  -H 'content-type: application/json' \
  -d "{\"contextType\":\"scrim\",\"contextId\":${scrim_id},\"submittedByTeamId\":${team_a_id},\"winningTeamId\":${team_a_id},\"losingTeamId\":${team_b_id},\"payloadJson\":{\"score\":\"3-1\",\"source\":\"seed-dev\"}}")
assert_code 201 "$submission_code" "create result submission"

exception_code=$(curl -s -o "${TMP_DIR}/seed_dev_exception.json" -w '%{http_code}' \
  -X POST "${API_BASE}/v1/exceptions/report" \
  -H 'content-type: application/json' \
  -d "{\"category\":\"scheduling_conflict\",\"contextType\":\"match\",\"contextId\":${match_id},\"reasonCode\":\"time_unavailable\",\"severity\":3,\"suggestedAction\":\"propose_reschedule\",\"detailsJson\":{\"source\":\"seed-dev\"}}")
assert_code 201 "$exception_code" "report exception"

cat <<EOF
Seeding complete.
League: ${league_slug}
Franchise: ${franchise_slug}
Queue: ${queue_slug}
Scrim ID: ${scrim_id}
EOF
