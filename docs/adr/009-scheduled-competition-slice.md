# ADR-009: Week 7 Scheduled Competition Slice Contract

- Status: Accepted
- Date: 2026-02-15
- Owner: jacbaile
- Related: ADR-008 (scheduled match model vs scrim model), ADR-010 (unified matchmaking and rating model), ADR-011 (unified matchmaking invariants and guardrails)

## Context

Week 6 delivered queue enrollment for scrims.
The platform also requires first-class scheduled league competition semantics, separate from scrims.

## Decision

Introduce scheduled competition primitives with this hierarchy:

`Season -> ScheduleGroup -> Fixture -> Match`

### Entity contract

1. Season
- `id` (bigserial, PK)
- `name` (text, required)
- `slug` (text, required, unique)
- `is_active` (boolean, required, default `true`)
- `created_at` (timestamptz, required, default `now()`)

2. ScheduleGroup (match week)
- `id` (bigserial, PK)
- `season_id` (bigint, FK -> `seasons.id`, required)
- `name` (text, required)
- `sequence` (int, required)
- `is_active` (boolean, required, default `true`)
- `created_at` (timestamptz, required, default `now()`)
- uniqueness: `(season_id, sequence)`

3. Fixture (club-vs-club pairing in a week)
- `id` (bigserial, PK)
- `schedule_group_id` (bigint, FK -> `schedule_groups.id`, required)
- `home_club_id` (bigint, FK -> `clubs.id`, required)
- `away_club_id` (bigint, FK -> `clubs.id`, required)
- `is_active` (boolean, required, default `true`)
- `created_at` (timestamptz, required, default `now()`)
- guardrail: home and away club must differ

4. Match (team-vs-team scheduled instance)
- `id` (bigserial, PK)
- `fixture_id` (bigint, FK -> `fixtures.id`, required)
- `home_team_id` (bigint, FK -> `teams.id`, required)
- `away_team_id` (bigint, FK -> `teams.id`, required)
- `state` (text, required; enum-like)
- `scheduled_for` (timestamptz, nullable)
- `home_time_ratified_at` (timestamptz, nullable)
- `away_time_ratified_at` (timestamptz, nullable)
- `created_at` (timestamptz, required, default `now()`)
- guardrail: home and away team must differ

### Lifecycle contract (Week 7 baseline)

- initial state: `planned`
- `ready` is allowed only when:
  - `scheduled_for` is set
  - `home_time_ratified_at` is set
  - `away_time_ratified_at` is set
- Week 7 includes create/list APIs only; state transition endpoints are deferred.

### API contract (Week 7 scope)

- `POST /v1/seasons`
- `GET /v1/seasons`
- `POST /v1/schedule-groups`
- `GET /v1/schedule-groups`
- `POST /v1/fixtures`
- `GET /v1/fixtures`
- `POST /v1/matches`
- `GET /v1/matches`

### Validation contract

- required IDs and names enforced at API boundary
- slug must be lowercase kebab-case where applicable
- `home_*` and `away_*` IDs must differ
- invalid lifecycle payloads map to `400`
- FK/uniqueness/dependency violations map to `409`

### Week 7 non-goals

- automatic schedule generation
- update/delete endpoints
- standings/MMR integration
- full submission/dispute workflow
- unified scrim matchmaking/rating implementation (covered by ADR-010/ADR-011)

## Consequences

### Positive

- Establishes scheduled league semantics independent from scrims.
- Makes ratified scheduling a first-class prerequisite for readiness.
- Creates clear schema/API foundations for later submission and ratification slices.

### Negative / Tradeoffs

- Adds another domain track that must stay synchronized with overall terminology.
- Lifecycle transitions are documented before full transition APIs are implemented.

## Follow-ups

- Add state transition endpoints with explicit transition matrix checks.
- Add submission/ratification workflow and dispute hooks.
- Add schedule import/generation tooling.
- Integrate scheduled match outcomes with unified rating update pipeline when scrim/match rating policy is finalized.
