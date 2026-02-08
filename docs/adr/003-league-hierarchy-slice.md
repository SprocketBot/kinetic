# ADR-003: Week 3 League Hierarchy Slice Contract

- Status: Accepted
- Date: 2026-02-08
- Owner: jacbaile

## Context

Week 3 needs a concrete, low-risk domain slice to start the league management data model.
To keep delivery within 10h/week constraints, this slice is limited to three entities:

- `League`
- `Franchise`
- `Club`

This provides a stable foundation for `Team` and `Player` in later weeks without over-expanding scope now.

## Decision

Implement the first hierarchy slice with strict relational integrity and minimal API behavior.

### Entity contract

1. League
- `id` (bigserial, PK)
- `name` (text, required, unique)
- `slug` (text, required, unique)
- `is_active` (boolean, required, default `true`)
- `created_at` (timestamptz, required, default `now()`)

2. Franchise
- `id` (bigserial, PK)
- `league_id` (bigint, FK -> `leagues.id`, required)
- `name` (text, required)
- `slug` (text, required, unique)
- `is_active` (boolean, required, default `true`)
- `created_at` (timestamptz, required, default `now()`)
- uniqueness: `(league_id, name)` to prevent duplicate franchise names inside a league

3. Club
- `id` (bigserial, PK)
- `franchise_id` (bigint, FK -> `franchises.id`, required)
- `name` (text, required)
- `slug` (text, required, unique)
- `is_active` (boolean, required, default `true`)
- `created_at` (timestamptz, required, default `now()`)
- uniqueness: `(franchise_id, name)` to prevent duplicate club names inside a franchise

### API contract (Week 3 scope)

Create + list only.

- `POST /v1/leagues`
- `GET /v1/leagues`
- `POST /v1/franchises`
- `GET /v1/franchises`
- `POST /v1/clubs`
- `GET /v1/clubs`

### Request validation

- `name` and `slug` required for create operations
- `slug` is lowercase kebab-case (reject invalid format with `400`)
- FK violations return `409` with stable error body
- unique constraint violations return `409` with stable error body

### Week 3 non-goals

- Teams/Players
- pagination and filtering
- update/delete endpoints
- advanced RBAC checks on hierarchy endpoints (baseline auth behavior remains)

## Consequences

### Positive

- Limits schema risk and keeps Week 3 scoped to one practical vertical slice.
- Provides normalized relational foundation for Week 4 expansion.
- Keeps API surface understandable for new contributors.

### Negative / Tradeoffs

- Requires migration in Week 4 if additional optional fields are needed.
- Global unique `slug` may be stricter than future multi-tenant requirements; revisit if needed.

## Follow-ups

- Add `Team` and `Player` with FK relationships in Week 4.
- Add pagination/filtering after baseline create/list stability is proven.
- Add tighter authz around hierarchy mutations once roles/scopes are expanded.

