# ADR-005: Week 5 Roster Membership Slice Contract

- Status: Accepted
- Date: 2026-02-09
- Owner: jacbaile

## Context

Week 4 introduced Team and Player with create/list APIs.
Week 5 needs explicit team assignment primitives that are auditable and enforce duplicate-active membership constraints.

## Decision

Add `RosterMembership` as an explicit Player-to-Team relationship with create/list APIs.

### Entity contract

1. RosterMembership
- `id` (bigserial, PK)
- `player_id` (bigint, FK -> `players.id`, required)
- `team_id` (bigint, FK -> `teams.id`, required)
- `is_active` (boolean, required, default `true`)
- `created_at` (timestamptz, required, default `now()`)
- uniqueness: active duplicate prevention via partial unique index on `(player_id, team_id)` where `is_active=true`

### API contract (Week 5 scope)

- `POST /v1/roster-memberships`
- `GET /v1/roster-memberships`

### Validation contract

- required IDs enforced at API boundary (`playerId > 0`, `teamId > 0`)
- malformed request payloads map to `400`
- FK violations map to `409`
- duplicate active pair violations map to `409`

### Week 5 non-goals

- transfer approvals
- historical lifecycle transitions (inactive/end timestamps)
- bulk assignment APIs
- role-scoped roster authorization expansion

## Consequences

### Positive

- Makes roster assignment explicit and queryable.
- Enforces active duplicate guardrails in the DB.
- Preserves additive, test-first vertical slice pattern.

### Negative / Tradeoffs

- `Player.team_id` and `RosterMembership.team_id` can diverge until a consolidation strategy is introduced.
- Historical membership lifecycle is deferred, so current-state behavior only is modeled.

## Follow-ups

- Define transfer workflow and deactivation semantics.
- Reconcile or deprecate direct `player.team_id` ownership semantics.
- Add filtered list/query endpoints (`activeOnly`, by team/player).
