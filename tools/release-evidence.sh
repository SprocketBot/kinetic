#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TIMESTAMP="$(date -u +%Y%m%dT%H%M%SZ)"
ARTIFACT_DIR="${RELEASE_EVIDENCE_ARTIFACT_DIR:-${ROOT_DIR}/artifacts/release-validation/local/${TIMESTAMP}}"
API_PORT="${RELEASE_EVIDENCE_API_PORT:-18080}"
WEB_PORT="${RELEASE_EVIDENCE_WEB_PORT:-4173}"
PG_NAME="kinetic-v3-pg-release-evidence"
PG_PORT="${RELEASE_EVIDENCE_PG_PORT:-56432}"
API_BASE_URL="http://127.0.0.1:${API_PORT}"
WEB_BASE_URL="http://127.0.0.1:${WEB_PORT}"
DB_URL=""
API_PID=""

mkdir -p "${ARTIFACT_DIR}/api" "${ARTIFACT_DIR}/browser" "${ARTIFACT_DIR}/logs"
exec > >(tee "${ARTIFACT_DIR}/release-evidence.log") 2>&1

cleanup() {
  if [[ -n "${API_PID}" ]]; then
    kill "${API_PID}" >/dev/null 2>&1 || true
    wait "${API_PID}" 2>/dev/null || true
  fi
  docker rm -f "${PG_NAME}" >/dev/null 2>&1 || true
}
trap cleanup EXIT

assert_code() {
  local expected="$1"
  local actual="$2"
  local label="$3"
  local body_path="$4"
  if [[ "${actual}" != "${expected}" ]]; then
    echo "${label}: expected HTTP ${expected}, got ${actual}" >&2
    cat "${body_path}" >&2 || true
    exit 1
  fi
}

curl_capture() {
  local label="$1"
  local expected="$2"
  shift 2

  local prefix="${ARTIFACT_DIR}/api/${label}"
  local code
  code="$(curl -sS -D "${prefix}.headers" -o "${prefix}.body" -w '%{http_code}' "$@")"
  printf '%s\n' "${code}" > "${prefix}.status"
  assert_code "${expected}" "${code}" "${label}" "${prefix}.body"
}

body_path() {
  printf '%s/api/%s.body' "${ARTIFACT_DIR}" "$1"
}

extract_id() {
  sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' "$1" | head -n1
}

require_json_field() {
  local label="$1"
  local field="$2"
  local file="$3"
  if ! grep -q "\"${field}\"" "${file}"; then
    echo "${label}: response missing JSON field ${field}" >&2
    cat "${file}" >&2 || true
    exit 1
  fi
}

cd "${ROOT_DIR}"

branch="$(git branch --show-current 2>/dev/null || true)"
commit="$(git rev-parse --short HEAD 2>/dev/null || true)"
cat > "${ARTIFACT_DIR}/metadata.json" <<EOF
{
  "environment": "local-release-evidence",
  "timestamp": "${TIMESTAMP}",
  "branch": "${branch}",
  "commit": "${commit}",
  "apiBaseUrl": "${API_BASE_URL}",
  "webBaseUrl": "${WEB_BASE_URL}",
  "artifactDir": "${ARTIFACT_DIR}"
}
EOF

if docker ps -a --format '{{.Names}}' | grep -q "^${PG_NAME}$"; then
  docker rm -f "${PG_NAME}" >/dev/null
fi

for offset in 0 1 2 3 4; do
  candidate_port=$((PG_PORT + offset))
  if docker run --name "${PG_NAME}" \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_PASSWORD=postgres \
    -e POSTGRES_DB=kinetic \
    -p "${candidate_port}:5432" \
    -d postgres:16 >/dev/null 2>"${ARTIFACT_DIR}/logs/postgres-start.err"; then
    PG_PORT="${candidate_port}"
    DB_URL="postgres://postgres:postgres@localhost:${PG_PORT}/kinetic?sslmode=disable"
    break
  fi
done

if [[ -z "${DB_URL}" ]]; then
  echo "Unable to start Postgres container on requested port range" >&2
  cat "${ARTIFACT_DIR}/logs/postgres-start.err" >&2 || true
  exit 1
fi

for i in $(seq 1 40); do
  if docker exec "${PG_NAME}" pg_isready -U postgres -d kinetic >/dev/null 2>&1; then
    break
  fi
  sleep 1
  if [[ "${i}" -eq 40 ]]; then
    echo "Postgres did not become ready in time" >&2
    exit 1
  fi
done

