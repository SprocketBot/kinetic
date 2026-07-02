# Week 8 Onboarding Notes (Scrim Orchestration Foundations)

Week 8 introduces scrim and unified matchmaking foundation primitives:

- Scrim lifecycle baseline (`created`, `in_progress`, `closed`, `voided`)
- Queue expansion-stage tracking (`expansion_stage`, `stage_advanced_at`)
- Queue-to-scrim promotion with decision observability metadata
- Player rating identity baseline (`player_ratings`) with read path

This is intentionally a foundation slice and does not implement the full async matchmaker.

## Related design docs

- `docs/adr/010-unified-matchmaking-rating-model.md`
- `docs/adr/011-unified-matchmaking-invariants-guardrails.md`
- `docs/adr/012-match-evidence-association-model.md`
- `docs/adr/013-replay-ingestion-and-submission-linking.md`
- `docs/matchmaking/unified-matchmaking-implementation-checklist.md`

## Prerequisites

- Week 7 smoke passes: `./tools/week7-smoke.sh`
- Postgres available on local test port (`55432` default)
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

Advance queue entry stage:

```bash
curl -s -X PATCH http://localhost:8080/v1/queue-entries \
  -H 'content-type: application/json' \
  -d '{"queueId":1,"teamId":1,"stage":2}'
```

Create scrim directly:

```bash
curl -s -X POST http://localhost:8080/v1/scrims \
  -H 'content-type: application/json' \
  -d '{"queueId":1,"homeTeamId":1,"awayTeamId":2,"state":"created"}'
```

Promote queue to scrim:

```bash
curl -s -X POST http://localhost:8080/v1/scrim-promotions \
  -H 'content-type: application/json' \
  -d '{"queueId":1}'
```

List new Week 8 resources:

```bash
curl -s http://localhost:8080/v1/scrims
curl -s http://localhost:8080/v1/player-ratings
curl -s http://localhost:8080/v1/matchmaking-decisions
```

## Expected constraint behavior

- queue stage must be `>= 1` (`PATCH /v1/queue-entries` -> `400` on invalid)
- queue stage cannot decrease for an active entry (`409`)
- scrim home/away teams must differ (`400`)
- scrim promotion requires at least two active queue entries (`409`)
- missing queue/team dependencies for scrim creation return `409`

## Test commands

```bash
go test ./...

# hierarchy-focused packages
go test ./internal/domain/hierarchy ./internal/platform/db ./internal/platform/http
```

## Full smoke

```bash
./tools/week8-smoke.sh
./tools/week8-k8s-smoke.sh
```

Notes:

- `week8-smoke.sh` runs API + Postgres locally via Docker.
- `week8-k8s-smoke.sh` deploys to minikube, port-forwards the service, and exercises the Week 8 flow against the in-cluster runtime.
- `week8-k8s-smoke.sh` creates a temporary in-cluster Postgres deployment/service for the test run and cleans it up on exit.
