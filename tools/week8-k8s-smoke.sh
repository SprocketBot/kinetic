#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
NAMESPACE="${K8S_NAMESPACE:-sprocket-v3}"
PF_PORT="${PF_PORT:-18080}"
API_BASE="http://localhost:${PF_PORT}"
PF_PID=""

cleanup() {
  if [[ -n "${PF_PID}" ]]; then
    kill "${PF_PID}" >/dev/null 2>&1 || true
    wait "${PF_PID}" 2>/dev/null || true
  fi
}
trap cleanup EXIT

assert_code() {
  local expected="$1"
  local actual="$2"
  local label="$3"
  if [[ "$actual" != "$expected" ]]; then
    echo "${label}: expected HTTP ${expected}, got ${actual}" >&2
    exit 1
  fi
}

current_ctx="$(kubectl config current-context 2>/dev/null || true)"
if [[ "$current_ctx" != "minikube" ]]; then
  echo "Current kubectl context is '${current_ctx}'. Expected 'minikube'." >&2
  echo "Run: kubectl config use-context minikube" >&2
  exit 1
fi

if ! kubectl cluster-info >/dev/null 2>&1; then
  echo "Minikube cluster is not reachable." >&2
  echo "Run: minikube start" >&2
  exit 1
fi

cd "$ROOT_DIR"

echo "Deploying latest image to minikube"
./deploy/scripts/apply-local.sh

echo "Starting port-forward on ${PF_PORT}"
kubectl -n "$NAMESPACE" port-forward svc/sprocket-v3-api "${PF_PORT}:8080" >/tmp/sprocket_v3_week8_k8s_pf.log 2>&1 &
PF_PID=$!

for i in $(seq 1 40); do
  code=$(curl -s -o /tmp/week8_k8s_healthz.json -w '%{http_code}' "${API_BASE}/healthz" || true)
  if [[ "$code" == "200" ]]; then
    break
  fi
  sleep 1
  if [[ "$i" -eq 40 ]]; then
    echo "k8s API did not become reachable via port-forward in time" >&2
    cat /tmp/sprocket_v3_week8_k8s_pf.log >&2 || true
    exit 1
  fi
done

suffix="$(date +%s)"
league_slug="week8-k8s-league-${suffix}"
franchise_slug="week8-k8s-franchise-${suffix}"
club_a_slug="week8-k8s-club-a-${suffix}"
club_b_slug="week8-k8s-club-b-${suffix}"
team_a_slug="week8-k8s-team-a-${suffix}"
team_b_slug="week8-k8s-team-b-${suffix}"
queue_slug="week8-k8s-queue-${suffix}"
player_slug="week8-k8s-player-${suffix}"

league_code=$(curl -s -o /tmp/week8_k8s_create_league.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/leagues" \
  -H 'content-type: application/json' \
  -d "{\"name\":\"Week8 K8s League ${suffix}\",\"slug\":\"${league_slug}\"}")
assert_code 201 "$league_code" "create league"
league_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week8_k8s_create_league.json)

franchise_code=$(curl -s -o /tmp/week8_k8s_create_franchise.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/franchises" \
  -H 'content-type: application/json' \
  -d "{\"leagueId\":${league_id},\"name\":\"Week8 K8s Franchise ${suffix}\",\"slug\":\"${franchise_slug}\"}")
assert_code 201 "$franchise_code" "create franchise"
franchise_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week8_k8s_create_franchise.json)

club_a_code=$(curl -s -o /tmp/week8_k8s_create_club_a.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/clubs" \
  -H 'content-type: application/json' \
  -d "{\"franchiseId\":${franchise_id},\"name\":\"Week8 K8s Club A ${suffix}\",\"slug\":\"${club_a_slug}\"}")
assert_code 201 "$club_a_code" "create club A"
club_a_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week8_k8s_create_club_a.json)

club_b_code=$(curl -s -o /tmp/week8_k8s_create_club_b.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/clubs" \
  -H 'content-type: application/json' \
  -d "{\"franchiseId\":${franchise_id},\"name\":\"Week8 K8s Club B ${suffix}\",\"slug\":\"${club_b_slug}\"}")
assert_code 201 "$club_b_code" "create club B"
club_b_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week8_k8s_create_club_b.json)

