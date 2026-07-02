# Week 9 Onboarding Notes (Deterministic Promotion + Scrim Lifecycle)

Week 9 extends scrim orchestration with:

- deterministic rating-first queue promotion ordering
- enriched matchmaking decision metadata for ordering rationale
- explicit scrim lifecycle transitions (`created -> in_progress -> closed|voided`)

## Related design docs

- `docs/adr/014-scrim-promotion-ordering-and-lifecycle.md`
- `docs/adr/010-unified-matchmaking-rating-model.md`
- `docs/adr/011-unified-matchmaking-invariants-guardrails.md`

## Prerequisites

- Week 8 local smoke passes: `./tools/week8-smoke.sh`
- Week 8 minikube smoke passes: `./tools/week8-k8s-smoke.sh`
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

Promote queue to scrim:

```bash
curl -s -X POST http://localhost:8080/v1/scrim-promotions \
  -H 'content-type: application/json' \
  -d '{"queueId":1}'
```

Transition scrim to `in_progress`:

```bash
curl -s -X PATCH http://localhost:8080/v1/scrims \
  -H 'content-type: application/json' \
  -d '{"scrimId":1,"state":"in_progress"}'
```

Transition scrim to `closed`:

```bash
curl -s -X PATCH http://localhost:8080/v1/scrims \
  -H 'content-type: application/json' \
  -d '{"scrimId":1,"state":"closed"}'
```

Inspect decision metadata:

```bash
curl -s http://localhost:8080/v1/matchmaking-decisions
```

## Expected constraint behavior

- scrim state target must be one of `in_progress`, `closed`, `voided`
- illegal transitions return `409` (for example `closed -> in_progress`)
- queue promotion requires at least two active entries (`409`)
- decision metadata includes ordering strategy and rating/wait rationale fields

## Test commands

```bash
go test ./...

# hierarchy-focused packages
go test ./internal/domain/hierarchy ./internal/platform/db ./internal/platform/http
```

## Full smoke

```bash
./tools/week9-smoke.sh
./tools/week9-k8s-smoke.sh
```

Notes:

- `week9-smoke.sh` runs API + Postgres locally via Docker.
- `week9-k8s-smoke.sh` deploys to minikube, runs with temporary in-cluster Postgres, and validates the week slice against in-cluster runtime.
