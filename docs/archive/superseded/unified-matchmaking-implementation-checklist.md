# Unified Matchmaking Implementation Checklist

Date: 2026-02-14  
Status: Draft  
Depends On: `docs/adr/010-unified-matchmaking-rating-model.md`, `docs/adr/011-unified-matchmaking-invariants-guardrails.md`

## Purpose

Map each matchmaking invariant to actionable implementation tasks across schema, API, services, operations, and tests.

## How to use

- Treat each checklist block as a vertical slice.
- Do not mark an invariant complete without test coverage and telemetry.
- Keep thresholds in config, not hardcoded service logic.

## Week 8 Mapping (Implemented 2026-02-15)

- Implemented foundations:
  - Invariant 1: `player_ratings` schema, unique active identity index, and read endpoint.
  - Invariant 4: queue entry monotonic stage tracking (`expansion_stage`, `stage_advanced_at`) with regression rejection.
  - Invariant 9: `matchmaking_decisions` persistence during queue-to-scrim promotion and read endpoint.
- Deferred by design:
  - ADR-012 and ADR-013 replay/evidence workflows.
  - async matcher orchestration and production rating update pipeline.

## Week 9 Mapping (Implemented 2026-02-14)

- Implemented foundations:
  - Invariant 5: rating-first deterministic candidate selection with queue-age skew tie-breaks.
  - Invariant 9: decision metadata enriched with ordering strategy and team rating/wait rationale fields.
- Still deferred:
  - full uncertainty-weighted candidate confidence model and property-based ordering test harness.

## Invariant 1: Single Rating Identity

Schema tasks:

- [x] Add `player_ratings` table with unique active row per `(player_id, context_key)`.
- [x] Add columns: `rating`, `uncertainty`, `matches_played`, `last_competed_at`, `updated_at`.
- [x] Add unique partial index for active identity and FK to `players`.

API tasks:

- [x] Add read endpoint for player rating snapshot.
- [ ] Ensure no write endpoint allows duplicate active ratings per context.

Service tasks:

- [ ] Implement rating repository with upsert semantics keyed by `(player_id, context_key)`.
- [ ] Enforce derived group computation from rating at read-time or materialized update-time.

Testing tasks:

- [ ] Unit test duplicate active identity rejection.
- [ ] Integration test uniqueness and FK constraints.

## Invariant 2: Group Overlays, Not Hard Walls

Schema tasks:

- [ ] Add `skill_group_ranges` config table or versioned config document source.
- [ ] Store overlapping range boundaries and effective dates.

API tasks:

- [ ] Expose effective group ranges in config/metadata endpoint.

Service tasks:

- [ ] Remove strict same-group filter from candidate eligibility.
- [ ] Evaluate candidates by rating distance under active stage constraints.

Testing tasks:

- [ ] Integration test valid cross-group match near overlapping boundaries.

## Invariant 3: Hysteresis Required

Schema tasks:

- [ ] Add `skill_group_thresholds` config with explicit `promote_above` and `demote_below`.

API tasks:

- [ ] Expose thresholds in progression payloads.

Service tasks:

- [ ] Implement transition engine using previous group + buffered thresholds.
- [ ] Emit reason code for promotion/demotion/no-change decisions.

Testing tasks:

- [ ] Unit tests for no-flap behavior around each boundary.
- [ ] Regression tests for boundary players over repeated small deltas.

## Invariant 4: Monotonic Expansion

Schema tasks:

- [x] Add queue attempt tracking fields: `enqueued_at`, `current_stage`, `stage_advanced_at`.

API tasks:

- [ ] Return stage and effective tolerance in queue status payload.

Service tasks:

- [ ] Implement stage function from queue age.
- [x] Block stage regression within one active queue attempt.

Testing tasks:

- [ ] Unit tests for stage transitions at boundary timestamps.
- [ ] Integration test preventing stage rollback.

## Invariant 5: Rating-First Ordering

Schema tasks:

