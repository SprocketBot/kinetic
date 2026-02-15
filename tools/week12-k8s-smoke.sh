#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
NAMESPACE="${K8S_NAMESPACE:-sprocket-v3}"
PF_PORT="${PF_PORT:-18080}"
API_BASE="http://localhost:${PF_PORT}"
PG_APP="sprocket-v3-pg-smoke"
PG_SVC="${PG_APP}"
PF_PID=""
SERVICE_PORT=""

cleanup() {
  if [[ -n "${PF_PID}" ]]; then
    kill "${PF_PID}" >/dev/null 2>&1 || true
    wait "${PF_PID}" 2>/dev/null || true
  fi

  kubectl -n "${NAMESPACE}" delete deployment "${PG_APP}" --ignore-not-found >/dev/null 2>&1 || true
  kubectl -n "${NAMESPACE}" delete service "${PG_SVC}" --ignore-not-found >/dev/null 2>&1 || true

  kubectl -n "${NAMESPACE}" set env deploy/sprocket-v3-api DATABASE_URL- REQUIRE_DATABASE- RUN_MIGRATIONS_ON_START- >/dev/null 2>&1 || true
  kubectl -n "${NAMESPACE}" rollout status deploy/sprocket-v3-api --timeout=180s >/dev/null 2>&1 || true
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

echo "Creating temporary in-cluster Postgres"
cat <<SQL | kubectl -n "${NAMESPACE}" apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${PG_APP}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ${PG_APP}
  template:
    metadata:
      labels:
        app: ${PG_APP}
    spec:
      containers:
        - name: postgres
          image: postgres:16
          env:
            - name: POSTGRES_USER
              value: postgres
            - name: POSTGRES_PASSWORD
              value: postgres
            - name: POSTGRES_DB
              value: sprocket
          ports:
            - containerPort: 5432
              name: pg
          readinessProbe:
            exec:
              command:
                - /bin/sh
                - -c
                - pg_isready -U postgres -d sprocket
            initialDelaySeconds: 5
            periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: ${PG_SVC}
spec:
  selector:
    app: ${PG_APP}
  ports:
    - name: pg
      port: 5432
      targetPort: 5432
SQL

kubectl -n "${NAMESPACE}" rollout status deployment/"${PG_APP}" --timeout=180s

echo "Configuring API deployment for DB-required mode"
kubectl -n "${NAMESPACE}" set env deploy/sprocket-v3-api \
  DATABASE_URL="postgres://postgres:postgres@${PG_SVC}:5432/sprocket?sslmode=disable" \
  REQUIRE_DATABASE=true \
  RUN_MIGRATIONS_ON_START=true
kubectl -n "${NAMESPACE}" rollout status deploy/sprocket-v3-api --timeout=180s

SERVICE_PORT="$(kubectl -n "$NAMESPACE" get svc sprocket-v3-api -o jsonpath='{.spec.ports[0].port}')"
if [[ -z "${SERVICE_PORT}" ]]; then
  echo "Unable to determine service port for ${NAMESPACE}/sprocket-v3-api" >&2
  exit 1
fi

echo "Starting port-forward on ${PF_PORT}"
kubectl -n "$NAMESPACE" port-forward svc/sprocket-v3-api "${PF_PORT}:${SERVICE_PORT}" >/tmp/sprocket_v3_week12_k8s_pf.log 2>&1 &
PF_PID=$!

for i in $(seq 1 40); do
  code="$(curl -s -o /tmp/week12_k8s_healthz.json -w '%{http_code}' "${API_BASE}/healthz" || true)"
  if [[ "$code" == "200" ]]; then
    break
  fi
  sleep 1
  if [[ "$i" -eq 40 ]]; then
    echo "k8s API did not become reachable via port-forward in time" >&2
    cat /tmp/sprocket_v3_week12_k8s_pf.log >&2 || true
    exit 1
  fi
done

suffix="$(date +%s)"
league_slug="week12-k8s-league-${suffix}"
franchise_slug="week12-k8s-franchise-${suffix}"
club_a_slug="week12-k8s-club-a-${suffix}"
club_b_slug="week12-k8s-club-b-${suffix}"
team_a_slug="week12-k8s-team-a-${suffix}"
team_b_slug="week12-k8s-team-b-${suffix}"
queue_slug="week12-k8s-queue-${suffix}"

league_code=$(curl -s -o /tmp/week12_k8s_create_league.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/leagues" \
  -H 'content-type: application/json' \
  -d "{\"name\":\"Week12 K8s League ${suffix}\",\"slug\":\"${league_slug}\"}" || true)
assert_code 201 "$league_code" "create league"
league_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week12_k8s_create_league.json)

franchise_code=$(curl -s -o /tmp/week12_k8s_create_franchise.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/franchises" \
  -H 'content-type: application/json' \
  -d "{\"leagueId\":${league_id},\"name\":\"Week12 K8s Franchise ${suffix}\",\"slug\":\"${franchise_slug}\"}" || true)
assert_code 201 "$franchise_code" "create franchise"
franchise_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week12_k8s_create_franchise.json)

club_a_code=$(curl -s -o /tmp/week12_k8s_create_club_a.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/clubs" \
  -H 'content-type: application/json' \
  -d "{\"franchiseId\":${franchise_id},\"name\":\"Week12 K8s Club A ${suffix}\",\"slug\":\"${club_a_slug}\"}" || true)
assert_code 201 "$club_a_code" "create club A"
club_a_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week12_k8s_create_club_a.json)

club_b_code=$(curl -s -o /tmp/week12_k8s_create_club_b.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/clubs" \
  -H 'content-type: application/json' \
  -d "{\"franchiseId\":${franchise_id},\"name\":\"Week12 K8s Club B ${suffix}\",\"slug\":\"${club_b_slug}\"}" || true)
assert_code 201 "$club_b_code" "create club B"
club_b_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week12_k8s_create_club_b.json)

team_a_code=$(curl -s -o /tmp/week12_k8s_create_team_a.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/teams" \
  -H 'content-type: application/json' \
  -d "{\"clubId\":${club_a_id},\"name\":\"Week12 K8s Team A ${suffix}\",\"slug\":\"${team_a_slug}\"}" || true)
assert_code 201 "$team_a_code" "create team A"
team_a_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week12_k8s_create_team_a.json)

team_b_code=$(curl -s -o /tmp/week12_k8s_create_team_b.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/teams" \
  -H 'content-type: application/json' \
  -d "{\"clubId\":${club_b_id},\"name\":\"Week12 K8s Team B ${suffix}\",\"slug\":\"${team_b_slug}\"}" || true)
assert_code 201 "$team_b_code" "create team B"
team_b_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week12_k8s_create_team_b.json)

queue_code=$(curl -s -o /tmp/week12_k8s_create_queue.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/queues" \
  -H 'content-type: application/json' \
  -d "{\"name\":\"Week12 K8s Queue ${suffix}\",\"slug\":\"${queue_slug}\"}" || true)
assert_code 201 "$queue_code" "create queue"
queue_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week12_k8s_create_queue.json)

enqueue_a_code=$(curl -s -o /tmp/week12_k8s_enqueue_a.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/queue-entries" \
  -H 'content-type: application/json' \
  -d "{\"queueId\":${queue_id},\"teamId\":${team_a_id}}" || true)
assert_code 201 "$enqueue_a_code" "enqueue team A"

enqueue_b_code=$(curl -s -o /tmp/week12_k8s_enqueue_b.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/queue-entries" \
  -H 'content-type: application/json' \
  -d "{\"queueId\":${queue_id},\"teamId\":${team_b_id}}" || true)
assert_code 201 "$enqueue_b_code" "enqueue team B"

process_code=$(curl -s -o /tmp/week12_k8s_process_ok.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/scrim-promotions/process" \
  -H 'content-type: application/json' \
  -d "{\"queueId\":${queue_id}}" || true)
assert_code 200 "$process_code" "process queue promotions"

list_scrims_code=$(curl -s -o /tmp/week12_k8s_list_scrims.json -w '%{http_code}' "${API_BASE}/v1/scrims" || true)
assert_code 200 "$list_scrims_code" "list scrims"
scrim_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week12_k8s_list_scrims.json | tail -n1)

in_progress_code=$(curl -s -o /tmp/week12_k8s_scrim_in_progress.json -w '%{http_code}' \
  -X PATCH "${API_BASE}/v1/scrims" \
  -H 'content-type: application/json' \
  -d "{\"scrimId\":${scrim_id},\"state\":\"in_progress\"}" || true)
assert_code 200 "$in_progress_code" "transition scrim to in_progress"

closed_code=$(curl -s -o /tmp/week12_k8s_scrim_closed.json -w '%{http_code}' \
  -X PATCH "${API_BASE}/v1/scrims" \
  -H 'content-type: application/json' \
  -d "{\"scrimId\":${scrim_id},\"state\":\"closed\"}" || true)
assert_code 200 "$closed_code" "transition scrim to closed"

invalid_transition_code=$(curl -s -o /tmp/week12_k8s_scrim_invalid_transition.json -w '%{http_code}' \
  -X PATCH "${API_BASE}/v1/scrims" \
  -H 'content-type: application/json' \
  -d "{\"scrimId\":${scrim_id},\"state\":\"in_progress\"}" || true)
assert_code 409 "$invalid_transition_code" "reject closed -> in_progress"

create_submission_code=$(curl -s -o /tmp/week12_k8s_create_submission_1.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/result-submissions" \
  -H 'content-type: application/json' \
  -d "{\"contextType\":\"scrim\",\"contextId\":${scrim_id},\"submittedByTeamId\":${team_a_id},\"winningTeamId\":${team_a_id},\"losingTeamId\":${team_b_id},\"payloadJson\":{\"score\":\"3-1\",\"source\":\"week12-k8s-smoke\"}}" || true)
assert_code 201 "$create_submission_code" "create submission 1"
submission_1_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week12_k8s_create_submission_1.json)

ratify_1_code=$(curl -s -o /tmp/week12_k8s_ratify_submission_1_team_a.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/result-submission-ratifications" \
  -H 'content-type: application/json' \
  -d "{\"submissionId\":${submission_1_id},\"teamId\":${team_a_id}}" || true)
assert_code 200 "$ratify_1_code" "ratify submission 1 by team A"

ratify_2_code=$(curl -s -o /tmp/week12_k8s_ratify_submission_1_team_b.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/result-submission-ratifications" \
  -H 'content-type: application/json' \
  -d "{\"submissionId\":${submission_1_id},\"teamId\":${team_b_id}}" || true)
assert_code 200 "$ratify_2_code" "ratify submission 1 by team B"
if ! grep -q '"state":"ratified"' /tmp/week12_k8s_ratify_submission_1_team_b.json; then
  echo "submission 1 did not transition to ratified" >&2
  cat /tmp/week12_k8s_ratify_submission_1_team_b.json >&2 || true
  exit 1
fi

create_submission_2_code=$(curl -s -o /tmp/week12_k8s_create_submission_2.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/result-submissions" \
  -H 'content-type: application/json' \
  -d "{\"contextType\":\"scrim\",\"contextId\":${scrim_id},\"submittedByTeamId\":${team_b_id},\"winningTeamId\":${team_b_id},\"losingTeamId\":${team_a_id},\"payloadJson\":{\"score\":\"2-1\",\"source\":\"week12-k8s-smoke\"}}" || true)
assert_code 201 "$create_submission_2_code" "create submission 2"
submission_2_id=$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' /tmp/week12_k8s_create_submission_2.json)

duplicate_pending_code=$(curl -s -o /tmp/week12_k8s_create_submission_duplicate_pending.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/result-submissions" \
  -H 'content-type: application/json' \
  -d "{\"contextType\":\"scrim\",\"contextId\":${scrim_id},\"submittedByTeamId\":${team_a_id},\"winningTeamId\":${team_a_id},\"losingTeamId\":${team_b_id},\"payloadJson\":{\"score\":\"1-0\",\"source\":\"week12-k8s-smoke\"}}" || true)
assert_code 409 "$duplicate_pending_code" "reject duplicate pending submission"

reject_submission_2_code=$(curl -s -o /tmp/week12_k8s_reject_submission_2.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/result-submission-rejections" \
  -H 'content-type: application/json' \
  -d "{\"submissionId\":${submission_2_id},\"teamId\":${team_a_id},\"reason\":\"replay mismatch\"}" || true)
assert_code 200 "$reject_submission_2_code" "reject submission 2"
if ! grep -q '"state":"rejected"' /tmp/week12_k8s_reject_submission_2.json; then
  echo "submission 2 did not transition to rejected" >&2
  cat /tmp/week12_k8s_reject_submission_2.json >&2 || true
  exit 1
fi

list_submissions_code=$(curl -s -o /tmp/week12_k8s_list_result_submissions.json -w '%{http_code}' "${API_BASE}/v1/result-submissions" || true)
assert_code 200 "$list_submissions_code" "list result submissions"
if ! grep -q '"contextType":"scrim"' /tmp/week12_k8s_list_result_submissions.json; then
  echo "result submissions payload missing expected contextType" >&2
  cat /tmp/week12_k8s_list_result_submissions.json >&2 || true
  exit 1
fi

ingest_replay_1_code=$(curl -s -o /tmp/week12_k8s_ingest_replay_1.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/replay-evidence" \
  -H 'content-type: application/json' \
  -d "{\"contextType\":\"scrim\",\"contextId\":${scrim_id},\"submittedByTeamId\":${team_a_id},\"replayBody\":\"week12-k8s-smoke-replay-body\",\"parserName\":\"sprocket-rl-parser\",\"parserVersion\":\"v0.1.0\",\"parserConfigDigest\":\"cfg-week12-k8s-smoke\",\"resultSubmissionId\":${submission_1_id},\"parseOutputJson\":{\"goals\":4}}" || true)
assert_code 201 "$ingest_replay_1_code" "ingest replay evidence first attempt"
if ! grep -q '"duplicate":false' /tmp/week12_k8s_ingest_replay_1.json; then
  echo "expected duplicate=false for first replay ingest" >&2
  cat /tmp/week12_k8s_ingest_replay_1.json >&2 || true
  exit 1
fi

ingest_replay_2_code=$(curl -s -o /tmp/week12_k8s_ingest_replay_2.json -w '%{http_code}' \
  -X POST "${API_BASE}/v1/replay-evidence" \
  -H 'content-type: application/json' \
  -d "{\"contextType\":\"scrim\",\"contextId\":${scrim_id},\"submittedByTeamId\":${team_a_id},\"replayBody\":\"week12-k8s-smoke-replay-body\",\"parserName\":\"sprocket-rl-parser\",\"parserVersion\":\"v0.1.0\",\"parserConfigDigest\":\"cfg-week12-k8s-smoke\",\"resultSubmissionId\":${submission_1_id},\"parseOutputJson\":{\"goals\":4}}" || true)
assert_code 200 "$ingest_replay_2_code" "ingest replay evidence duplicate attempt"
if ! grep -q '"duplicate":true' /tmp/week12_k8s_ingest_replay_2.json; then
  echo "expected duplicate=true for second replay ingest" >&2
  cat /tmp/week12_k8s_ingest_replay_2.json >&2 || true
  exit 1
fi

list_replay_evidence_code=$(curl -s -o /tmp/week12_k8s_list_replay_evidence.json -w '%{http_code}' "${API_BASE}/v1/replay-evidence" || true)
assert_code 200 "$list_replay_evidence_code" "list replay evidence"
if ! grep -q '"replaySha256"' /tmp/week12_k8s_list_replay_evidence.json; then
  echo "replay evidence payload missing replaySha256" >&2
  cat /tmp/week12_k8s_list_replay_evidence.json >&2 || true
  exit 1
fi

list_parse_runs_code=$(curl -s -o /tmp/week12_k8s_list_replay_parse_runs.json -w '%{http_code}' "${API_BASE}/v1/replay-parse-runs" || true)
assert_code 200 "$list_parse_runs_code" "list replay parse runs"
if ! grep -q '"parserVersion"' /tmp/week12_k8s_list_replay_parse_runs.json; then
  echo "replay parse runs payload missing parserVersion" >&2
  cat /tmp/week12_k8s_list_replay_parse_runs.json >&2 || true
  exit 1
fi

list_replay_links_code=$(curl -s -o /tmp/week12_k8s_list_result_submission_replay_links.json -w '%{http_code}' "${API_BASE}/v1/result-submission-replay-links" || true)
assert_code 200 "$list_replay_links_code" "list result submission replay links"
if ! grep -q '"resultSubmissionId"' /tmp/week12_k8s_list_result_submission_replay_links.json; then
  echo "result submission replay links payload missing resultSubmissionId" >&2
  cat /tmp/week12_k8s_list_result_submission_replay_links.json >&2 || true
  exit 1
fi

decisions_code=$(curl -s -o /tmp/week12_k8s_list_decisions.json -w '%{http_code}' "${API_BASE}/v1/matchmaking-decisions" || true)
assert_code 200 "$decisions_code" "list matchmaking decisions"
if ! grep -q '"orderingStrategy"' /tmp/week12_k8s_list_decisions.json; then
  echo "orderingStrategy field not found in matchmaking decisions payload" >&2
  cat /tmp/week12_k8s_list_decisions.json >&2 || true
  exit 1
fi

list_runs_code=$(curl -s -o /tmp/week12_k8s_list_processing_runs.json -w '%{http_code}' "${API_BASE}/v1/promotion-processing-runs" || true)
assert_code 200 "$list_runs_code" "list promotion processing runs"
if ! grep -q '"processedQueues"' /tmp/week12_k8s_list_processing_runs.json; then
  echo "processedQueues field not found in promotion processing runs payload" >&2
  cat /tmp/week12_k8s_list_processing_runs.json >&2 || true
  exit 1
fi

echo "Week 12 k8s smoke passed."
