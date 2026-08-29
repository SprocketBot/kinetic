# Week 4 Onboarding Notes (Team and Player Slice)

Week 4 extends the hierarchy with:

- Team
- Player

Hierarchy chain is now:

`League -> Franchise -> Club -> Team -> Player`

## Prerequisites

- Week 3 smoke passes: `./tools/week3-smoke.sh`
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

Create team:

```bash
curl -s -X POST http://localhost:8080/v1/teams \
  -H 'content-type: application/json' \
  -d '{"clubId":1,"name":"Guardians Alpha","slug":"guardians-alpha"}'
```

Create player:

```bash
curl -s -X POST http://localhost:8080/v1/players \
  -H 'content-type: application/json' \
  -d '{"displayName":"Player One","slug":"player-one"}'
```

List entities:

```bash
curl -s http://localhost:8080/v1/teams
curl -s http://localhost:8080/v1/players
```

## Expected constraint behavior

- duplicate team/player `slug` -> `409`
- missing parent FK (`clubId` for team) -> `409`
- invalid slug format -> `400`

## Test commands

```bash
go test ./...

# hierarchy-focused packages
go test ./internal/domain/hierarchy ./internal/platform/db ./internal/platform/http
```

## Optional full smoke

```bash
./tools/week4-smoke.sh
```
