# ADR-011: Unified Matchmaking Invariants And Guardrails

- Status: Accepted
- Date: 2026-02-14
- Owner: jacbaile
- Related: ADR-010 (unified matchmaking and rating model)
- Source Proposal: <https://minor-league-esports.github.io/knowledgeBase/departments/development/features-and-designs/kinetic-v2-unified-matchmaking-proposal/>
- Detailed References: `docs/matchmaking/unified-matchmaking-invariants-and-guardrails.md`, `docs/matchmaking/unified-matchmaking-implementation-checklist.md`

## Context

A unified matchmaking model without explicit invariant rules will drift over time as thresholds, queue logic, and APIs evolve.
The system needs clear non-negotiables for correctness and consistent live tuning practices.

## Decision

Define and enforce invariant rules and operating guardrails for unified matchmaking.

### Invariants

1. Single rating identity
- One active unified rating per player per competitive context.
- Skill group is derived, never a competing skill source of truth.

2. Group overlays, not hard walls
- Same-group-only pairing cannot be a hard constraint.
- Cross-group eligibility is permitted when active rating constraints allow it.

3. Hysteresis required
- Promotion and demotion thresholds differ at every group boundary.

4. Monotonic expansion
- Search tolerances widen as queue age increases and never re-tighten during one queue attempt.

5. Rating-first ordering
- Candidate ordering prioritizes rating proximity with uncertainty-aware confidence.
- Queue age is escalation/tie-breaker input.

6. Boundary-safe cross-group behavior
- Early/mid windows allow cross-group matching only near boundaries.
- Full-spectrum cross-group pairing is late-stage only.

7. Bounded mismatch
- Every match satisfies a maximum rating-gap cap for its active expansion stage.

8. Atomic rating updates
- Match result updates for all participants are atomic; partial updates are invalid.

9. Decision observability
- Persist decision metadata for each created match (queue age, expansion stage, rating spread, cross-group flag).

10. Progression transparency
- API/UI must expose post-update group state and transition reason codes.

### Baseline guardrail values (initial)

- Stage 1 (`0-2 min`): `+/-100`, same-group preference
- Stage 2 (`2-5 min`): `+/-150`, boundary cross-group (`<= 50`)
- Stage 3 (`5+ min`): `+/-250`, broader cross-group
- Stage 4 (`10+ min`): `+/-400`, emergency

Example hysteresis thresholds:

- Foundation promote `> 820`, demote `< 680`
- Academy promote `> 1220`, demote `< 1080`
- Champion promote `> 1620`, demote `< 1480`
- Master promote `> 2020`, demote `< 1880`

### Success metrics guardrails (initial)

- `>= 70%` of matches within `100` rating points
- `>= 50%` queue-time reduction vs segmented baseline
- `<= 10%` boundary oscillation rate
- `<= 5%` unintended match-quality drift

### Change management rule

Any threshold/overlap/expansion change must ship with:

- docs update
- migration/config notes
- before/after simulation evidence
- rollback plan

## Consequences

### Positive

- Improves correctness and auditability of matchmaking behavior.
- Makes parameter tuning safer and more repeatable.
- Produces clearer product and operational accountability.

### Negative / Tradeoffs

- Adds process overhead to tuning iterations.
- Requires telemetry and simulation capabilities earlier.

## Follow-ups

- Build invariant test suite and simulation harness.
- Add operational dashboards for queue quality and transition stability.
- Produce implementation checklist mapped by invariant.
