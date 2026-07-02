# Week 12 Onboarding Notes (Replay Evidence + Parser Provenance MVP)

Week 12 adds replay-ingestion primitives with idempotent evidence identity and parser provenance tracking:

- ingest replay evidence with parser metadata
- deduplicate by replay content hash
- link replay evidence to result submissions
- list evidence, parse runs, and submission links

## Related design docs

- `docs/adr/017-replay-evidence-and-parser-provenance-mvp.md`
- `docs/adr/012-replay-parsing-and-platform-account-association-model.md`
- `docs/adr/013-replay-parsing-invariants-and-guardrails.md`
- `docs/adr/016-result-submission-and-ratification-mvp.md`

## Prerequisites

- Week 11 local smoke passes: `./tools/week11-smoke.sh`
- Week 11 minikube smoke passes: `./tools/week11-k8s-smoke.sh`
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

Ingest replay evidence:

```bash
curl -s -X POST http://localhost:8080/v1/replay-evidence \
  -H 'content-type: application/json' \
  -d '{"contextType":"scrim","contextId":1,"submittedByTeamId":1,"replayBody":"sample-replay-bytes","parserName":"kinetic-rl-parser","parserVersion":"v0.1.0","parserConfigDigest":"cfg-week12","resultSubmissionId":1,"parseOutputJson":{"goals":4}}'
```

Duplicate ingest of same replay body:

```bash
curl -s -X POST http://localhost:8080/v1/replay-evidence \
  -H 'content-type: application/json' \
  -d '{"contextType":"scrim","contextId":1,"submittedByTeamId":1,"replayBody":"sample-replay-bytes","parserName":"kinetic-rl-parser","parserVersion":"v0.1.0","parserConfigDigest":"cfg-week12","resultSubmissionId":1,"parseOutputJson":{"goals":4}}'
```

List replay evidence:

```bash
curl -s http://localhost:8080/v1/replay-evidence
```

List replay parse runs:

```bash
curl -s http://localhost:8080/v1/replay-parse-runs
```

List result submission replay links:

```bash
curl -s http://localhost:8080/v1/result-submission-replay-links
```

## Expected behavior

- first replay ingest returns `201` with `duplicate=false`
- duplicate replay ingest returns `200` with `duplicate=true`
- replay evidence rows are unique per content hash
- parser provenance rows grow with each ingest attempt
- submission links enforce matching context and participant ownership

## Test commands

```bash
go test ./...

go test ./internal/domain/hierarchy ./internal/platform/db ./internal/platform/http
```

## Full smoke

```bash
./tools/week12-smoke.sh
./tools/week12-k8s-smoke.sh
```
