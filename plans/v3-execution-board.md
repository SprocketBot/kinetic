# Sprocket v3 Execution Board

Last updated: 2026-02-15
Owner capacity: 10 hrs/week gross, 8 hrs/week planned delivery

## Current Progress

- Completed: `W1-01` through `W1-10` on 2026-02-07
- Week 1 objective achieved and validated (local, DB, and minikube paths)
- Week 2 completed: `W2-01` through `W2-10` on 2026-02-08
- Week 3 in progress: `W3-01`, `W3-02`, and `W3-03` completed on 2026-02-08
- Week 3 in progress: `W3-04`, `W3-05`, `W3-06`, and `W3-07` completed on 2026-02-08
- Week 3 in progress: `W3-08` and `W3-09` completed on 2026-02-08
- Week 3 completed: `W3-10` consumed for end-to-end hierarchy smoke hardening
- Week 4 planning prepared (implementation not started)
- Week 4 started: `W4-01` completed on 2026-02-08
- Week 4 in progress: `W4-02` through `W4-09` completed on 2026-02-08
- Week 4 completed: `W4-10` consumed for week4 smoke reliability hardening (port fallback)
- Week 5 planning prepared (implementation not started)
- Week 5 started: `W5-01` through `W5-03` completed on 2026-02-09
- Week 5 in progress: `W5-04` through `W5-09` completed on 2026-02-09
- Week 5 completed: `W5-10` not consumed (no blockers)
- Week 6 planning prepared (implementation not started)
- Week 6 started: `W6-01` through `W6-03` completed on 2026-02-14
- Week 6 in progress: `W6-04` through `W6-09` completed on 2026-02-14
- Week 6 completed: `W6-10` not consumed (no blockers)
- Week 7 planning prepared (implementation not started)
- Week 7 started: `W7-01` through `W7-03` completed on 2026-02-15
- Week 7 in progress: `W7-04` through `W7-09` completed on 2026-02-15
- Week 7 completed: `W7-10` not consumed (no blockers)
- Week 8 started: `W8-01` through `W8-03` completed on 2026-02-15
- Week 8 in progress: `W8-04` through `W8-09` completed on 2026-02-15
- Week 8 completed: `W8-10` not consumed (no blockers)
- Week 9 planning prepared (implementation not started)
- Week 9 started: `W9-01` completed on 2026-02-14
- Week 9 in progress: `W9-02` through `W9-09` completed on 2026-02-14
- Week 9 completed: `W9-10` not consumed (no blockers)
- Week 10 planning prepared (implementation not started)
- Week 10 started: `W10-01` completed on 2026-02-14
- Week 10 in progress: `W10-02` through `W10-09` completed on 2026-02-14
- Week 10 completed: `W10-10` not consumed (no blockers)
- Week 11 started: `W11-01` completed on 2026-02-15
- Week 11 in progress: `W11-02` through `W11-09` completed on 2026-02-15
- Week 11 completed: `W11-10` not consumed (no blockers)
- Week 12 started: `W12-01` completed on 2026-02-15
- Week 12 in progress: `W12-02` through `W12-09` completed on 2026-02-15
- Week 12 completed: `W12-10` not consumed (no blockers)
- Next up: Week 13 planning/execution (`W13-01` onward)

## Capacity Guardrails

- Weekly budget: 10h
- Delivery target: 8h
- Reserve: 2h (planning, CI/debug, context switching)
- Hard rule: no week is complete unless code + tests + docs are all done
- Verification rule: each week must include both local smoke (`tools/weekX-smoke.sh`) and minikube smoke (`tools/weekX-k8s-smoke.sh` once introduced) for that slice.

## 14-Week Macro Plan

| Weeks | Focus | Planned Delivery Hours | Exit Criteria |
| --- | --- | ---: | --- |
| 1-2 | Foundation | 16h | Go service scaffold + local dev + first K8s deploy |
| 3-4 | Auth + RBAC baseline | 16h | Authenticated API + role checks + tests |
| 5-6 | League hierarchy core | 16h | League/Franchise/Club/Team/Player/roster/queue vertical slices |
| 7 | Scheduled competition model kickoff | 8h | Season -> ScheduleGroup -> Fixture -> Match schema + APIs |
| 8-10 | Queue + scrim flow MVP | 24h | Join/leave/process queue end-to-end |
| 11-12 | Submission + ratification + replay ingestion MVP | 16h | Submission lifecycle, replay evidence pipeline, and state transitions |
| 13-14 | Handoff + hardening | 16h | Contributor-ready docs/runbooks + CI quality gates |

---

## Week 1 Execution Board (2026-02-09 to 2026-02-15)

Week objective: create the v3 platform skeleton with a runnable Go service and Kubernetes deployment path.

Definition of done for Week 1:

- Service builds and runs locally
- Health/readiness endpoints implemented
- Postgres connectivity wired
- Kubernetes manifests (or Helm chart) deploy successfully to local cluster
- CI runs lint + unit tests + build
- Onboarding doc allows a new contributor to run in < 30 minutes

### Day Plan (5 x 2h sessions)

| Day | Time Budget | Work Items | Deliverables |
| --- | ---: | --- | --- |
| Mon | 2h | Finalize repo structure, architecture ADR-001, choose stack libs | `README.md`, `docs/adr/001-architecture.md`, repo skeleton |
| Tue | 2h | Implement Go service bootstrap (config, logger, health/readiness) | `cmd/api/main.go`, `internal/platform/http`, smoke unit tests |
| Wed | 2h | Wire Postgres (migrations + connection checks), add first domain placeholder module | `internal/platform/db`, `migrations/`, DB integration test |
| Thu | 2h | Add Kubernetes baseline (namespace, deployment, service, config/secret handling) | `deploy/k8s/*` or `deploy/helm/*`, deploy script |
| Fri | 2h | CI pipeline + onboarding pass + cleanup buffer | `.github/workflows/ci.yml`, `docs/onboarding/week1.md` |

### Ticket Board

| ID | Task | Estimate | Status | Acceptance Criteria |
| --- | --- | ---: | --- | --- |
| W1-01 | Create repo skeleton for modular monolith | 1.5h | Done | `cmd/`, `internal/`, `pkg/`, `deploy/`, `docs/` present |
| W1-02 | Add ADR for architecture and migration approach | 0.5h | Done | ADR merged with rationale + tradeoffs |
| W1-03 | Implement HTTP server with health endpoints | 1.0h | Done | `/healthz` and `/readyz` return 200 in local run |
| W1-04 | Add config loading + structured logging | 1.0h | Done | env-config works; logs are structured JSON |
| W1-05 | Add DB bootstrap + migration runner | 1.5h | Done | app starts with DB; migration command is repeatable |
| W1-06 | Write baseline tests (unit + integration smoke) | 1.0h | Done | tests pass locally in one command |
| W1-07 | Add Kubernetes deployment baseline | 1.0h | Done | deploy to local cluster succeeds |
| W1-08 | Add CI workflow | 0.75h | Done | CI runs fmt/lint/test/build on PR |
| W1-09 | Write onboarding runbook | 0.75h | Done | fresh clone to running app in <= 30 minutes |
| W1-10 | Risk/overflow buffer | 1.0h | Done | consumed for local k8s image/debug + DB smoke verification |

