# Week 7 Onboarding Notes (Scheduled Competition Slice)

Week 7 adds scheduled competition primitives:

- Season
- ScheduleGroup (match week)
- Fixture (club vs club)
- Match (team vs team)

Scheduled hierarchy is:

`Season -> ScheduleGroup -> Fixture -> Match`

This is distinct from queue/scrim flow.

## Prerequisites

- Week 6 smoke passes: `./tools/week6-smoke.sh`
- Postgres available on local test port (`55432` default)

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

Create season:

```bash
curl -s -X POST http://localhost:8080/v1/seasons \
  -H 'content-type: application/json' \
  -d '{"name":"Season 1","slug":"season-1"}'
```

Create schedule group:

```bash
curl -s -X POST http://localhost:8080/v1/schedule-groups \
  -H 'content-type: application/json' \
  -d '{"seasonId":1,"name":"Week 1","sequence":1}'
```

Create fixture:

```bash
curl -s -X POST http://localhost:8080/v1/fixtures \
  -H 'content-type: application/json' \
  -d '{"scheduleGroupId":1,"homeClubId":1,"awayClubId":2}'
```

Create planned match:

```bash
curl -s -X POST http://localhost:8080/v1/matches \
  -H 'content-type: application/json' \
  -d '{"fixtureId":1,"homeTeamId":1,"awayTeamId":2,"state":"planned"}'
```

Create ready match (requires ratified schedule time):

```bash
curl -s -X POST http://localhost:8080/v1/matches \
  -H 'content-type: application/json' \
  -d '{"fixtureId":1,"homeTeamId":1,"awayTeamId":2,"state":"ready","scheduledFor":"2030-01-01T10:00:00Z","homeTimeRatifiedAt":"2030-01-01T08:00:00Z","awayTimeRatifiedAt":"2030-01-01T09:00:00Z"}'
```

## Expected constraint behavior

- invalid hierarchy IDs -> `409`
- equal home/away clubs or teams -> `400`
- `ready` without ratified schedule fields -> `400`
- malformed payloads -> `400`

## Test commands

```bash
go test ./...

# hierarchy-focused packages
go test ./internal/domain/hierarchy ./internal/platform/db ./internal/platform/http
```

## Optional full smoke

```bash
./tools/week7-smoke.sh
```
