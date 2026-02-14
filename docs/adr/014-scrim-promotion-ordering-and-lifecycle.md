# ADR-014: Scrim Promotion Ordering And Lifecycle

- Status: Accepted
- Date: 2026-02-14
- Owner: jacbaile
- Related: ADR-008 (scheduled match vs scrim), ADR-010 (unified matchmaking and rating model), ADR-011 (unified matchmaking invariants)
- Detailed References: `docs/matchmaking/unified-matchmaking-implementation-checklist.md`, `plans/v3-execution-board.md`

## Context

Week 8 established queue-to-scrim promotion foundations, but promotion is still effectively FIFO-first and does not yet encode deterministic rating-first ordering behavior.
Scrim lifecycle is also create-only in current APIs; lifecycle transitions need explicit, constrained behavior before submission/dispute work.

Without a clear contract, contributors will make incompatible assumptions about:

- which two teams should be promoted when more than two entries are active,
- how tie-breaks are resolved,
- what lifecycle transitions are legal,
- what decision metadata is mandatory for debugging and audits.

## Decision

Define deterministic promotion ordering and scrim lifecycle transitions as Week 9 contract.

### 1. Promotion ordering policy

Queue-to-scrim promotion selects two active entries from one queue using this deterministic priority:

1. Minimize team-rating distance (rating-first)
2. Minimize queue wait skew (older pairs preferred when rating distance ties)
3. Stable lexical tie-break by `(queue_entry.created_at, queue_entry.id)` ordering

This policy applies to one promotion transaction at a time per queue.

### 2. Team rating snapshot derivation

Team rating for promotion scoring is derived at promotion time as:

- active roster members for the team,
- each member's active unified rating for the queue context,
- baseline aggregate: arithmetic mean of available player ratings,
- fallback when no rating exists: default unified baseline (`1000`).

The derived value is a promotion-time snapshot and is written into decision metadata context (not treated as authoritative long-term team rating state).

### 3. Scrim lifecycle state machine

Allowed scrim states remain:

- `created`
- `in_progress`
- `closed`
- `voided`

Allowed transitions:

- `created -> in_progress`
- `created -> voided`
- `in_progress -> closed`
- `in_progress -> voided`

Disallowed transitions include (non-exhaustive):

- terminal to non-terminal (`closed -> *`, `voided -> *`)
- skipping execution (`created -> closed`)
- no-op transition (`state -> same state`)

### 4. Transition timestamp rules

- Entering `in_progress` sets `started_at` if unset.
- Entering terminal state (`closed` or `voided`) sets `ended_at`.
- `ended_at` must not be earlier than `started_at` when both exist.

### 5. Observability requirements

Each successful promotion must persist decision metadata sufficient to reconstruct ordering:

- queue wait seconds
- expansion stage used
- rating spread for selected pair
- cross-group flag
- ordering strategy identifier/version (new field if needed)

## Consequences

### Positive

- Contributors get one unambiguous promotion contract.
- Promotion outcomes become reproducible and debuggable.
- Lifecycle behavior is constrained before submission/dispute features land.

### Negative / Tradeoffs

- More query/service complexity than FIFO selection.
- Added schema/service overhead for richer decision metadata.
- Need explicit tests to keep deterministic behavior stable over refactors.

## Non-Goals (Current Decision Scope)

- async matchmaking worker orchestration,
- full uncertainty-weighted rating updates,
- replay/result application lifecycle,
- schedule-to-scrim convergence behavior.

## Follow-ups

- Implement promotion ordering path and lifecycle transition API in Week 9.
- Add deterministic ordering tests (unit + integration).
- Extend smoke scripts (local + minikube) to exercise lifecycle transitions.