### Week 1 Risks and Mitigations

- Risk: toolchain churn (Go/K8s/bootstrap details) burns time.
  - Mitigation: lock minimal versions in docs and avoid optional tooling.
- Risk: deployment complexity exceeds a 2h session.
  - Mitigation: start with plain manifests before Helm.
- Risk: test harness setup stalls.
  - Mitigation: ship small smoke tests first, expand next week.

---

## Weekly Template (for Weeks 2+)

Copy this section per week.

### Week X Objective

- One-sentence objective

### Week X Definition of Done

- Behavior shipped
- Unit + integration/functional tests shipped
- Local smoke script and minikube smoke script pass for the week slice
- Docs/runbook updated

### Week X Day Plan (2h sessions)

| Day | Time Budget | Work Items | Deliverables |
| --- | ---: | --- | --- |
| Mon | 2h |  |  |
| Tue | 2h |  |  |
| Wed | 2h |  |  |
| Thu | 2h |  |  |
| Fri | 2h |  |  |

### Week X Ticket Board

| ID | Task | Estimate | Status | Acceptance Criteria |
| --- | --- | ---: | --- | --- |
| WX-01 |  |  | Todo |  |

### Week X Start Checklist

- Previous week local smoke passes (`./tools/week{X-1}-smoke.sh`)
- Previous week minikube smoke passes (`./tools/week{X-1}-k8s-smoke.sh` once available)
- Full test suite passes: `go test ./...`
- Minikube context selected: `kubectl config use-context minikube`
- Minikube cluster reachable: `minikube status`

---

## Week 2 Execution Board (2026-02-16 to 2026-02-22)

Week objective: establish auth and RBAC foundation primitives (identity + policy enforcement skeleton) without over-building product features.

Definition of done for Week 2:

- Auth domain model and DB schema baseline merged
- Request authentication middleware in place (token parsing + principal context)
- RBAC policy model wired with at least one protected endpoint
- Unit + integration tests covering allow/deny behavior paths
- Updated contributor docs for auth/RBAC local testing

### Scope Boundaries

In scope:

- auth primitives and enforcement scaffolding
- one reference protected endpoint
- test harness for permission checks

Out of scope:

- full user management UI
- full role management CRUD
- external OAuth provider integrations

### Day Plan (5 x 2h sessions)

| Day | Time Budget | Work Items | Deliverables |
| --- | ---: | --- | --- |
| Mon | 2h | Finalize auth/RBAC ADR notes, define DB entities and migration plan | `docs/adr/002-auth-rbac-baseline.md` (or update ADR-001), ticket-ready schema notes |
| Tue | 2h | Implement auth context extraction + token validation interface | `internal/platform/auth/*`, request principal plumbing |
| Wed | 2h | Implement RBAC evaluator + seed minimal roles/policies | `internal/domain/authz/*`, migration/seed artifacts |
| Thu | 2h | Add protected endpoint and integration tests (allow + deny) | reference route + authz integration tests |
| Fri | 2h | Docs/runbook updates + cleanup + buffer | `docs/onboarding/week2.md`, updated test commands |

### Ticket Board

| ID | Task | Estimate | Status | Acceptance Criteria |
| --- | --- | ---: | --- | --- |
| W2-01 | Define auth/RBAC baseline contract | 0.75h | Done | clear token/principal/permission model documented |
| W2-02 | Add auth package skeleton | 0.75h | Done | package compiles, interfaces defined, no dead code |
| W2-03 | Implement token parsing and principal injection middleware | 1.25h | Done | request context includes principal or explicit anonymous state |
| W2-04 | Add DB entities/migration for role + policy baseline | 1.25h | Done | migration is repeatable and tested |
| W2-05 | Implement RBAC evaluator service | 1.25h | Done | evaluator supports allow/deny checks by action/resource |
| W2-06 | Wire one protected endpoint | 0.75h | Done | endpoint rejects unauthorized and allows authorized principal |
| W2-07 | Write authz unit tests | 0.75h | Done | policy evaluation edge cases covered |
| W2-08 | Write authz integration tests | 0.75h | Done | end-to-end allow/deny behavior validated |
| W2-09 | Write Week 2 onboarding notes | 0.5h | Done | new contributor can run auth tests quickly |
| W2-10 | Risk/overflow buffer | 1.0h | Done | consumed for DB-backed authz policy loading + fallback hardening |

### Week 2 Risks and Mitigations

- Risk: token format decision churn delays implementation.
  - Mitigation: support one simple internal token format first; defer external providers.
- Risk: RBAC model over-design.
  - Mitigation: enforce only one concrete action/resource pair in Week 2.
- Risk: integration tests become flaky.
  - Mitigation: keep fixtures small and deterministic; avoid network dependencies.

### Week 2 Start Checklist

- Week 1 smoke script passes: `./tools/week1-smoke.sh`
- Minikube context set if testing deploy path: `kubectl config use-context minikube`
- Clean test DB port available (default `55432`)
- No open Week 1 blockers

---

## Week 3 Execution Board (2026-02-23 to 2026-03-01)

Week objective: deliver the first League hierarchy domain slice (League, Franchise, Club) with DB-backed persistence, API endpoints, and behavior-focused tests.

Definition of done for Week 3:

- `League`, `Franchise`, and `Club` schema/models exist and are migrated
- Create/list API endpoints implemented for this hierarchy slice
- Referential integrity enforced in DB (FK + uniqueness constraints)
- Unit + integration tests cover happy path and key constraint failures
- Week 3 onboarding notes document how to exercise hierarchy endpoints

### Scope Boundaries

In scope:

- data model + migrations for league hierarchy core (League, Franchise, Club)
- minimal API surface: create + list for each entity
- validation and integrity constraints
- tests and docs

Out of scope:

- `Team` and `Player` modeling (defer to Week 4)
- advanced filtering/pagination
- UI work
- RBAC enforcement expansion beyond Week 2 baseline

### Day Plan (5 x 2h sessions)

