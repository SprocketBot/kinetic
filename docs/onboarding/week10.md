# Week 10 Onboarding Notes (Promotion Processing MVP)

Week 10 adds queue promotion processing operations:

- synchronous promotion processing trigger endpoint
- idempotent re-run behavior for queue processing
- persisted promotion processing run summaries

## Related design docs

- `docs/adr/015-queue-promotion-processing-mvp.md`
- `docs/adr/014-scrim-promotion-ordering-and-lifecycle.md`

## Prerequisites

- Week 9 local smoke passes: `./tools/week9-smoke.sh`
- Week 9 minikube smoke passes: `./tools/week9-k8s-smoke.sh`
- Minikube running for in-cluster smoke: `minikube start`

## Run migrations

```bash
DATABASE_URL="postgres://postgres:postgres@localhost:55432/kinetic?sslmode=disable" \
MIGRATIONS_DIR="./migrations" \
go run ./cmd/migrate
```

## Start API with DB required

```bash
DATABASE_URL="postgres://postgres:postgres@localhost:55432/kinetic?sslmode=disable" \
REQUIRE_DATABASE=true \
go run ./cmd/api
```

## Manual API checks

Trigger processing for one queue:

```bash
curl -s -X POST http://localhost:8080/v1/scrim-promotions/process \
  -H 'content-type: application/json' \
  -d '{"queueId":1}'
```

Trigger processing for all active queues:

```bash
curl -s -X POST http://localhost:8080/v1/scrim-promotions/process \
  -H 'content-type: application/json' \
  -d '{"queueId":0}'
```

List processing run summaries:

```bash
curl -s http://localhost:8080/v1/promotion-processing-runs
```

## Expected behavior

- processing returns summary fields: `processedQueues`, `promotionsCreated`, `conflicts`
- re-running processing is safe (no duplicate promotions once queues are drained)
- processing runs are persisted and queryable via `/v1/promotion-processing-runs`

## Test commands

```bash
go test ./...

# hierarchy-focused packages
go test ./internal/domain/hierarchy ./internal/platform/db ./internal/platform/http
```

## Full smoke

```bash
./tools/week10-smoke.sh
./tools/week10-k8s-smoke.sh
```

Notes:

- `week10-smoke.sh` runs API + Postgres locally via Docker.
- `week10-k8s-smoke.sh` deploys to minikube, runs with temporary in-cluster Postgres, and validates Week 10 behavior against in-cluster runtime.
