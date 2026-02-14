# Unified Matchmaking Invariants And Guardrails

Date: 2026-02-14  
Status: Draft  
Depends On: `docs/matchmaking/unified-matchmaking-core-concepts.md`

## Purpose

Define non-negotiable system rules ("invariants") and measurable guardrails for implementation and operations.

## Invariants

1. Single rating identity
- A player must have exactly one active unified rating for a given competitive context.
- Skill group is derived; it is not a separate source of truth for skill.

2. Group overlays, not hard walls
- Matchmaking cannot enforce strict same-group-only pairing.
- Cross-group eligibility is allowed when rating proximity satisfies active search constraints.

3. Hysteresis is required
- Promotion and demotion thresholds must differ for every boundary.
- Group assignment changes only when buffered thresholds are crossed.

4. Time-ordered expansion
- Search tolerance must widen monotonically with queue age.
- Expansion stages cannot tighten once widened for a given queue attempt.

5. Rating-first candidate ordering
- Candidate quality is sorted by rating proximity and uncertainty-aware confidence.
- Queue age acts as tie-breaker/escalation factor, not the primary quality metric.

6. Boundary-safe cross-group matching
- In early/mid search windows, cross-group matches are restricted to players near group boundaries.
- Full-spectrum cross-group matching is only allowed in late-stage expansion.

7. Bounded mismatch
- Every produced match must satisfy a maximum rating-gap cap tied to current expansion stage.
- Emergency widening is allowed only after configured wait thresholds.

8. Atomic rating update
- Match result processing must update all involved player ratings atomically.
- Partial updates are invalid and must roll back.

9. Observable matchmaking decisions
- Each match creation event must record decision metadata:
  - queue wait duration
  - expansion stage used
  - rating spread in the match
  - cross-group or same-group classification

10. Transparent progression contract
- API/UI must expose current group, thresholds, and transition reason codes after rating changes.

## Baseline Parameter Guardrails (Initial)

These are starting values from the proposal and are expected to be tuned with live data:

- Stage 1 (`0-2 min`): `+/-100`, same-group preference
- Stage 2 (`2-5 min`): `+/-150`, cross-group near boundary (`<= 50` from boundary)
- Stage 3 (`5+ min`): `+/-250`, expanded cross-group
- Stage 4 (`10+ min`): `+/-400`, emergency

Example hysteresis thresholds:

- Foundation promote `> 820`, demote `< 680`
- Academy promote `> 1220`, demote `< 1080`
- Champion promote `> 1620`, demote `< 1480`
- Master promote `> 2020`, demote `< 1880`

## Success Metrics Guardrails

Initial operating targets from the proposal:

- `>= 70%` of matches within `100` rating points
- `>= 50%` reduction in average queue time versus prior segmented model
- `<= 10%` oscillation rate for boundary players
- `<= 5%` unintended match quality drift after rollout

## Required Test Coverage

1. Unit tests
- group derivation and hysteresis behavior
- expansion stage selection by queue age
- candidate ranking logic with rating + uncertainty

2. Integration tests
- cross-group boundary matching behavior
- emergency widening behavior
- atomic rating update and rollback behavior

3. Simulation/regression tests
- synthetic population tests for queue-time and quality tradeoffs
- boundary-player oscillation checks
- rating distribution drift over long runs

## Open Calibration Questions

- Should thresholds vary by queue type (2s, 3s, BYOT) or remain global?
- How should uncertainty decay for inactive players?
- What is the minimum match quality score allowed in emergency stage?

## Change Management Rule

Any change to thresholds, overlap ranges, or expansion windows must include:

- updated docs
- migration/config change notes
- before/after simulation evidence
- rollback plan