| Day | Time Budget | Work Items | Deliverables |
| --- | ---: | --- | --- |
| Mon | 2h | Finalize hierarchy contract and migration design | `docs/adr/003-league-hierarchy-slice.md`, migration plan |
| Tue | 2h | Implement migrations + storage layer for League/Franchise/Club | `migrations/000003_*.sql`, repositories/store code |
| Wed | 2h | Implement create/list API handlers and request validation | `/v1/leagues`, `/v1/franchises`, `/v1/clubs` endpoints |
| Thu | 2h | Integration tests for hierarchy behavior + constraints | allow/create/list tests + FK violation coverage |
| Fri | 2h | Docs/onboarding updates + cleanup + buffer | `docs/onboarding/week3.md`, command examples |

### Ticket Board

| ID | Task | Estimate | Status | Acceptance Criteria |
| --- | --- | ---: | --- | --- |
| W3-01 | Define hierarchy slice contract | 0.75h | Done | League/Franchise/Club fields and invariants documented |
| W3-02 | Add migrations for League/Franchise/Club | 1.25h | Done | schema migrates cleanly and is idempotent |
| W3-03 | Add domain models/store interfaces | 1.0h | Done | compile-safe model and store boundaries in place |
| W3-04 | Implement DB store methods | 1.25h | Done | create/list operations work for all 3 entities |
| W3-05 | Implement API handlers/routes | 1.25h | Done | create/list endpoints return JSON contracts |
| W3-06 | Add request validation and error mapping | 0.75h | Done | invalid payloads return stable 4xx errors |
| W3-07 | Write unit tests for model/validation logic | 0.75h | Done | edge cases covered for required fields/uniqueness assumptions |
| W3-08 | Write integration tests for DB/API behavior | 1.0h | Done | FK + duplicate constraints validated end-to-end |
| W3-09 | Write Week 3 onboarding notes | 0.5h | Done | new contributor can run hierarchy checks quickly |
| W3-10 | Risk/overflow buffer | 0.75h | Done | consumed for week3 smoke script and runtime hardening |

### Week 3 Risks and Mitigations

- Risk: schema churn spills into Week 4.
  - Mitigation: lock Week 3 to 3 entities only; defer Team/Player.
- Risk: endpoint contract uncertainty.
  - Mitigation: publish ADR/API examples before coding handlers.
- Risk: slow integration tests.
  - Mitigation: keep fixtures tiny and avoid unnecessary seed volume.

### Week 3 Start Checklist

- Week 2 auth/RBAC baseline tests pass: `go test ./...`
- Week 1 smoke script passes: `./tools/week1-smoke.sh`
- Postgres test port available (`55432` default)
- No open blockers on migrations or API runtime bootstrap

---

## Week 4 Execution Board (2026-03-02 to 2026-03-08)

Week objective: extend the hierarchy slice by adding `Team` and `Player` models with relational integrity, plus create/list API endpoints and behavior-focused tests.

Definition of done for Week 4:

- `Team` and `Player` schema/models exist and are migrated
- Create/list endpoints implemented for Team and Player
- FK and uniqueness constraints enforced and tested
- Integration tests validate hierarchy linkage from League -> Franchise -> Club -> Team -> Player
- Week 4 onboarding notes explain end-to-end usage and checks

### Scope Boundaries

In scope:

- data model + migrations for Team and Player entities
- create + list API endpoints for Team and Player
- validation/error mapping for these endpoints
- integration tests for dependency/constraint behavior

Out of scope:

- roster assignment workflows and transfer logic
- update/delete endpoints
- pagination/filtering expansions
- role-scoped authz for hierarchy mutations (defer)

### Day Plan (5 x 2h sessions)

| Day | Time Budget | Work Items | Deliverables |
| --- | ---: | --- | --- |
| Mon | 2h | Finalize Team/Player contract and migration design | `docs/adr/004-team-player-slice.md`, migration notes |
| Tue | 2h | Implement Team/Player migrations + domain model updates | `migrations/000004_*.sql`, `internal/domain/hierarchy/*` updates |
| Wed | 2h | Implement DB store methods and API routes for Team/Player | store + `/v1/teams`, `/v1/players` |
| Thu | 2h | Integration tests for Team/Player constraints and hierarchy links | FK/duplicate/validation API tests |
| Fri | 2h | Onboarding docs + smoke updates + cleanup/buffer | `docs/onboarding/week4.md`, `tools/week4-smoke.sh` |

### Ticket Board

| ID | Task | Estimate | Status | Acceptance Criteria |
| --- | --- | ---: | --- | --- |
| W4-01 | Define Team/Player slice contract | 0.75h | Done | fields + invariants documented in ADR |
| W4-02 | Add migrations for Team/Player | 1.25h | Done | schema migrates cleanly and is idempotent |
| W4-03 | Extend hierarchy domain models/store interface | 0.75h | Done | compile-safe model changes in place |
| W4-04 | Implement DB store methods for Team/Player | 1.25h | Done | create/list for Team/Player works |
| W4-05 | Implement API handlers/routes for Team/Player | 1.0h | Done | create/list endpoints return stable JSON |
| W4-06 | Add validation + error mapping for Team/Player payloads | 0.75h | Done | invalid payloads return stable 4xx responses |
| W4-07 | Write unit tests for Team/Player validation | 0.5h | Done | slug/required field checks covered |
| W4-08 | Write integration tests for DB/API Team/Player behavior | 1.25h | Done | FK + duplicate constraints validated end-to-end |
| W4-09 | Write Week 4 onboarding notes | 0.5h | Done | contributor can run Team/Player checks quickly |
| W4-10 | Risk/overflow buffer | 1.0h | Done | consumed for week4 smoke hardening and port fallback handling |

### Week 4 Risks and Mitigations

- Risk: Team/Player schema assumptions conflict with future roster model.
  - Mitigation: keep fields minimal and additive; avoid embedding roster workflows now.
- Risk: integration tests become too coupled to specific IDs.
  - Mitigation: use generated unique suffixes and avoid static fixture IDs.
- Risk: API contract drift between docs and runtime.
  - Mitigation: update onboarding examples and smoke script in same PR as endpoints.

### Week 4 Start Checklist

- Week 3 smoke script passes: `./tools/week3-smoke.sh`
- Full test suite passes: `go test ./...`
- Postgres test port available (`55432` default)
- Working tree clean before first Week 4 implementation commit

---

## Week 5 Execution Board (2026-03-09 to 2026-03-15)

Week objective: introduce roster assignment primitives and workflow endpoints so players can be explicitly attached to teams with auditable membership state.

Definition of done for Week 5:

- Roster membership schema exists and is migrated
- Create/list endpoints for roster membership are implemented
- Membership constraints enforced (no duplicate active membership for same player/team)
- Integration tests cover assignment, duplicate prevention, and missing-dependency cases
- Week 5 onboarding notes and smoke automation are available

### Scope Boundaries

In scope:

- roster membership data model (player <-> team relationship)
- assignment/list API endpoints
- constraint/error handling and tests
- onboarding docs and smoke script

Out of scope:

