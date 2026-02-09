# Sprocket v3

Sprocket v3 is a reimagining of the platform as a statically typed, compiled backend (Go) with Kubernetes-native operations.

## Current Status

Week 1 is in progress. Foundation work is focused on:

- modular monolith repository layout
- architecture decision records (ADRs)
- service bootstrap (health/readiness, config, logging)
- PostgreSQL wiring and migrations
- Kubernetes deployment baseline

Planning source of truth:

- `/Users/jacbaile/Sprocket-v3/plans/v3-execution-board.md`

## Repository Layout

- `cmd/`: entrypoints (binaries)
- `internal/`: app-private code (domain + platform)
- `pkg/`: optional reusable packages intended for external use
- `migrations/`: database migration files
- `deploy/`: Kubernetes manifests/charts and deployment scripts
- `docs/`: ADRs, runbooks, onboarding
- `test/`: integration/functional test assets
- `tools/`: development tooling scripts

## Working Agreement

- Ship vertical slices weekly.
- Every shipped slice includes code + tests + docs.
- Keep architecture simple: modular monolith first, split later only with evidence.

## Quick Commands

```bash
# test
go test ./...

# run API
go run ./cmd/api

# run migrations
go run ./cmd/migrate

# run week 1 smoke checks
./tools/week1-smoke.sh

# run week 3 hierarchy smoke checks
./tools/week3-smoke.sh

# run week 4 team/player smoke checks
./tools/week4-smoke.sh
```

Week 1 onboarding guide:

- `/Users/jacbaile/Sprocket-v3/docs/onboarding/week1.md`
- `/Users/jacbaile/Sprocket-v3/docs/onboarding/week2.md`
- `/Users/jacbaile/Sprocket-v3/docs/onboarding/week3.md`
- `/Users/jacbaile/Sprocket-v3/docs/onboarding/week4.md`
