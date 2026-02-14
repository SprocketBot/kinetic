# ADR-008: Scheduled Match Model vs Scrim Model

- Status: Accepted
- Date: 2026-02-14
- Owner: jacbaile

## Context

The platform needs to support both:

- scheduled league play
- ad hoc queued games

Prior terminology conflated these and caused ambiguity around `match` semantics.

## Decision

Define `match` as scheduled league competition only.
Define queue-created ad hoc games as `scrim`.

### Scheduled hierarchy contract

`season -> schedule_group (match week) -> fixture -> match`

1. Season
- top-level competition period

2. ScheduleGroup
- week/bucket inside a season for scheduled play

3. Fixture
- club-vs-club scheduled pairing in a schedule group

4. Match
- team-vs-team scheduled pairing under a fixture

### Scrim contract

- Scrim is created from queue enrollment flow
- Scrim is not part of season scheduling hierarchy
- Scrim and Match have separate lifecycle/state models

## Match lifecycle contract

1. `planned`
- match exists under fixture, teams are known
- play time not yet ratified

2. `ready`
- play time is scheduled by teams and ratified by both teams
- this is the explicit gate to readiness

3. `in_progress`
- match is actively being played

4. `submitted`
- result/evidence submitted

5. `under_review`
- dispute/review or ratification checks in progress

6. `finalized` (terminal)
- official outcome locked

7. `voided` (terminal alternative)
- invalid/cancelled with reason

### Readiness clarification

A match is **not** ready only because it is pre-created in the league schedule.
A match becomes ready only after the teams agree on an exact play time and that time is ratified by both teams.

## Consequences

### Positive

- Removes ambiguity between scheduled and ad hoc game flows.
- Preserves league scheduling requirements as first-class concepts.
- Allows queue/scrim delivery to continue without polluting scheduled match semantics.

### Negative / Tradeoffs

- Requires explicit dual-domain naming and lifecycle handling in upcoming slices.
- May require renaming previously planned "match flow" milestones to "scrim flow".

## Follow-ups

- Update roadmap language to separate scrim flow from scheduled match flow.
- Add scheduled competition schema and APIs in a dedicated week slice.
- Define scrim lifecycle ADR to mirror this separation explicitly.
