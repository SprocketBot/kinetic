# Week 5 Onboarding Notes (Roster Membership Slice)

Week 5 adds:

- RosterMembership

Hierarchy + roster chain is now:

`League -> Franchise -> Club -> Team -> Player -> RosterMembership`

## Prerequisites

- Week 4 smoke passes: `./tools/week4-smoke.sh`
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

Create roster membership:

```bash
curl -s -X POST http://localhost:8080/v1/roster-memberships \
  -H 'content-type: application/json' \
  -d '{"playerId":1,"teamId":1}'
```

List roster memberships:

```bash
curl -s http://localhost:8080/v1/roster-memberships
```

## Expected constraint behavior

- duplicate active (`playerId`, `teamId`) pair -> `409`
- missing parent FK (`playerId` or `teamId`) -> `409`
- invalid IDs (`<= 0`) -> `400`

## Test commands

```bash
go test ./...

# hierarchy-focused packages
go test ./internal/domain/hierarchy ./internal/platform/db ./internal/platform/http
```

## Optional full smoke

```bash
./tools/week5-smoke.sh
```