- transfer approval workflows
- historical contract/season modeling
- bulk roster operations
- advanced RBAC scopes for roster operations

### Day Plan (5 x 2h sessions)

| Day | Time Budget | Work Items | Deliverables |
| --- | ---: | --- | --- |
| Mon | 2h | Finalize roster contract and migration design | `docs/adr/005-roster-membership-slice.md`, migration notes |
| Tue | 2h | Implement migration + domain/store interfaces for roster membership | `migrations/000005_*.sql`, domain/store updates |
| Wed | 2h | Implement DB methods and API routes for roster assignment/listing | `/v1/roster-memberships` endpoints |
| Thu | 2h | Integration tests for assignment constraints and dependency errors | API and DB integration tests |
| Fri | 2h | Onboarding + smoke automation + cleanup/buffer | `docs/onboarding/week5.md`, `tools/week5-smoke.sh` |

### Ticket Board

| ID | Task | Estimate | Status | Acceptance Criteria |
| --- | --- | ---: | --- | --- |
| W5-01 | Define roster membership contract | 0.75h | Done | fields + invariants documented in ADR |
| W5-02 | Add migration for roster membership schema | 1.25h | Done | migration is idempotent and applies cleanly |
| W5-03 | Extend domain/store interface for roster membership | 0.75h | Done | compile-safe contracts in place |
| W5-04 | Implement DB methods for assignment/listing | 1.25h | Done | create/list works with constraint mapping |
| W5-05 | Implement API routes for roster memberships | 1.0h | Done | create/list endpoints return stable JSON |
| W5-06 | Add validation and error mapping for roster payloads | 0.75h | Done | invalid payloads map to stable 4xx |
| W5-07 | Write unit tests for roster validation rules | 0.5h | Done | required field and format checks covered |
| W5-08 | Write integration tests for roster constraints | 1.25h | Done | duplicate active membership and FK errors validated |
| W5-09 | Write Week 5 onboarding notes | 0.5h | Done | contributor can run roster checks quickly |
| W5-10 | Risk/overflow buffer | 1.0h | Done | not consumed (no blockers) |

### Week 5 Risks and Mitigations

- Risk: ambiguity between player-team ownership and roster history.
  - Mitigation: keep Week 5 to current-state membership only; add history later.
- Risk: duplicate membership edge cases.
  - Mitigation: enforce unique active membership constraints at DB level.
- Risk: API semantics drift from future transfer workflows.
  - Mitigation: use neutral naming (`roster-membership`) and additive design.

### Week 5 Start Checklist

- Week 4 smoke script passes: `./tools/week4-smoke.sh`
- Full test suite passes: `go test ./...`
- Postgres test port available (`55432` default)
- Working tree clean before first Week 5 implementation commit

---

## Week 6 Execution Board (2026-03-16 to 2026-03-22)

Week objective: begin queue + scrim flow MVP by shipping deterministic queue enrollment behavior (join/leave/list) for teams.

Definition of done for Week 6:

- Queue and queue entry schema exists and is migrated
- Join/leave/list queue endpoints are implemented
- Duplicate active queue entries are prevented at DB level
- Integration tests cover queue ordering, duplicate prevention, and dependency violations
- Week 6 onboarding notes and smoke automation are available

### Scope Boundaries

In scope:

- queue primitives (`Queue`, `QueueEntry`)
- API behavior for queue join/leave/list
- FIFO ordering contract for active entries
- tests and onboarding/smoke docs

Out of scope:

- full match creation pipeline
- MMR/skill balancing
- backfill and timeout processing workers
- season-specific queue policy engine

### Day Plan (5 x 2h sessions)

| Day | Time Budget | Work Items | Deliverables |
| --- | ---: | --- | --- |
| Mon | 2h | Finalize queue contract and constraints | `docs/adr/007-queue-enrollment-slice.md`, migration notes |
| Tue | 2h | Implement queue migrations + domain/store interfaces | `migrations/000007_*.sql`, `internal/domain/hierarchy/*` updates |
| Wed | 2h | Implement DB methods and API routes for queue join/leave/list | `/v1/queues`, `/v1/queue-entries` |
| Thu | 2h | Integration tests for ordering and constraints | DB/API integration tests for duplicates/FK/order |
| Fri | 2h | Onboarding docs + smoke script + cleanup/buffer | `docs/onboarding/week6.md`, `tools/week6-smoke.sh` |

### Ticket Board

| ID | Task | Estimate | Status | Acceptance Criteria |
| --- | --- | ---: | --- | --- |
| W6-01 | Define queue enrollment contract | 0.75h | Done | fields + invariants documented in ADR |
| W6-02 | Add queue + queue entry migrations | 1.25h | Done | schema migrates cleanly and is idempotent |
| W6-03 | Extend domain/store interfaces for queue behavior | 0.75h | Done | compile-safe contracts in place |
| W6-04 | Implement DB methods for join/leave/list | 1.25h | Done | active queue operations map errors correctly |
| W6-05 | Implement API handlers/routes for queue operations | 1.25h | Done | stable JSON and status codes for join/leave/list |
| W6-06 | Add validation and error mapping for queue payloads | 0.75h | Done | invalid payloads return stable 4xx |
| W6-07 | Write unit tests for queue validation + ordering helper | 0.75h | Done | required fields and FIFO ordering checks covered |
| W6-08 | Write integration tests for queue constraints | 1.0h | Done | duplicate active entries and missing dependencies validated |
| W6-09 | Write Week 6 onboarding notes + smoke automation | 0.5h | Done | contributor can run queue checks quickly |
| W6-10 | Risk/overflow buffer | 1.0h | Done | not consumed (no blockers) |

### Week 6 Risks and Mitigations

- Risk: queue API shape may conflict with later match orchestration.
  - Mitigation: keep Week 6 contract minimal and additive; defer orchestration concerns.
- Risk: ordering bugs create hidden fairness issues.
  - Mitigation: enforce deterministic FIFO (`created_at`, then `id`) in DB queries and tests.
- Risk: drift between team readiness and queue state assumptions.
  - Mitigation: Week 6 keeps readiness out of scope; add explicit readiness checks in Week 7+.

### Week 6 Start Checklist

- Week 5 smoke script passes: `./tools/week5-smoke.sh`
- Full test suite passes: `go test ./...`
- Postgres test port available (`55432` default)
- Working tree clean before first Week 6 implementation commit

---

## Week 7 Execution Board (2026-03-23 to 2026-03-29)

Week objective: introduce scheduled competition primitives (`season -> schedule_group -> fixture -> match`) as a distinct domain from scrims.

Definition of done for Week 7:

- Scheduled competition schema exists and is migrated
- Create/list endpoints exist for season, schedule groups, fixtures, and matches
- Match lifecycle includes explicit scheduling metadata and ratification state
- Validation/error handling and integration tests cover hierarchy dependencies and lifecycle constraints
- Week 7 onboarding notes and smoke automation are available

