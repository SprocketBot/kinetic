# Dev Setup And Verification Runbook

## Prerequisites

- Go `1.24.6`
- Docker
- kubectl
- minikube

## Local baseline

```bash
go test ./...
./tools/quality-gate.sh
```

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
- `kubectl -n sprocket-v3 get pods`
- `kubectl -n sprocket-v3 logs deploy/sprocket-v3-api --tail=200`
- verify service port discovery in smoke script logs
