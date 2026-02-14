# ADR-010: Unified Matchmaking And Rating Model

- Status: Accepted
- Date: 2026-02-14
- Owner: jacbaile
- Related: ADR-007 (queue enrollment), ADR-008 (scheduled match vs scrim), ADR-009 (scheduled competition slice)
- Source Proposal: <https://minor-league-esports.github.io/knowledgeBase/departments/development/features-and-designs/sprocket-v2-unified-matchmaking-proposal/>
- Detailed Reference: `docs/matchmaking/unified-matchmaking-core-concepts.md`

## Context

Segmented matchmaking/rating pools by skill group create avoidable queue fragmentation, rating comparability gaps, and unstable boundary behavior.
Sprocket v3 needs one coherent model that preserves skill-group identity while matching and rating players on a shared scale.

## Decision

Adopt a unified matchmaking and rating model built on one continuous rating spectrum.

### 1. Single rating pool

- Each player has one primary skill rating (`unified_elo_rating`) per competitive context.
- Skill group membership is derived from rating and transition rules, not an independent rating silo.

### 2. Skill groups as overlapping overlays

Skill groups remain a UX and organizational primitive, but not a hard matchmaking barrier.

Initial overlapping ranges:

- Foundation: `0-800`
- Academy: `700-1200`
- Champion: `1100-1600`
- Master: `1500-2000`
- Premier: `1900+`

### 3. Hysteresis for promotions/demotions

Group transitions require separate promotion and demotion thresholds to avoid boundary oscillation.

Example:

- Foundation -> Academy promotion above `820`
- Academy -> Foundation demotion below `680`

### 4. Unified queue evaluation

- Players may queue for multiple supported scrim types.
- Matchmaker evaluates one combined eligible pool per mode/type and prioritizes rating proximity first.

### 5. Time-based search expansion

Search tolerance widens with queue time:

1. `0-2 min`: same-group preference, `+/-100`
2. `2-5 min`: `+/-150`, cross-group near boundaries (`<= 50`)
3. `5+ min`: `+/-250`, broader cross-group
4. `10+ min`: `+/-400` emergency expansion

### 6. Rating uncertainty

Track an uncertainty value (`elo_uncertainty`, Glicko-like RD) with rating to support confidence-aware updates and matching.

### 7. Transparent progression UX

Expose rating, current group, promotion/demotion thresholds, and match-formation rationale in player-visible surfaces.

## Consequences

### Positive

- Reduces queue fragmentation and expected wait times.
- Improves rating comparability across the ecosystem.
- Stabilizes boundary behavior with buffered transitions.
- Preserves group identity without forcing hard pool isolation.

### Negative / Tradeoffs

- Requires careful calibration and observability to avoid quality drift.
- Introduces additional complexity in candidate ranking logic.
- Creates stronger dependency on robust rating-update correctness.

## Non-Goals (Current Decision Scope)

- Full schema/API implementation details.
- Final tuned thresholds for all queue modes.
- Rollout plan and operational runbooks.

## Follow-ups

- Define invariant-level guardrails and change controls (ADR-011).
- Implement simulation harness for threshold tuning and drift detection.
- Add product telemetry and dashboards for queue quality and transition stability.