### Scope Boundaries

In scope:

- scheduled hierarchy entities and constraints
- match scheduling fields (`scheduled_for`, ratification timestamps/status)
- create/list APIs for core scheduled entities
- tests and onboarding/smoke docs

Out of scope:

- automatic schedule generation
- standings computation
- full dispute workflow UI
- scrim-to-match convergence logic

### Day Plan (5 x 2h sessions)

| Day | Time Budget | Work Items | Deliverables |
| --- | ---: | --- | --- |
| Mon | 2h | Finalize scheduled competition contract + lifecycle | `docs/adr/009-scheduled-competition-slice.md` |
| Tue | 2h | Implement migrations for season/schedule_group/fixture/match | `migrations/000008_*.sql` |
| Wed | 2h | Implement domain/store + API create/list routes | `/v1/seasons`, `/v1/schedule-groups`, `/v1/fixtures`, `/v1/matches` |
| Thu | 2h | Integration tests for hierarchy and lifecycle constraints | DB/API integration coverage |
| Fri | 2h | Onboarding docs + smoke script + cleanup/buffer | `docs/onboarding/week7.md`, `tools/week7-smoke.sh` |

### Ticket Board

| ID | Task | Estimate | Status | Acceptance Criteria |
| --- | --- | ---: | --- | --- |
| W7-01 | Define scheduled competition slice contract | 0.75h | Done | hierarchy + lifecycle documented in ADR |
| W7-02 | Add migrations for season/schedule_group/fixture/match | 1.5h | Done | schema migrates cleanly and is idempotent |
| W7-03 | Extend domain/store interfaces for scheduled entities | 0.75h | Done | compile-safe contracts in place |
| W7-04 | Implement DB methods for create/list scheduled entities | 1.25h | Done | create/list operations work with dependency mapping |
| W7-05 | Implement API handlers/routes for scheduled entities | 1.25h | Done | stable JSON and status codes for create/list |
| W7-06 | Add validation + lifecycle guardrails for matches | 0.75h | Done | `ready` requires ratified agreed play time |
| W7-07 | Write unit tests for scheduled validation rules | 0.75h | Done | required fields + lifecycle checks covered |
| W7-08 | Write integration tests for hierarchy/lifecycle constraints | 1.0h | Done | FK + lifecycle violations validated end-to-end |
| W7-09 | Write Week 7 onboarding notes + smoke automation | 0.5h | Done | contributor can run scheduled checks quickly |
| W7-10 | Risk/overflow buffer | 1.0h | Done | not consumed (no blockers) |

### Week 7 Risks and Mitigations

- Risk: scheduled vs scrim domain boundaries drift.
  - Mitigation: enforce naming and ADR-defined ownership in APIs and docs.
- Risk: lifecycle ambiguity around scheduling and readiness.
  - Mitigation: define concrete fields and transitions where `ready` requires ratified schedule time.
- Risk: schema complexity exceeds weekly time budget.
  - Mitigation: scope to create/list plus lifecycle invariants only; defer scheduling generation.

### Week 7 Start Checklist

- Week 6 smoke script passes: `./tools/week6-smoke.sh`
- Full test suite passes: `go test ./...`
- Postgres test port available (`55432` default)
- Working tree clean before first Week 7 implementation commit

---

## Week 8 Execution Board (2026-03-30 to 2026-04-05)

Week objective: establish unified matchmaking foundations (ADR-010/011) while shipping scrim orchestration baseline paths that remain compatible with future replay association (ADR-012/013).

Definition of done for Week 8:

- Scrim schema and baseline lifecycle (`created`, `in_progress`, `closed`, `voided`) exist and are migrated
- Unified rating identity foundation exists (`player_ratings` baseline schema + read path)
- Queue attempt expansion-stage tracking foundation exists (monotonic-stage baseline data model)
- Queue-to-scrim promotion deactivates consumed queue entries atomically and records decision metadata
- Integration tests cover foundational invariants and queue-to-scrim conflict/dependency failures
- Week 8 onboarding notes and both smoke automations are available (`week8-smoke.sh` and `week8-k8s-smoke.sh`)

### Scope Boundaries

In scope:

- scrim entity and lifecycle baseline (`created`, `in_progress`, `closed`, `voided`)
- unified matchmaking invariant foundations from ADR-011:
  - single rating identity baseline
  - monotonic expansion stage tracking baseline
  - decision observability baseline
- promotion workflow from active queue entries to scrim
- deterministic pairing strategy for Week 8 baseline
- tests and onboarding/smoke docs

Out of scope:

- full asynchronous matchmaker worker orchestration
- full cross-group rating-first candidate selection
- production rating update pipeline
- scheduled match integration
- dispute/ratification flow for scrims
- replay upload/parsing implementation (reserved for ADR-012/013 slices)

### Day Plan (5 x 2h sessions)

| Day | Time Budget | Work Items | Deliverables |
| --- | ---: | --- | --- |
| Mon | 2h | Lock Week 8 invariant mapping against ADR-010/011 and replay boundaries (ADR-012/013) | `docs/matchmaking/unified-matchmaking-implementation-checklist.md` (Week 8 mapping notes) |
| Tue | 2h | Implement schema foundations (scrims, player ratings baseline, queue attempt stage + decision metadata) | `migrations/000009_*.sql` |
| Wed | 2h | Implement DB methods/API routes for scrim create/list/promotion + rating snapshot read path | `/v1/scrims`, `/v1/scrim-promotions`, `/v1/player-ratings` |
| Thu | 2h | Integration tests for invariant foundations and queue consumption/conflicts | DB/API integration test coverage |
| Fri | 2h | Onboarding docs + smoke scripts + cleanup/buffer | `docs/onboarding/week8.md`, `tools/week8-smoke.sh`, `tools/week8-k8s-smoke.sh` |

### Ticket Board

| ID | Task | Estimate | Status | Acceptance Criteria |
| --- | --- | ---: | --- | --- |
| W8-01 | Map Week 8 slice to ADR-010/011 invariants and ADR-012/013 boundaries | 0.75h | Done | Week 8 tasks linked to invariant checklist with explicit replay boundary notes |
| W8-02 | Add migration for scrim + rating identity + queue stage/decision metadata foundations | 1.25h | Done | schema migrates cleanly and is idempotent |
| W8-03 | Extend domain/store interfaces for scrim and rating snapshot workflows | 0.75h | Done | compile-safe contracts in place |
| W8-04 | Implement DB methods for scrim create/list/promotion + decision metadata writes | 1.25h | Done | promotion is transactional and observability fields persist |
| W8-05 | Implement API handlers/routes for scrim workflows and rating snapshot reads | 1.25h | Done | stable JSON/status codes for create/list/promote/read |
| W8-06 | Add validation and error mapping for scrim/rating payloads | 0.75h | Done | invalid payloads map to stable 4xx |
| W8-07 | Write unit tests for foundational invariants | 0.75h | Done | single identity + monotonic stage + lifecycle guard checks covered |
| W8-08 | Write integration tests for queue-to-scrim + invariant persistence behavior | 1.0h | Done | duplicate/insufficient queue entry and decision metadata cases validated |
| W8-09 | Write Week 8 onboarding notes + smoke automations | 0.5h | Done | contributor can run local and minikube scrim checks quickly |
| W8-10 | Risk/overflow buffer | 1.0h | Done | not consumed (no blockers) |

