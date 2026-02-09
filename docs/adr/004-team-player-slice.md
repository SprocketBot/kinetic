# ADR-004: Week 4 Team and Player Slice Contract

- Status: Accepted
- Date: 2026-02-08
- Owner: jacbaile

## Context

Week 3 established League -> Franchise -> Club with create/list behavior.
Week 4 extends the hierarchy with Team and Player while keeping scope narrow enough for reliable weekly delivery.

## Decision

Add Team and Player as the next relational slice with create/list APIs.

### Entity contract

1. Team
- `id` (bigserial, PK)
- `club_id` (bigint, FK -> `clubs.id`, required)
- `name` (text, required)
- `slug` (text, required, unique)
- `is_active` (boolean, required, default `true`)
- `created_at` (timestamptz, required, default `now()`)
- uniqueness: `(club_id, name)`

2. Player
- `id` (bigserial, PK)
- `team_id` (bigint, FK -> `teams.id`, required)
- `display_name` (text, required)
- `slug` (text, required, unique)
- `is_active` (boolean, required, default `true`)
- `created_at` (timestamptz, required, default `now()`)
- uniqueness: `(team_id, display_name)`

### API contract (Week 4 scope)

- `POST /v1/teams`
- `GET /v1/teams`
- `POST /v1/players`
- `GET /v1/players`

### Validation contract

- required fields enforced at API boundary
- slug must be lowercase kebab-case
- FK violations map to `409`
- uniqueness violations map to `409`
- malformed request payloads map to `400`

### Week 4 non-goals

- roster spots/assignments
- transfer workflows
- update/delete for Team/Player
- search/pagination
- expanded role-scoped authz

## Consequences

### Positive

- Keeps hierarchy momentum with clear, additive schema.
- Unlocks next-week roster and seat assignment workflows.
- Preserves predictable API/test patterns from Week 3.

### Negative / Tradeoffs

- `player` tied to one team in this slice; multi-team participation is deferred.
- API may require additive changes when roster workflow is introduced.

## Follow-ups

- Introduce roster spots and membership history in Week 5+.
- Add update/deactivate semantics for Team/Player lifecycle.
- Evaluate scoped authz per organization layer for mutating routes.