echo "[release-evidence] running migrations"
DATABASE_URL="${DB_URL}" MIGRATIONS_DIR="./migrations" go run ./cmd/migrate >"${ARTIFACT_DIR}/logs/migrate.log" 2>&1

echo "[release-evidence] starting API"
DEPLOYMENT_ENV=local \
PORT="${API_PORT}" \
DATABASE_URL="${DB_URL}" \
REQUIRE_DATABASE=true \
AUTH_SESSION_SECRET="release-evidence-session-secret" \
AUTH_LOCAL_LOGIN_ENABLED=true \
WEB_BASE_URL="${WEB_BASE_URL}" \
CORS_ALLOWED_ORIGINS="${WEB_BASE_URL}" \
go run ./cmd/api >"${ARTIFACT_DIR}/logs/api.log" 2>&1 &
API_PID=$!

for i in $(seq 1 40); do
  if curl -fsS "${API_BASE_URL}/readyz" >"${ARTIFACT_DIR}/api/readyz.body" 2>"${ARTIFACT_DIR}/api/readyz.err"; then
    printf '200\n' >"${ARTIFACT_DIR}/api/readyz.status"
    break
  fi
  sleep 1
  if [[ "${i}" -eq 40 ]]; then
    echo "API did not become ready in time" >&2
    cat "${ARTIFACT_DIR}/logs/api.log" >&2 || true
    exit 1
  fi
done

echo "[release-evidence] checking CORS controls"
curl_capture "cors-preflight-allowed" 204 \
  -X OPTIONS "${API_BASE_URL}/v1/session" \
  -H "Origin: ${WEB_BASE_URL}" \
  -H "Access-Control-Request-Method: GET" \
  -H "Access-Control-Request-Headers: content-type"
if ! grep -Fqi "access-control-allow-origin: ${WEB_BASE_URL}" "${ARTIFACT_DIR}/api/cors-preflight-allowed.headers"; then
  echo "allowed CORS preflight did not reflect ${WEB_BASE_URL}" >&2
  cat "${ARTIFACT_DIR}/api/cors-preflight-allowed.headers" >&2
  exit 1
fi
if ! grep -qi "access-control-allow-credentials: true" "${ARTIFACT_DIR}/api/cors-preflight-allowed.headers"; then
  echo "allowed CORS preflight did not allow credentials" >&2
  cat "${ARTIFACT_DIR}/api/cors-preflight-allowed.headers" >&2
  exit 1
fi

curl_capture "cors-preflight-rejected" 403 \
  -X OPTIONS "${API_BASE_URL}/v1/session" \
  -H "Origin: https://not-kinetic.example" \
  -H "Access-Control-Request-Method: GET"
if grep -qi "access-control-allow-origin:" "${ARTIFACT_DIR}/api/cors-preflight-rejected.headers"; then
  echo "rejected CORS preflight unexpectedly included access-control-allow-origin" >&2
  cat "${ARTIFACT_DIR}/api/cors-preflight-rejected.headers" >&2
  exit 1
fi

echo "[release-evidence] checking API session identity and privilege isolation"
player_cookie="${ARTIFACT_DIR}/api/player.cookies"
support_cookie="${ARTIFACT_DIR}/api/support.cookies"
curl_capture "auth-player-callback" 302 \
  -c "${player_cookie}" \
  "${API_BASE_URL}/v1/auth/callback?subject=release-player&displayName=Release%20Player&roles=player&redirect=${WEB_BASE_URL}/app/player"
curl_capture "auth-player-session" 200 \
  -b "${player_cookie}" \
  -H "Origin: ${WEB_BASE_URL}" \
  "${API_BASE_URL}/v1/session"
if ! grep -q '"subject":"release-player"' "$(body_path auth-player-session)"; then
  echo "player session did not preserve release-player subject" >&2
  cat "$(body_path auth-player-session)" >&2
  exit 1
fi
if grep -q '"admin"' "$(body_path auth-player-session)"; then
  echo "player session unexpectedly included admin role" >&2
  cat "$(body_path auth-player-session)" >&2
  exit 1
fi

curl_capture "auth-support-callback" 302 \
  -c "${support_cookie}" \
  "${API_BASE_URL}/v1/auth/callback?subject=release-support&displayName=Release%20Support&roles=league_support&redirect=${WEB_BASE_URL}/app/support"
curl_capture "auth-support-session" 200 \
  -b "${support_cookie}" \
  -H "Origin: ${WEB_BASE_URL}" \
  "${API_BASE_URL}/v1/session"
