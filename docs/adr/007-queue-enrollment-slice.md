# ADR-007: Week 6 Queue Enrollment Slice Contract

- Status: Accepted
- Date: 2026-02-14
- Owner: jacbaile

## Context

Week 5 completed roster assignment primitives.
Week 6 begins the queue + match flow MVP by adding deterministic queue enrollment behavior without introducing match orchestration complexity yet.

## Decision

Add `Queue` and `QueueEntry` primitives with join/leave/list API behavior.

### Entity contract

1. Queue
- `id` (bigserial, PK)
- `name` (text, required, unique)
- `slug` (text, required, unique)
- `is_active` (boolean, required, default `true`)
- `created_at` (timestamptz, required, default `now()`)

2. QueueEntry
- `id` (bigserial, PK)
- `queue_id` (bigint, FK -> `queues.id`, required)
- `team_id` (bigint, FK -> `teams.id`, required)
- `is_active` (boolean, required, default `true`)
- `created_at` (timestamptz, required, default `now()`)
- `left_at` (timestamptz, nullable)
- uniqueness: one active entry per `(queue_id, team_id)`

### API contract (Week 6 scope)

- `POST /v1/queues`
- `GET /v1/queues`
- `POST /v1/queue-entries` (join)
- `DELETE /v1/queue-entries` (leave)
- `GET /v1/queue-entries` (list active entries)

### Behavior contract

- queue listing is stable (`ORDER BY id`)
- active queue entry listing is deterministic FIFO per queue (`ORDER BY queue_id, created_at, id`)
- leave operation deactivates the active entry (`is_active=false`, `left_at=now()`)

### Validation contract

- queue create requires `name` and lowercase kebab-case `slug`
- queue join/leave requires `queueId > 0` and `teamId > 0`
- malformed payloads map to `400`
- FK violations and active-entry conflicts map to `409`

### Week 6 non-goals

- match creation and team pairing
- MMR/rating logic
- queue worker orchestration or timeout handling
- readiness checks and policy engine

## Consequences

### Positive

- Introduces a clean queue boundary for match flow MVP.
- Encodes active enrollment constraints in the database.
- Keeps deterministic ordering explicit and testable.

### Negative / Tradeoffs

- Queue behavior is synchronous/API-driven only in this slice.
- Additional lifecycle states beyond active/inactive are deferred.

## Follow-ups

- Add match candidate promotion from queue entries.
- Introduce queue policies (size, region, mode, readiness checks).
- Add queue metrics and operational visibility.
