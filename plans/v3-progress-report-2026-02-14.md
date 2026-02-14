# Sprocket v3 Progress Report

Date: 2026-02-14
Scope covered: Weeks 1-5 complete, Week 6 planned

## Executive Summary

The project is on track and currently ahead of the original macro pacing for domain depth.
Foundations, auth/RBAC baseline, hierarchy core, and roster ownership semantics are in place with automated tests and local smoke automation.
Remaining work is meaningful but bounded, and is achievable as a hobby project at 10 hrs/week if scope discipline is maintained.

## Achievements So Far

### Week 1: Foundation

- Go service scaffolded with health/readiness endpoints
- DB connectivity and migration runner in place
- Kubernetes baseline manifests + local apply script
- CI pipeline established
- Week 1 onboarding and smoke path created

### Week 2: Auth and RBAC Baseline

- Auth token parsing + principal injection middleware
- RBAC evaluator wired with protected endpoint behavior
- AuthZ schema and test coverage added
- Contributor onboarding for auth/RBAC workflow

### Week 3: League Hierarchy Core (Part 1)

- League, Franchise, Club schema + create/list APIs
- Referential constraints and validation/error mapping
- DB/API integration coverage for happy path + conflicts
- Week 3 onboarding and smoke path

### Week 4: League Hierarchy Core (Part 2)

- Team and Player schema/API delivered
- End-to-end hierarchy chain operational
- Validation + integration tests expanded
- Week 4 onboarding and smoke path

### Week 5: Roster Membership Slice

- Roster membership schema + API delivered
- Duplicate/constraint protections tested
- Ownership semantics consolidated so roster membership is authoritative
- `players.team_id` removed via migration (`000006`)
- ADRs and onboarding updated to the new contract

## Current System Snapshot

- ADRs: 6 (`/Users/jacbaile/Sprocket-v3/docs/adr`)
- Migrations: 6 (`/Users/jacbaile/Sprocket-v3/migrations`)
- Week smoke scripts: 4 (`week1`, `week3`, `week4`, `week5`)
- HTTP v1 routes in service: admin ping + 6 hierarchy/roster routes
- Test cases in `internal/*`: 33 (`func Test...` count)

## Remaining Work

### Product and Platform Roadmap (High Level)

1. Week 6: queue enrollment slice (join/leave/list, FIFO ordering contract)
2. Week 7: match flow MVP entry point (promote queued teams into match candidates)
3. Weeks 8-10: queue + match lifecycle completion, reliability passes, deeper functional coverage
4. Weeks 11-12: submission and ratification lifecycle (state transitions + constraints)
5. Weeks 13-14: handoff hardening (runbooks, contributor docs, CI quality gates, operability polish)

### Technical Debt and Open Decisions

- Roster lifecycle currently active-state only (no transfer/deactivate workflow yet)
- Queue design must remain additive to avoid rework when match orchestration lands
- Operational maturity is baseline-level; hardening is intentionally deferred

## Capacity and Estimate

Assumption:

- Available time: 10 hrs/week gross
- Planned delivery: 8 hrs/week
- Reserve: 2 hrs/week

Estimate from current state:

- Remaining planned delivery work: ~72 hours (9 planned weeks)
- Remaining gross time at 10 hrs/week: ~9 weeks

Range:

- Conservative: 9-10 weeks (normal interruptions + some rework)
- Expected: 8-9 weeks (current delivery pace)
- Aggressive: 7-8 weeks (if scope stays tight and blockers stay low)

## Risk Watchlist

- Scope expansion before queue/match MVP stabilizes
- Hidden complexity in asynchronous match processing
- Handoff quality slipping if docs/tests are postponed late

## Recommendation

Continue with strict vertical slices and keep the current weekly discipline (code + tests + docs + smoke in one pass).
At current velocity, this is achievable and not a pipe dream.
