# Dev Setup And Verification Runbook

## Prerequisites

- Go `1.24.6`
- Docker for `./tools/week12-smoke.sh`
- kubectl
- minikube
- Node.js 22.x and npm 10+ for the web client

## Local baseline

```bash
go test ./...
./tools/quality-gate.sh
```

## Local K8s-backed dev environment

```bash
minikube start
kubectl config use-context minikube
./tools/start-dev.sh
```

In a second shell, seed sample data:

```bash
./tools/seed-dev.sh
```

To pair the web client with the Minikube-backed API:

```bash
./tools/start-dev.sh --with-web
```

What this provides:

- API and Postgres running inside Minikube
- API port-forwarded to `http://localhost:8080`
- optional Vite web client on `http://localhost:5173`
- no host Postgres container for the supported K8s dev path

## Local full-stack validation

```bash
./tools/week12-smoke.sh
```

What this validates:

- migrations against ephemeral Postgres
- API start in DB-required mode
- queue/scrim flow
- result submission + ratification/rejection
- replay evidence ingestion + dedupe

## Minikube full-stack validation

```bash
kubectl config use-context minikube
minikube status
./tools/week12-k8s-smoke.sh
```

What this validates:

- container build and deploy to local cluster
- in-cluster Postgres wiring
- Kubernetes service reachability via port-forward
- full behavior flow through deployed service

## Troubleshooting quick checks

- `kubectl config current-context` should be `minikube`
- `kubectl -n kinetic-v3 get deploy,svc,pods`
- `kubectl -n kinetic-v3 get pods`
- `kubectl -n kinetic-v3 logs deploy/kinetic-v3-api --tail=200`
- `kubectl -n kinetic-v3 logs deploy/kinetic-v3-pg-dev --tail=200`
- verify service port discovery in smoke script logs
- stop only local processes with `Ctrl+C`; remove dev cluster resources with `kubectl delete -k deploy/k8s-local-dev`
