#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
NAMESPACE="${K8S_NAMESPACE:-kinetic-v3}"
KUSTOMIZE_DIR="${ROOT_DIR}/deploy/k8s-local-dev"
API_PORT="${API_PORT:-8080}"
WEB_PORT="${WEB_PORT:-5173}"
WITH_WEB="false"
PF_PID=""
CLIENT_PID=""

usage() {
  cat <<EOF
Usage: ./tools/start-dev.sh [--with-web] [--api-port PORT] [--web-port PORT]

Starts the supported local dev path:
- deploy API and Postgres to minikube
- port-forward the API service to localhost
- optionally start the Vite web client on the host
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --with-web)
      WITH_WEB="true"
      shift
      ;;
    --api-port)
      API_PORT="$2"
      shift 2
      ;;
    --web-port)
      WEB_PORT="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

cleanup() {
  if [[ -n "${CLIENT_PID}" ]]; then
    kill "${CLIENT_PID}" >/dev/null 2>&1 || true
    wait "${CLIENT_PID}" 2>/dev/null || true
  fi
  if [[ -n "${PF_PID}" ]]; then
    kill "${PF_PID}" >/dev/null 2>&1 || true
    wait "${PF_PID}" 2>/dev/null || true
  fi
}
trap cleanup EXIT

cd "$ROOT_DIR"

echo "Deploying local dev overlay to minikube"
KUSTOMIZE_DIR="$KUSTOMIZE_DIR" K8S_NAMESPACE="$NAMESPACE" ./deploy/scripts/apply-local.sh

SERVICE_PORT="$(kubectl -n "$NAMESPACE" get svc kinetic-v3-api -o jsonpath='{.spec.ports[0].port}')"
if [[ -z "${SERVICE_PORT}" ]]; then
  echo "Unable to determine service port for ${NAMESPACE}/kinetic-v3-api" >&2
  exit 1
fi

echo "Starting API port-forward on localhost:${API_PORT}"
kubectl -n "$NAMESPACE" port-forward svc/kinetic-v3-api "${API_PORT}:${SERVICE_PORT}" >/tmp/kinetic_v3_start_dev_pf.log 2>&1 &
PF_PID=$!

for i in $(seq 1 40); do
  code="$(curl -s -o /tmp/kinetic_v3_start_dev_healthz.json -w '%{http_code}' "http://localhost:${API_PORT}/healthz" || true)"
  if [[ "$code" == "200" ]]; then
    break
  fi
  sleep 1
  if [[ "$i" -eq 40 ]]; then
    echo "API did not become reachable via port-forward in time" >&2
    cat /tmp/kinetic_v3_start_dev_pf.log >&2 || true
    exit 1
  fi
done

echo "API is available at http://localhost:${API_PORT}"
echo "Seed sample data with: ./tools/seed-dev.sh"
echo "Delete dev resources with: kubectl delete -k deploy/k8s-local-dev"

if [[ "${WITH_WEB}" == "true" ]]; then
  echo "Starting web client on localhost:${WEB_PORT}"
  (
    cd "${ROOT_DIR}/web/client"
    VITE_API_BASE_URL="http://localhost:${API_PORT}" \
    VITE_AUTH_MODE="${VITE_AUTH_MODE:-mock}" \
    npm run dev -- --host 0.0.0.0 --port "${WEB_PORT}"
  ) &
  CLIENT_PID=$!
  echo "Web client is available at http://localhost:${WEB_PORT}"
fi

echo "Press Ctrl+C to stop local processes. Kubernetes resources remain in minikube."
wait