### Week 8 Risks and Mitigations

- Risk: scrim promotion semantics overlap with scheduled match semantics.
  - Mitigation: keep Week 8 strictly in scrim domain and preserve ADR-008 terminology boundaries.
- Risk: ADR-010/011 scope overflows weekly budget if fully attempted.
  - Mitigation: limit Week 8 to invariant foundations only (identity, stage tracking, observability), defer advanced matching/rating updates.
- Risk: pairing logic becomes hard to reason about for contributors.
  - Mitigation: ship deterministic baseline pairing first; defer advanced matchmaking.
- Risk: queue entry consumption race conditions.
  - Mitigation: use transactional promotion path and explicit integration tests for contention/conflict cases.
- Risk: replay-association concerns bleed into scrim orchestration implementation.
  - Mitigation: keep ADR-012/013 as explicit out-of-scope for Week 8, but preserve context/evidence linkage hooks only.

### Week 8 Start Checklist

- Week 7 smoke script passes: `./tools/week7-smoke.sh`
- Week 7 minikube smoke script passes when available: `./tools/week7-k8s-smoke.sh`
- Full test suite passes: `go test ./...`
- Postgres test port available (`55432` default)
- Minikube context selected and cluster reachable (`kubectl config use-context minikube`, `minikube status`)
- Working tree clean before first Week 8 implementation commit

---

## Week 9 Execution Board (2026-04-06 to 2026-04-12)

Week objective: deliver deterministic queue processing behavior for scrim creation (rating-first + queue-age tie-break) and explicit scrim lifecycle transitions (`created -> in_progress -> closed|voided`).

Definition of done for Week 9:

- Deterministic candidate ordering is implemented for queue-to-scrim promotion (rating-first with deterministic tie-breaks)
- Team rating snapshot derivation path exists for matchmaking decisions (baseline aggregate model)
- Scrim lifecycle transition endpoint(s) enforce legal state transitions
- Matchmaking decision metadata captures ordering rationale fields needed for debugging
- Integration tests cover ordering behavior and invalid lifecycle transitions
- Week 9 onboarding notes and both smoke automations are available (`week9-smoke.sh` and `week9-k8s-smoke.sh`)

### Scope Boundaries

In scope:

- deterministic promotion candidate ordering implementation
- baseline team rating derivation for promotion decisions
- scrim lifecycle transitions and validation
- decision observability enrichment for ordering rationale
- tests and onboarding/smoke docs

Out of scope:

- async queue worker/scheduler infrastructure
- full rating update pipeline post-result submission
- cross-mode tuning/config simulation tooling
- scheduled-match to scrim convergence

### Day Plan (5 x 2h sessions)

| Day | Time Budget | Work Items | Deliverables |
| --- | ---: | --- | --- |
| Mon | 2h | Finalize Week 9 ordering + lifecycle contract and invariants | `docs/adr/014-scrim-promotion-ordering-and-lifecycle.md` |
| Tue | 2h | Implement store-layer ordering logic and decision metadata enrichment | DB/store updates + migration adjustments if required |
| Wed | 2h | Implement scrim lifecycle transition API/routes + validation | `/v1/scrims/{id}/state` (or equivalent contract) |
| Thu | 2h | Integration tests for deterministic ordering + lifecycle guardrails | DB/API integration coverage |
| Fri | 2h | Onboarding docs + local/k8s smoke scripts + cleanup/buffer | `docs/onboarding/week9.md`, `tools/week9-smoke.sh`, `tools/week9-k8s-smoke.sh` |

### Ticket Board

| ID | Task | Estimate | Status | Acceptance Criteria |
| --- | --- | ---: | --- | --- |
| W9-01 | Define deterministic ordering and scrim lifecycle transition contract | 0.75h | Done | ordering fields, tie-breaks, and legal transitions documented |
| W9-02 | Extend persistence model for ordering rationale fields (if needed) | 1.0h | Done | schema remains idempotent and backward-compatible |
| W9-03 | Implement rating-first deterministic candidate ordering in promotion path | 1.25h | Done | ordering is deterministic and stable under equal ratings |
| W9-04 | Implement team rating snapshot derivation for promotion decisions | 0.75h | Done | promotion path computes consistent team rating baseline |
| W9-05 | Implement scrim lifecycle transition endpoint(s) and handlers | 1.25h | Done | only legal state transitions succeed |
| W9-06 | Add validation/error mapping for transition and promotion edge cases | 0.75h | Done | invalid transitions/payloads map to stable 4xx |
| W9-07 | Write unit tests for ordering and lifecycle validation logic | 0.75h | Done | deterministic tie-break + transition matrix covered |
| W9-08 | Write integration tests for ordering outcomes and lifecycle constraints | 1.0h | Done | expected pairing and state transitions validated end-to-end |
| W9-09 | Write Week 9 onboarding notes + local/k8s smoke automations | 0.5h | Done | contributor can run week slice in local and minikube paths |
| W9-10 | Risk/overflow buffer | 1.0h | Done | not consumed (no blockers) |

### Week 9 Risks and Mitigations

- Risk: ordering policy churn causes rework mid-week.
  - Mitigation: lock deterministic tie-break hierarchy in ADR before coding.
- Risk: lifecycle transitions conflict with future dispute/submission model.
  - Mitigation: keep transition contract minimal and additive; avoid embedding result semantics.
- Risk: rating snapshot derivation is expensive or ambiguous.
  - Mitigation: ship simple baseline aggregation now; defer optimization and advanced weighting.
- Risk: k8s smoke becomes flaky due to environment setup.
  - Mitigation: reuse week8-k8s approach (ephemeral in-cluster Postgres + explicit preflight checks).

### Week 9 Start Checklist

- Week 8 smoke script passes: `./tools/week8-smoke.sh`
- Week 8 minikube smoke script passes: `./tools/week8-k8s-smoke.sh`
- Full test suite passes: `go test ./...`
- Postgres test port available (`55432` default)
- Minikube context selected and cluster reachable (`kubectl config use-context minikube`, `minikube status`)
- Working tree clean before first Week 9 implementation commit