- [ ] Add optional materialized candidate quality fields if needed for observability.

API tasks:

- [ ] Expose candidate quality explanation in debug/admin endpoint.

Service tasks:

- [ ] Implement scoring: rating distance + uncertainty confidence + queue-age tie-break.
- [ ] Keep tie-break deterministic.

Testing tasks:

- [ ] Unit tests for candidate sort stability and expected ordering.
- [ ] Property tests for deterministic outputs given identical inputs.

## Invariant 6: Boundary-Safe Cross-Group Behavior

Schema tasks:

- [ ] Add configurable `boundary_window` per expansion stage.

API tasks:

- [ ] Expose active boundary window in matchmaking metadata.

Service tasks:

- [ ] Gate cross-group eligibility by distance-to-boundary in early/mid stages.
- [ ] Allow full cross-group eligibility only in late stage.

Testing tasks:

- [ ] Integration tests for allowed and denied cross-group candidates by stage.

## Invariant 7: Bounded Mismatch

Schema tasks:

- [ ] Add config for max rating gap per stage.

API tasks:

- [ ] Include applied max-gap in match creation metadata.

Service tasks:

- [ ] Enforce stage-specific max rating-gap hard cap.
- [ ] Apply emergency cap only after minimum wait threshold.

Testing tasks:

- [ ] Unit tests for cap enforcement by stage.
- [ ] Integration tests for emergency behavior after threshold time.

## Invariant 8: Atomic Rating Updates

Schema tasks:

- [ ] Add `rating_events` ledger table keyed by `match_id` and `player_id`.
- [ ] Add idempotency constraint for result reprocessing safety.

API tasks:

- [ ] Add result-processing endpoint contract with idempotency key.

Service tasks:

- [ ] Process match results and rating updates in one DB transaction.
- [ ] Roll back all participant updates on any failure.

Testing tasks:

- [ ] Integration tests for full rollback under induced failure.
- [ ] Idempotency tests for duplicate result submission.

## Invariant 9: Decision Observability

Schema tasks:

- [x] Add `matchmaking_decisions` table with queue age, stage, rating spread, cross-group flag.

API tasks:

- [x] Add admin endpoint to query recent decision records.

Service tasks:

- [x] Emit decision record at match creation time.
- [ ] Log decision ID in structured logs for traceability.

Testing tasks:

- [ ] Integration test ensuring decision record persistence per created match.

## Invariant 10: Progression Transparency

Schema tasks:

- [ ] Add transition reason enum or constrained text values.

API tasks:

- [ ] Extend player progression payload with `rating_before`, `rating_after`, `group_before`, `group_after`, and `reason_code`.

Service tasks:

- [ ] Populate transition reason and threshold context after each rating update.

Testing tasks:

- [ ] Contract tests for progression payload completeness and stable reason codes.

## Cross-Cutting Tasks

Configuration and rollout:

- [ ] Centralize thresholds/ranges in versioned config.
- [ ] Add feature flag for unified matcher rollout.
- [ ] Add per-mode override support only if justified by data.

Telemetry:

- [ ] Emit metrics for queue time, rating spread, cross-group rate, boundary oscillation.
- [ ] Create dashboards aligned to ADR-011 success metrics.

Simulation:

- [ ] Build synthetic population simulator for threshold tuning.
- [ ] Run before/after simulations for any config change.

Docs:

- [ ] Keep `docs/matchmaking/unified-matchmaking-core-concepts.md` in sync with live config.
- [ ] Update ADR follow-ups when implementation milestones land.

## Suggested Delivery Sequence

1. Data and config foundations (`Invariant 1`, `Invariant 3`, `Invariant 4`)
2. Candidate selection and match assembly (`Invariant 2`, `Invariant 5`, `Invariant 6`, `Invariant 7`)
3. Result processing and transparency (`Invariant 8`, `Invariant 10`)
4. Observability and safe tuning (`Invariant 9`, cross-cutting telemetry/simulation)