if ! grep -q '"subject":"release-support"' "$(body_path auth-support-session)"; then
  echo "support session did not preserve release-support subject" >&2
  cat "$(body_path auth-support-session)" >&2
  exit 1
fi

curl_capture "player-result-override-forbidden" 403 \
  -b "${player_cookie}" \
  -X POST "${API_BASE_URL}/v1/result-overrides" \
  -H "content-type: application/json" \
  -d '{"submissionId":1,"actor":"release-evidence","reason":"identity isolation","winningTeamId":1,"losingTeamId":2}'

echo "[release-evidence] creating replay intake fixture"
suffix="$(date +%s)"
curl_capture "create-league" 201 \
  -X POST "${API_BASE_URL}/v1/leagues" \
  -H "content-type: application/json" \
  -d "{\"name\":\"Release Evidence League ${suffix}\",\"slug\":\"release-evidence-league-${suffix}\"}"
league_id="$(extract_id "$(body_path create-league)")"

curl_capture "create-franchise" 201 \
  -X POST "${API_BASE_URL}/v1/franchises" \
  -H "content-type: application/json" \
  -d "{\"leagueId\":${league_id},\"name\":\"Release Evidence Franchise ${suffix}\",\"slug\":\"release-evidence-franchise-${suffix}\"}"
franchise_id="$(extract_id "$(body_path create-franchise)")"

curl_capture "create-club-a" 201 \
  -X POST "${API_BASE_URL}/v1/clubs" \
  -H "content-type: application/json" \
  -d "{\"franchiseId\":${franchise_id},\"name\":\"Release Evidence Club A ${suffix}\",\"slug\":\"release-evidence-club-a-${suffix}\"}"
club_a_id="$(extract_id "$(body_path create-club-a)")"

curl_capture "create-club-b" 201 \
  -X POST "${API_BASE_URL}/v1/clubs" \
  -H "content-type: application/json" \
  -d "{\"franchiseId\":${franchise_id},\"name\":\"Release Evidence Club B ${suffix}\",\"slug\":\"release-evidence-club-b-${suffix}\"}"
club_b_id="$(extract_id "$(body_path create-club-b)")"

curl_capture "create-team-a" 201 \
  -X POST "${API_BASE_URL}/v1/teams" \
  -H "content-type: application/json" \
  -d "{\"clubId\":${club_a_id},\"name\":\"Release Evidence Team A ${suffix}\",\"slug\":\"release-evidence-team-a-${suffix}\"}"
team_a_id="$(extract_id "$(body_path create-team-a)")"

curl_capture "create-team-b" 201 \
  -X POST "${API_BASE_URL}/v1/teams" \
  -H "content-type: application/json" \
  -d "{\"clubId\":${club_b_id},\"name\":\"Release Evidence Team B ${suffix}\",\"slug\":\"release-evidence-team-b-${suffix}\"}"
team_b_id="$(extract_id "$(body_path create-team-b)")"

curl_capture "create-queue" 201 \
  -X POST "${API_BASE_URL}/v1/queues" \
  -H "content-type: application/json" \
  -d "{\"name\":\"Release Evidence Queue ${suffix}\",\"slug\":\"release-evidence-queue-${suffix}\"}"
queue_id="$(extract_id "$(body_path create-queue)")"

curl_capture "create-scrim" 201 \
  -X POST "${API_BASE_URL}/v1/scrims" \
  -H "content-type: application/json" \
  -d "{\"queueId\":${queue_id},\"homeTeamId\":${team_a_id},\"awayTeamId\":${team_b_id},\"state\":\"created\"}"
scrim_id="$(extract_id "$(body_path create-scrim)")"

curl_capture "create-second-scrim" 201 \
  -X POST "${API_BASE_URL}/v1/scrims" \
  -H "content-type: application/json" \
  -d "{\"queueId\":${queue_id},\"homeTeamId\":${team_a_id},\"awayTeamId\":${team_b_id},\"state\":\"created\"}"
second_scrim_id="$(extract_id "$(body_path create-second-scrim)")"

curl_capture "create-submission" 201 \
  -X POST "${API_BASE_URL}/v1/result-submissions" \
  -H "content-type: application/json" \
  -d "{\"contextType\":\"scrim\",\"contextId\":${scrim_id},\"submittedByTeamId\":${team_a_id},\"winningTeamId\":${team_a_id},\"losingTeamId\":${team_b_id},\"payloadJson\":{\"score\":\"3-1\",\"source\":\"release-evidence\"}}"
