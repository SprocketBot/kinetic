# ADR-016: Result Submission And Ratification MVP

- Status: Accepted
- Date: 2026-02-14
- Owner: jacbaile
- Related: ADR-008 (scheduled match vs scrim), ADR-012 (replay parsing and platform account association model), ADR-013 (replay parsing invariants)

## Context

Week 11 starts the submission/ratification track before full replay ingestion automation.
The platform needs a concrete submission lifecycle that can attach results to scrims or scheduled matches and require both teams to ratify.

## Decision

Adopt a result-submission MVP lifecycle with explicit pending/ratified/rejected states.

### Lifecycle

- `pending` on submission creation
- `ratified` when both context teams ratify
- `rejected` when a participant team rejects with reason

### Scope

- Supports `scrim` and scheduled `match` contexts.
- Submission payload is stored as JSON evidence metadata.
- Ratification/rejection permissions are limited to context participant teams.

### API MVP

- `POST /v1/result-submissions`
- `GET /v1/result-submissions`
- `POST /v1/result-submission-ratifications`
- `POST /v1/result-submission-rejections`

### Constraints

- `winningTeamId` and `losingTeamId` must be context participants and must differ.
- Only one `pending` submission per context at a time.
- Terminal states cannot transition back to `pending` or to each other.

## Consequences

### Positive

- Defines auditable result flow ahead of automated replay application.
- Enables contributor-visible behavior and functional tests for Week 11.

### Tradeoffs

- Requires polymorphic context checks in service/store logic.
- Replay bytes and parser provenance remain out-of-band until Week 12.

## Follow-ups

- Add replay evidence entity and parser provenance linkage (Week 12).
- Add explicit dispute-resolution workflow beyond single-step rejection.