team_a_code=$(curl -s -o /tmp/week8_k8s_create_team_a.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/teams" \
  -H 'content-type: application/json' \
  -d "{\"clubId\":${club_a_id},\"name\":\"Week8 K8s Team A ${suffix}\",\"slug\":\"${team_a_slug}\"}")
assert_code 201 "$team_a_code" "create team A"
team_a_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week8_k8s_create_team_a.json)

team_b_code=$(curl -s -o /tmp/week8_k8s_create_team_b.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/teams" \
  -H 'content-type: application/json' \
  -d "{\"clubId\":${club_b_id},\"name\":\"Week8 K8s Team B ${suffix}\",\"slug\":\"${team_b_slug}\"}")
assert_code 201 "$team_b_code" "create team B"
team_b_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week8_k8s_create_team_b.json)

queue_code=$(curl -s -o /tmp/week8_k8s_create_queue.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/queues" \
  -H 'content-type: application/json' \
  -d "{\"name\":\"Week8 K8s Queue ${suffix}\",\"slug\":\"${queue_slug}\"}")
assert_code 201 "$queue_code" "create queue"
queue_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week8_k8s_create_queue.json)

enqueue_a_code=$(curl -s -o /tmp/week8_k8s_enqueue_a.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/queue-entries" \
  -H 'content-type: application/json' \
  -d "{\"queueId\":${queue_id},\"teamId\":${team_a_id}}")
assert_code 201 "$enqueue_a_code" "enqueue team A"

enqueue_b_code=$(curl -s -o /tmp/week8_k8s_enqueue_b.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/queue-entries" \
  -H 'content-type: application/json' \
  -d "{\"queueId\":${queue_id},\"teamId\":${team_b_id}}")
assert_code 201 "$enqueue_b_code" "enqueue team B"

patch_stage_code=$(curl -s -o /tmp/week8_k8s_patch_stage_ok.json -w '%{http_code}' \
  -X PATCH "${API_BASE}/v1/queue-entries" \
  -H 'content-type: application/json' \
  -d "{\"queueId\":${queue_id},\"teamId\":${team_a_id},\"stage\":2}")
assert_code 200 "$patch_stage_code" "advance queue stage"

promote_code=$(curl -s -o /tmp/week8_k8s_promote_ok.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/scrim-promotions" \
  -H 'content-type: application/json' \
  -d "{\"queueId\":${queue_id}}")
assert_code 201 "$promote_code" "promote queue to scrim"

promote_again_code=$(curl -s -o /tmp/week8_k8s_promote_conflict.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/scrim-promotions" \
  -H 'content-type: application/json' \
  -d "{\"queueId\":${queue_id}}")
assert_code 409 "$promote_again_code" "insufficient entries conflict"

scrim_code=$(curl -s -o /tmp/week8_k8s_create_scrim.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/scrims" \
  -H 'content-type: application/json' \
  -d "{\"queueId\":${queue_id},\"homeTeamId\":${team_a_id},\"awayTeamId\":${team_b_id},\"state\":\"created\"}")
assert_code 201 "$scrim_code" "create scrim"

player_code=$(curl -s -o /tmp/week8_k8s_create_player.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/players" \
  -H 'content-type: application/json' \
  -d "{\"displayName\":\"Week8 K8s Player ${suffix}\",\"slug\":\"${player_slug}\"}")
assert_code 201 "$player_code" "create player"

ratings_code=$(curl -s -o /tmp/week8_k8s_list_ratings.json -w '%{http_code}' "${API_BASE}/v1/player-ratings")
assert_code 200 "$ratings_code" "list player ratings"

decisions_code=$(curl -s -o /tmp/week8_k8s_list_decisions.json -w '%{http_code}' "${API_BASE}/v1/matchmaking-decisions")
assert_code 200 "$decisions_code" "list matchmaking decisions"
if ! grep -q "\"queueId\":${queue_id}" /tmp/week8_k8s_list_decisions.json; then
  echo "matchmaking decisions did not include expected queueId=${queue_id}" >&2
  cat /tmp/week8_k8s_list_decisions.json >&2 || true
  exit 1
fi

echo "Week 8 k8s smoke passed."