---

## Week 10 Execution Board (2026-04-13 to 2026-04-19)

Week objective: finish queue + scrim flow MVP by adding a deterministic promotion processing path and operational controls to run it safely and observably.

Definition of done for Week 10:

- Promotion orchestration endpoint/job path exists and can process eligible queues without manual SQL intervention
- Promotion processing is idempotent and safe under concurrent invocations
- Promotion run summary/metrics are emitted in API response and logs
- Integration tests cover re-run idempotency and concurrency conflict handling
- Week 10 onboarding notes and both smoke automations are available (`week10-smoke.sh` and `week10-k8s-smoke.sh`)

### Scope Boundaries

In scope:

- synchronous promotion processing endpoint for MVP operations
- idempotency and conflict handling around repeated processing attempts
- operator-visible processing summary output (processed queues, promotions created, conflicts)
- tests and onboarding/smoke docs

Out of scope:

- fully asynchronous scheduler/worker framework
- distributed queue partitioning strategy
- advanced rate limiting and autoscaling policy tuning
- replay/result submission lifecycle work (starts Week 11)

### Day Plan (5 x 2h sessions)

| Day | Time Budget | Work Items | Deliverables |
| --- | ---: | --- | --- |
| Mon | 2h | Define Week 10 processing contract and idempotency semantics | `docs/adr/015-queue-promotion-processing-mvp.md` |
| Tue | 2h | Implement promotion processing service/store path with safety guards | service/store code updates |
| Wed | 2h | Add API endpoint for processing queues + structured run summary | `/v1/scrim-promotions/process` (or equivalent) |
| Thu | 2h | Integration tests for idempotency/concurrency and error mapping | DB/API integration coverage |
| Fri | 2h | Onboarding docs + local/k8s smoke scripts + cleanup/buffer | `docs/onboarding/week10.md`, `tools/week10-smoke.sh`, `tools/week10-k8s-smoke.sh` |

### Ticket Board

| ID | Task | Estimate | Status | Acceptance Criteria |
| --- | --- | ---: | --- | --- |
| W10-01 | Define queue promotion processing contract and idempotency semantics | 0.75h | Done | processing trigger, safety constraints, and response contract documented |
| W10-02 | Add persistence hooks for processing run metadata (if needed) | 0.75h | Done | schema remains additive and idempotent |
| W10-03 | Implement processing service to evaluate/promote eligible queues deterministically | 1.25h | Done | eligible queues produce deterministic promotions |
| W10-04 | Add concurrency/idempotency guards for repeated processing calls | 1.0h | Done | duplicate invocations do not duplicate promotions |
| W10-05 | Implement API endpoint/handler for promotion processing trigger | 1.0h | Done | endpoint returns stable processing summary payload |
| W10-06 | Add validation and error mapping for processing requests | 0.75h | Done | malformed/unsupported requests return stable 4xx |
| W10-07 | Write unit tests for processing summary and safety logic | 0.75h | Done | idempotency and conflict paths covered |
| W10-08 | Write integration tests for concurrent/duplicate processing behavior | 1.0h | Done | no duplicate promotion side effects under repeated calls |
| W10-09 | Write Week 10 onboarding notes + local/k8s smoke automations | 0.5h | Done | contributor can run week slice in local and minikube paths |
| W10-10 | Risk/overflow buffer | 1.25h | Done | not consumed (no blockers) |

### Week 10 Risks and Mitigations

- Risk: processing endpoint semantics drift from eventual async worker model.
  - Mitigation: keep endpoint contract command-like and service-oriented so it can back future worker triggers.
- Risk: concurrency races create duplicate scrims.
  - Mitigation: enforce transaction boundaries, locking, and idempotency checks; validate with concurrency integration tests.
- Risk: insufficient observability makes operator behavior unclear.
  - Mitigation: include explicit processing summary payload plus structured logs per run.
- Risk: timeline pressure from upcoming Week 11 replay/submission work.
  - Mitigation: keep Week 10 focused on MVP processing controls and defer non-essential scheduler features; use end-of-roadmap slack for spillover.

### Week 10 Start Checklist

- Week 9 smoke script passes: `./tools/week9-smoke.sh`
- Week 9 minikube smoke script passes: `./tools/week9-k8s-smoke.sh`
- Full test suite passes: `go test ./...`
- Postgres test port available (`55432` default)
- Minikube context selected and cluster reachable (`kubectl config use-context minikube`, `minikube status`)
- Working tree clean before first Week 10 implementation commit

---

## Week 11 Execution Board (2026-04-20 to 2026-04-26)

Week objective: ship result submission and ratification MVP for scrim/match contexts, with auditable state transitions and functional coverage in both local and minikube environments.

Definition of done for Week 11:

- Result submission entity/schema supports `scrim` and `match` context binding
- Submission lifecycle endpoints exist for create/list/ratify/reject
- State transitions enforce participant-team ownership and terminal-state rules
- Integration tests cover ratification completion, duplicate pending submission conflict, and rejection flow
- Week 11 onboarding notes and both smoke automations are available (`week11-smoke.sh` and `week11-k8s-smoke.sh`)

### Scope Boundaries

In scope:

- result submission lifecycle primitives (`pending`, `ratified`, `rejected`)
- participant-team authorization semantics for ratify/reject
- JSON payload evidence storage for MVP
- API + DB + integration tests and onboarding/smoke docs

Out of scope:

- replay binary ingestion/parsing pipeline (Week 12)
- dispute arbitration workflow beyond single-step rejection
- automated stat application to standings/ratings
- background job orchestration for submission processing

### Day Plan (5 x 2h sessions)

| Day | Time Budget | Work Items | Deliverables |
| --- | ---: | --- | --- |
| Mon | 2h | Define submission lifecycle contract and invariants | `docs/adr/016-result-submission-and-ratification-mvp.md` |
| Tue | 2h | Implement submission schema/store path for create/list | `migrations/000012_create_result_submissions.up.sql`, store code |
| Wed | 2h | Implement ratify/reject API and state transitions | `/v1/result-submissions`, `/v1/result-submission-ratifications`, `/v1/result-submission-rejections` |
| Thu | 2h | Integration tests for ratification, rejection, and conflict paths | DB/API integration coverage |
| Fri | 2h | Onboarding docs + local/k8s smoke scripts + cleanup/buffer | `docs/onboarding/week11.md`, `tools/week11-smoke.sh`, `tools/week11-k8s-smoke.sh` |

### Ticket Board

