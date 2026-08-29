# Week 1 Onboarding Runbook

This runbook gets a new contributor from clone to a running API in under 30 minutes.

## Prerequisites

- Go `1.24.6`
- Docker
- `kubectl` (optional for K8s validation)

## 1) Clone and test baseline

```bash
git clone <repo-url> Kinetic-v3
cd Kinetic-v3
go test ./...
```

Expected: all tests pass.

## 2) Run API locally (no DB required)

```bash
go run ./cmd/api
```

In another terminal:

```bash
curl -sSf http://localhost:8080/healthz
curl -sSf http://localhost:8080/readyz
```

Expected: JSON responses with `status: ok` and `status: ready`.

## 3) Start local Postgres (optional, for DB work)

```bash
docker run --name kinetic-v3-pg \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=kinetic \
  -p 55432:5432 \
  -d postgres:16
```

## 4) Run migrations

```bash
DATABASE_URL="postgres://postgres:postgres@localhost:55432/kinetic?sslmode=disable" \
MIGRATIONS_DIR="./migrations" \
go run ./cmd/migrate
```

Expected: migration command exits successfully and reports applied count.

## 5) Run API with required DB connectivity

```bash
DATABASE_URL="postgres://postgres:postgres@localhost:55432/kinetic?sslmode=disable" \
REQUIRE_DATABASE=true \
go run ./cmd/api
```

Expected: API starts and logs `database connectivity validated`.

## 6) Optional: apply Kubernetes baseline

```bash
./deploy/scripts/apply-local.sh
```

Notes:

- This script expects `kubectl` context `minikube`.
- It builds a local API image, loads it into minikube, applies manifests, sets the deployment image, and waits for rollout.

## 7) Optional: run full Week 1 smoke script

```bash
./tools/week1-smoke.sh
```

Notes:

- Uses Postgres host port `55432` by default to avoid collisions with existing local services.
- Override with `PG_PORT=<port>` if needed.
