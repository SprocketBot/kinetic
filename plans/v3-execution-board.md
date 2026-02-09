# Sprocket v3 Execution Board

Last updated: 2026-02-07
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
- Next up: `W4-02` Team/Player migration implementation

## Capacity Guardrails

- Weekly budget: 10h
- Delivery target: 8h
- Reserve: 2h (planning, CI/debug, context switching)
- Hard rule: no week is complete unless code + tests + docs are all done

## 14-Week Macro Plan

| Weeks | Focus | Planned Delivery Hours | Exit Criteria |
| --- | --- | ---: | --- |
| 1-2 | Foundation | 16h | Go service scaffold + local dev + first K8s deploy |
| 3-4 | Auth + RBAC baseline | 16h | Authenticated API + role checks + tests |
| 5-7 | League hierarchy core | 24h | League/Franchise/Club/Team/Player vertical slice |
| 8-10 | Queue + match flow MVP | 24h | Join/leave/process queue end-to-end |
| 11-12 | Submission + ratification MVP | 16h | Submission lifecycle with state transitions |
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
| W4-02 | Add migrations for Team/Player | 1.25h | Todo | schema migrates cleanly and is idempotent |
| W4-03 | Extend hierarchy domain models/store interface | 0.75h | Todo | compile-safe model changes in place |
| W4-04 | Implement DB store methods for Team/Player | 1.25h | Todo | create/list for Team/Player works |
| W4-05 | Implement API handlers/routes for Team/Player | 1.0h | Todo | create/list endpoints return stable JSON |
| W4-06 | Add validation + error mapping for Team/Player payloads | 0.75h | Todo | invalid payloads return stable 4xx responses |
| W4-07 | Write unit tests for Team/Player validation | 0.5h | Todo | slug/required field checks covered |
| W4-08 | Write integration tests for DB/API Team/Player behavior | 1.25h | Todo | FK + duplicate constraints validated end-to-end |
| W4-09 | Write Week 4 onboarding notes | 0.5h | Todo | contributor can run Team/Player checks quickly |
| W4-10 | Risk/overflow buffer | 1.0h | Todo | consumed only for blockers |

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
