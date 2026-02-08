#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
IMAGE="sprocket-v3-api:dev-$(date +%s)"

current_ctx="$(kubectl config current-context 2>/dev/null || true)"
if [[ "$current_ctx" != "minikube" ]]; then
  echo "Current kubectl context is '$current_ctx'. Expected 'minikube' for local deploy." >&2
  echo "Run: kubectl config use-context minikube" >&2
  exit 1
fi

echo "Building local image: $IMAGE"
docker build -t "$IMAGE" "$ROOT_DIR"

echo "Loading image into minikube"
minikube image load "$IMAGE"

echo "Applying Kubernetes manifests"
kubectl apply -k "$ROOT_DIR/deploy/k8s"

echo "Setting deployment image to $IMAGE"
kubectl -n sprocket-v3 set image deploy/sprocket-v3-api api="$IMAGE"
kubectl -n sprocket-v3 rollout status deploy/sprocket-v3-api --timeout=180s

echo "Sprocket v3 local deploy is healthy."