submission_id="$(extract_id "$(body_path create-submission)")"

replay_body="release-evidence-replay-body-${suffix}"
curl_capture "ingest-replay-first" 201 \
  -X POST "${API_BASE_URL}/v1/replay-evidence" \
  -H "content-type: application/json" \
  -d "{\"contextType\":\"scrim\",\"contextId\":${scrim_id},\"submittedByTeamId\":${team_a_id},\"replayBody\":\"${replay_body}\",\"parserName\":\"kinetic-rl-parser\",\"parserVersion\":\"release-evidence\",\"parserConfigDigest\":\"default\",\"resultSubmissionId\":${submission_id},\"parseOutputJson\":{\"goals\":4}}"
if ! grep -q '"duplicate":false' "$(body_path ingest-replay-first)"; then
  echo "first replay ingest did not return duplicate=false" >&2
  cat "$(body_path ingest-replay-first)" >&2
  exit 1
fi

curl_capture "ingest-replay-duplicate" 200 \
  -X POST "${API_BASE_URL}/v1/replay-evidence" \
  -H "content-type: application/json" \
  -d "{\"contextType\":\"scrim\",\"contextId\":${scrim_id},\"submittedByTeamId\":${team_a_id},\"replayBody\":\"${replay_body}\",\"parserName\":\"kinetic-rl-parser\",\"parserVersion\":\"release-evidence\",\"parserConfigDigest\":\"default\",\"resultSubmissionId\":${submission_id},\"parseOutputJson\":{\"goals\":4}}"
if ! grep -q '"duplicate":true' "$(body_path ingest-replay-duplicate)"; then
  echo "duplicate replay ingest did not return duplicate=true" >&2
  cat "$(body_path ingest-replay-duplicate)" >&2
  exit 1
fi

curl_capture "ingest-replay-context-rejected" 409 \
  -X POST "${API_BASE_URL}/v1/replay-evidence" \
  -H "content-type: application/json" \
  -d "{\"contextType\":\"scrim\",\"contextId\":${second_scrim_id},\"submittedByTeamId\":${team_a_id},\"replayBody\":\"${replay_body}\",\"parserName\":\"kinetic-rl-parser\",\"parserVersion\":\"release-evidence\",\"parserConfigDigest\":\"default\",\"parseOutputJson\":{\"goals\":4}}"

curl_capture "list-replay-evidence" 200 "${API_BASE_URL}/v1/replay-evidence"
require_json_field "list replay evidence" "replaySha256" "$(body_path list-replay-evidence)"
curl_capture "list-replay-parse-runs" 200 "${API_BASE_URL}/v1/replay-parse-runs"
require_json_field "list replay parse runs" "parserVersion" "$(body_path list-replay-parse-runs)"
curl_capture "list-replay-links" 200 "${API_BASE_URL}/v1/result-submission-replay-links"
require_json_field "list replay links" "resultSubmissionId" "$(body_path list-replay-links)"

echo "[release-evidence] running real-browser auth/CORS evidence"
(
  cd "${ROOT_DIR}/web/client"
  npm ci
  RELEASE_EVIDENCE=1 \
  RELEASE_EVIDENCE_ARTIFACT_DIR="${ARTIFACT_DIR}/browser" \
  VITE_AUTH_MODE=api \
  VITE_API_BASE_URL="${API_BASE_URL}" \
  VITE_EVIDENCE_BASE_URL="http://127.0.0.1:9" \
  PLAYWRIGHT_WEB_PORT="${WEB_PORT}" \
  npx playwright test tests/e2e/release-evidence.spec.ts --output "${ARTIFACT_DIR}/playwright-output"
)

cat > "${ARTIFACT_DIR}/summary.md" <<EOF
# Release Evidence Summary

- Result: pass
- Timestamp: ${TIMESTAMP}
- Branch: ${branch}
- Commit: ${commit}
- API base URL: ${API_BASE_URL}
- Web base URL: ${WEB_BASE_URL}
- Replay submission ID: ${submission_id}
- Replay context: scrim:${scrim_id}
- Artifacts: ${ARTIFACT_DIR}

Covered gates:
- credentialed CORS allowed for the configured web origin
- disallowed CORS preflight rejected
- API cookie sessions preserve distinct player/support subjects
- player session cannot perform privileged result override
- replay intake accepts first upload, deduplicates second upload, and rejects same replay in another context
- browser API-mode login/session flow works without mocked localStorage or mocked API routes
EOF

echo "Release evidence passed. Artifacts: ${ARTIFACT_DIR}"
