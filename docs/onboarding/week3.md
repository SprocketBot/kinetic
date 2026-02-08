# Week 3 Onboarding Notes (League Hierarchy Slice)

Week 3 introduces the first hierarchy domain slice:

- League
- Franchise
- Club

## Prerequisites

- Week 1 baseline smoke works: `./tools/week1-smoke.sh`
- A Postgres instance is available (default local test port: `55432`)

## Run migrations

```bash
DATABASE_URL="postgres://postgres:postgres@localhost:55432/sprocket?sslmode=disable" \
MIGRATIONS_DIR="./migrations" \
go run ./cmd/migrate
```

## Start API in DB-required mode

```bash
DATABASE_URL="postgres://postgres:postgres@localhost:55432/sprocket?sslmode=disable" \
REQUIRE_DATABASE=true \
go run ./cmd/api
```

## Manual API checks

Create league:

```bash
curl -s -X POST http://localhost:8080/v1/leagues \
  -H 'content-type: application/json' \
  -d '{"name":"Minor League Esports","slug":"minor-league-esports"}'
```

Create franchise:

```bash
curl -s -X POST http://localhost:8080/v1/franchises \
  -H 'content-type: application/json' \
  -d '{"leagueId":1,"name":"Guardians","slug":"guardians"}'
```

Create club:

```bash
curl -s -X POST http://localhost:8080/v1/clubs \
  -H 'content-type: application/json' \
  -d '{"franchiseId":1,"name":"Guardians RL","slug":"guardians-rl"}'
```

List entities:

```bash
curl -s http://localhost:8080/v1/leagues
curl -s http://localhost:8080/v1/franchises
curl -s http://localhost:8080/v1/clubs
```

## Expected constraint behavior

- duplicate `slug` values -> `409`
- missing parent FK (`leagueId` or `franchiseId`) -> `409`
- invalid slug format (not lowercase kebab-case) -> `400`

## Test commands

```bash
go test ./...

# hierarchy-focused packages
go test ./internal/domain/hierarchy ./internal/platform/db ./internal/platform/http
```
