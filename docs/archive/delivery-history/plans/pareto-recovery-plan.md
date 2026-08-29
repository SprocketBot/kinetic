# Kinetic v3 Pareto Recovery Plan

Date: 2026-02-15  
Owner capacity: 10 hrs/week gross, 8 hrs/week delivery target

## Execution Status

Accelerated execution complete on 2026-02-15 for P1-P6 MVP scope:

- P1 instrumentation and metrics endpoint implemented
- P2 operator inbox + triage/resolve workflow implemented
- P3 scheduling automation rules implemented
- P4 no-show/forfeit automation rules implemented
- P5 replay/dispute triage automation rules implemented
- P6 continuation guardrails and baseline reporting artifacts published

## Problem Statement

Current v3 solves core platform structure and reliability, but it does not yet reduce the highest-cost operational burden: exception-heavy league administration.

The target now is explicit:

- stop optimizing generic platform completeness
- optimize for reduction of weekly human operator time

## Success Metrics (Primary)

Track these as weekly scoreboard metrics:

1. `admin_hours_per_week` (target: down 50% from baseline in 6 weeks)
2. `manual_touches_per_fixture` (target: <= 1.5 median)
3. `zero_touch_fixture_rate` (target: >= 60%)
4. `time_to_close_exception` (target: <= 24h median)

## What Counts As An Exception

Only these categories are in scope:

- scheduling conflict / reschedule request
- no-show / forfeit determination
- result dispute
- replay ingest failure / identity mismatch
- roster eligibility exception near match time

Anything else is out of scope until these metrics improve.

## Product Strategy: Operator Inbox First

Build one opinionated operations surface in backend APIs first:

- single `operator inbox` feed of unresolved exceptions
- machine-assigned `reason_code`, `severity`, `suggested_action`
- deterministic state machine: `open -> triaged -> resolved`
- auditable actor + reason on every manual action

No broad UI build needed initially; APIs + scripts + smoke-driven flows are enough.

## 6-Week Delivery Plan (Post-Week14)

## Week P1: Instrumentation + Baseline Capture

Objective: measure current pain before adding automation.

Deliverables:

- exception event schema and append-only event log
- baseline metrics endpoint/report generation
- backfill script for existing Week 11/12 flows where possible
- `docs/reports/pareto-baseline-YYYY-MM-DD.md`

Acceptance:

- can compute all 4 primary metrics from system data
- baseline report committed

## Week P2: Inbox API + Exception Taxonomy

Objective: centralize unresolved operational work.

Deliverables:

- `GET /v1/operator-inbox`
- `POST /v1/operator-inbox/triage`
- reason-code taxonomy and severity model
- deterministic sorting (oldest high-severity first)

Acceptance:

- inbox lists all open exceptions from one endpoint
- triage action writes audit trail

## Week P3: Scheduling Exception Automation

Objective: reduce time lost to scheduling churn.

Deliverables:

- match scheduling conflict detector rules
- reschedule proposal + ratification workflow for conflicts
- auto-close for straightforward accepted reschedules

Acceptance:

- at least 30% of scheduling exceptions auto-resolve without admin touch

## Week P4: No-show / Forfeit Decision Engine

Objective: remove repetitive adjudication work.

Deliverables:

- no-show evidence capture contract
- rule-based forfeit recommendation engine
- operator one-click accept/override endpoints with audit fields

Acceptance:

- at least 50% of no-show/forfeit cases resolved with one operator action or less

## Week P5: Replay Failure + Dispute Triage

Objective: absorb replay and dispute chaos into deterministic workflows.

Deliverables:

- replay ingest failure classifier (`parse_failed`, `context_mismatch`, `identity_mismatch`)
- dispute ticket state machine with suggested next step
- explicit `needs_human_review` queue integration

Acceptance:

- replay/dispute exceptions appear in inbox with reason + suggested_action
- median time-to-triage for these exceptions <= 12h

## Week P6: Optimization + Kill/Continue Decision

Objective: verify actual time savings and decide continuation.

Deliverables:

- before/after metrics report
- top unresolved friction list with effort estimates
- keep/kill decision memo for next 6 weeks

Acceptance:

- if `admin_hours_per_week` has not improved by >= 30%, pause feature work and redesign exception model

## Engineering Guardrails

- Every new exception flow requires:
  - one API integration test
  - one local smoke assertion
  - one minikube smoke assertion
  - one runbook update for operator usage
- No new domain feature accepted unless it moves one primary metric.
- Prefer rule-based deterministic automation over probabilistic heuristics.

## Weekly Time Budget (10h)

- 6h implementation
- 2h test/smoke hardening
- 1h docs/runbook updates
- 1h metrics review and reprioritization

## Kill Criteria (Avoid More Platform Drift)

Stop or reset the track if any of these occur for 2 consecutive weeks:

- `admin_hours_per_week` flat or worse
- `manual_touches_per_fixture` flat or worse
- backlog grows while inbox closure rate declines

## Immediate Next Actions

1. Add exception event tables and APIs (Week P1 scope only)
2. Capture first baseline report from current data
3. Freeze unrelated feature work until P1 baseline is committed
