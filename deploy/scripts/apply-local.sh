#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
NAMESPACE="${K8S_NAMESPACE:-kinetic-v3}"
KUSTOMIZE_DIR="${KUSTOMIZE_DIR:-$ROOT_DIR/deploy/k8s}"
IMAGE_REPO="${IMAGE_REPO:-kinetic-v3-api}"
IMAGE_TAG="${IMAGE_TAG:-dev-$(date +%s)}"
IMAGE="${IMAGE_REPO}:${IMAGE_TAG}"

current_ctx="$(kubectl config current-context 2>/dev/null || true)"
if [[ "$current_ctx" != "minikube" ]]; then
  echo "Current kubectl context is '$current_ctx'. Expected 'minikube' for local deploy." >&2
  echo "Run: kubectl config use-context minikube" >&2
  exit 1
fi

if ! kubectl cluster-info >/dev/null 2>&1; then
  echo "Minikube cluster is not reachable." >&2
  echo "Run: minikube start" >&2
  exit 1
fi

echo "Building image in minikube: $IMAGE"
(
  cd "$ROOT_DIR"
  minikube image build -t "$IMAGE" -f Dockerfile .
)

if ! minikube ssh -- docker image inspect "$IMAGE" >/dev/null 2>&1; then
  echo "Built image is not present in minikube: $IMAGE" >&2
  exit 1
fi

echo "Applying Kubernetes manifests from ${KUSTOMIZE_DIR}"
kubectl apply -k "$KUSTOMIZE_DIR"

echo "Setting deployment image to $IMAGE"
kubectl -n "$NAMESPACE" set image deploy/kinetic-v3-api api="$IMAGE"
kubectl -n "$NAMESPACE" rollout status deploy/kinetic-v3-api --timeout=180s

echo "Kinetic v3 local deploy is healthy."