| ID | Task | Estimate | Status | Acceptance Criteria |
| --- | --- | ---: | --- | --- |
| W11-01 | Define result submission lifecycle contract and invariants | 0.75h | Done | states, transition rules, and participant semantics documented |
| W11-02 | Add persistence model/migration for result submissions | 1.0h | Done | additive schema with pending uniqueness guard and transition metadata |
| W11-03 | Implement create/list submission store/service path | 1.0h | Done | submissions can be created/listed for valid scrim/match contexts |
| W11-04 | Implement ratification transition logic | 1.25h | Done | second participant ratification transitions submission to `ratified` |
| W11-05 | Implement rejection transition logic | 0.75h | Done | participant rejection transitions pending submission to `rejected` with reason |
| W11-06 | Wire submission lifecycle API handlers and validation/error mapping | 1.0h | Done | handlers return stable 2xx/4xx and map domain/store errors correctly |
| W11-07 | Write unit tests for submission validation and transition invariants | 0.75h | Done | invalid payload/state/participant scenarios covered |
| W11-08 | Write DB/API integration tests for end-to-end lifecycle behavior | 1.25h | Done | ratify/reject/conflict behavior validated end-to-end |
| W11-09 | Write Week 11 onboarding notes + local/k8s smoke automations | 0.5h | Done | contributor can execute week slice locally and on minikube |
| W11-10 | Risk/overflow buffer | 1.0h | Done | not consumed (no blockers) |

### Week 11 Risks and Mitigations

- Risk: context polymorphism (`scrim` vs `match`) introduces authorization/validation drift.
  - Mitigation: centralize context participant resolution in store logic and assert via integration tests.
- Risk: duplicate pending submissions for same context create ambiguity.
  - Mitigation: enforce partial unique DB index for pending state and verify conflict behavior in smoke/integration paths.
- Risk: contributor confusion around lifecycle semantics.
  - Mitigation: publish ADR + onboarding examples using explicit create/ratify/reject flows and expected responses.
- Risk: local-only verification misses k8s runtime behavior differences.
  - Mitigation: require both `week11-smoke.sh` and `week11-k8s-smoke.sh` pass before week completion.

### Week 11 Start Checklist

- Week 10 smoke script passes: `./tools/week10-smoke.sh`
- Week 10 minikube smoke script passes: `./tools/week10-k8s-smoke.sh`
- Full test suite passes: `go test ./...`
- Postgres test port available (`55432` default)
- Minikube context selected and cluster reachable (`kubectl config use-context minikube`, `minikube status`)
- Working tree clean before first Week 11 implementation commit

---

## Week 12 Execution Board (2026-04-27 to 2026-05-03)

Week objective: deliver replay evidence ingestion MVP with parser provenance and idempotent duplicate convergence, linked to the Week 11 result-submission lifecycle.

Definition of done for Week 12:

- Replay evidence schema exists with immutable content-hash identity
- Parser provenance runs are persisted per ingest attempt
- Replay ingest endpoint enforces context/team constraints and idempotent duplicate convergence
- Replay evidence can be linked to result submissions with explicit provenance
- Integration tests and local/minikube smoke scripts validate duplicate ingest and link/list behavior

### Scope Boundaries

In scope:

- replay evidence entity keyed by content hash
- parser provenance run recording
- submission-link association for replay evidence
- API + DB + integration tests and onboarding/smoke docs

Out of scope:

- external object-storage replay blob persistence
- full parser failure/review workflow and operator override tooling
- platform-account identity resolution to player attribution
- automatic standings/stat application from replay parse outputs

### Day Plan (5 x 2h sessions)

| Day | Time Budget | Work Items | Deliverables |
| --- | ---: | --- | --- |
| Mon | 2h | Define replay evidence/provenance MVP contract | `docs/adr/017-replay-evidence-and-parser-provenance-mvp.md` |
| Tue | 2h | Implement schema and store path for replay evidence + parse runs + links | `migrations/000013_create_replay_ingestion_slice.up.sql`, store code |
| Wed | 2h | Implement replay ingestion/list APIs with idempotent duplicate behavior | `/v1/replay-evidence`, `/v1/replay-parse-runs`, `/v1/result-submission-replay-links` |
| Thu | 2h | Integration tests for duplicate convergence and link constraints | DB/API integration coverage |
| Fri | 2h | Onboarding docs + local/k8s smoke scripts + cleanup/buffer | `docs/onboarding/week12.md`, `tools/week12-smoke.sh`, `tools/week12-k8s-smoke.sh` |

### Ticket Board

| ID | Task | Estimate | Status | Acceptance Criteria |
| --- | --- | ---: | --- | --- |
| W12-01 | Define replay evidence and parser-provenance MVP contract | 0.75h | Done | Week 12 replay ingestion invariants and API contract documented |
| W12-02 | Add replay ingestion persistence schema (evidence, parse runs, links) | 1.0h | Done | additive migration with hash uniqueness and provenance constraints |
| W12-03 | Implement replay evidence ingest/list store path | 1.0h | Done | ingest creates/reuses evidence and lists evidence rows |
| W12-04 | Implement parser provenance run recording | 0.75h | Done | parser metadata and output JSON persisted per ingest attempt |
| W12-05 | Implement replay evidence to result submission linking | 0.75h | Done | link creation enforces submission/context and participant consistency |
| W12-06 | Wire replay ingestion/list API handlers and validation/error mapping | 1.0h | Done | stable 2xx/4xx behavior for ingest/list endpoints |
| W12-07 | Write unit tests for replay-ingestion input validation | 0.75h | Done | invalid replay ingest payloads are rejected predictably |
| W12-08 | Write DB/API integration tests for dedupe/link behavior | 1.25h | Done | duplicate replay ingest converges and links remain queryable |
| W12-09 | Write Week 12 onboarding notes + local/k8s smoke automations | 0.5h | Done | contributor can execute replay ingest slice locally and on minikube |
| W12-10 | Risk/overflow buffer | 1.25h | Done | not consumed (no blockers) |

### Week 12 Risks and Mitigations

- Risk: replay dedupe behavior drifts from immutable evidence invariant.
  - Mitigation: enforce unique replay hash at DB level and verify duplicate convergence in integration + smoke tests.
- Risk: replay evidence links could attach to mismatched submissions.
  - Mitigation: require context match and participant ownership checks before link creation.
- Risk: parser provenance is under-specified for future parser upgrades.
  - Mitigation: persist parser name/version/config digest now and keep parse run model append-only.
- Risk: local success misses in-cluster behavior regressions.
  - Mitigation: require both `week12-smoke.sh` and `week12-k8s-smoke.sh` passing before week completion.

### Week 12 Start Checklist

- Week 11 smoke script passes: `./tools/week11-smoke.sh`
- Week 11 minikube smoke script passes: `./tools/week11-k8s-smoke.sh`
- Full test suite passes: `go test ./...`
- Postgres test port available (`55432` default)
- Minikube context selected and cluster reachable (`kubectl config use-context minikube`, `minikube status`)
- Working tree clean before first Week 12 implementation commit
