# Week 6 Onboarding Notes (Queue Enrollment Slice)

Week 6 adds:

- Queue
- QueueEntry (join/leave/list)

Queue enrollment chain is now:

`Team -> Queue -> QueueEntry(active/inactive)`

## Prerequisites

- Week 5 smoke passes: `./tools/week5-smoke.sh`
- Postgres available on local test port (`55432` default)

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

Create queue:

```bash
curl -s -X POST http://localhost:8080/v1/queues \
  -H 'content-type: application/json' \
  -d '{"name":"3v3 Ranked","slug":"3v3-ranked"}'
```

Join queue:

```bash
curl -s -X POST http://localhost:8080/v1/queue-entries \
  -H 'content-type: application/json' \
  -d '{"queueId":1,"teamId":1}'
```

Leave queue:

```bash
curl -s -X DELETE http://localhost:8080/v1/queue-entries \
  -H 'content-type: application/json' \
  -d '{"queueId":1,"teamId":1}'
```

List active entries:

```bash
curl -s http://localhost:8080/v1/queue-entries
```

## Expected constraint behavior

- duplicate active queue entry for same (`queueId`, `teamId`) -> `409`
- missing parent FK (`queueId` or `teamId`) -> `409`
- leaving when no active entry exists -> `409`
- invalid IDs or slug format -> `400`

## Test commands

```bash
go test ./...

# hierarchy-focused packages
go test ./internal/domain/hierarchy ./internal/platform/db ./internal/platform/http
```

## Optional full smoke

```bash
./tools/week6-smoke.sh
```
