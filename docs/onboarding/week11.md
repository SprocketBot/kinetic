# Week 11 Onboarding Notes (Result Submission + Ratification MVP)

Week 11 adds result submission lifecycle primitives:

- create/list result submissions for `scrim` or `match` context
- participant-team ratification path
- participant-team rejection path with reason

## Related design docs

- `docs/adr/016-result-submission-and-ratification-mvp.md`
- `docs/adr/012-replay-parsing-and-platform-account-association-model.md`
- `docs/adr/013-replay-parsing-invariants-and-guardrails.md`

## Prerequisites

- Week 10 local smoke passes: `./tools/week10-smoke.sh`
- Week 10 minikube smoke passes: `./tools/week10-k8s-smoke.sh`
- Minikube running for in-cluster smoke: `minikube start`

## Run migrations

```bash
DATABASE_URL="postgres://postgres:postgres@localhost:55432/sprocket?sslmode=disable" \
MIGRATIONS_DIR="./migrations" \
go run ./cmd/migrate
```

## Start API with DB required

```bash
DATABASE_URL="postgres://postgres:postgres@localhost:55432/sprocket?sslmode=disable" \
REQUIRE_DATABASE=true \
go run ./cmd/api
```

## Manual API checks

Create submission:

```bash
curl -s -X POST http://localhost:8080/v1/result-submissions \
  -H 'content-type: application/json' \
  -d '{"contextType":"scrim","contextId":1,"submittedByTeamId":1,"winningTeamId":1,"losingTeamId":2,"payloadJson":{"score":"3-1"}}'
```

Ratify by team:

```bash
curl -s -X POST http://localhost:8080/v1/result-submission-ratifications \
  -H 'content-type: application/json' \
  -d '{"submissionId":1,"teamId":1}'
```

Reject by team:

```bash
curl -s -X POST http://localhost:8080/v1/result-submission-rejections \
  -H 'content-type: application/json' \
  -d '{"submissionId":1,"teamId":2,"reason":"replay mismatch"}'
```

List submissions:

```bash
curl -s http://localhost:8080/v1/result-submissions
```

## Expected behavior

- invalid context/team relationships return `409`
- invalid payloads return `400`
- second ratification finalizes to `ratified`
- rejected submissions remain terminal

## Test commands

```bash
go test ./...

go test ./internal/domain/hierarchy ./internal/platform/db ./internal/platform/http
```

## Full smoke

```bash
./tools/week11-smoke.sh
./tools/week11-k8s-smoke.sh
```
