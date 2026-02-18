# Sprocket v3 Web Client Execution Board

Last updated: 2026-02-15  
Owner capacity: 10 hrs/week gross, 8 hrs/week planned delivery  
Source plan: `/Users/jacbaile/Sprocket-v3/plans/web-client-delivery-plan.md`

## Program Objective

Deliver a role-aware web client that covers support/operator workflows first, then player and league admin workflows, while reusing Evidence for read-only reporting views.

## Program Guardrails

- Every phase ships code + tests + docs.
- No frontend phase is marked done without at least one Playwright end-to-end path for that slice.
- Prefer API reuse over introducing new backend endpoints unless blocked.
- Route-level authorization and server-side authorization must both be enforced.

## Phase Timeline

| Weeks | Phase | Planned Delivery Hours | Exit Criteria |
| --- | --- | ---: | --- |
| 1 | Phase 0: Foundation | 8h | Frontend scaffold, CI quality checks, role-aware routing shell |
| 2-3 | Phase 1: Support + Ops | 16h | Operator inbox + scrim/support workflows operational in UI |
| 4-5 | Phase 2: Player Scrim Core | 16h | Player queue/scrim/submission flow works end-to-end |
| 6-7 | Phase 3: Scheduling Admin | 16h | Seasons/schedule groups/fixtures/matches CRUD in UI |
| 8-9 | Phase 4: Roster + Role Delegation | 16h | FM/GM/AGM/Captain roster operations enforced by scope |
| 10-11 | Phase 5: Accounts + Eligibility + Overrides | 16h | Account linking, eligibility, and NCP override flows complete |

## Dependency Board (API-WEB)

| ID | Task | Priority | Status | Acceptance Criteria |
| --- | --- | --- | --- | --- |
| API-WEB-01 | Auth/OIDC login + callback + session endpoints | P0 | Todo | Web client can complete login/logout without local passwords |
| API-WEB-02 | Platform account link/unlink endpoints | P0 | Todo | User can link/unlink Steam/Xbox/PSN/Epic with audited ownership |
| API-WEB-03 | Eligibility points + decay endpoint | P1 | Todo | User can fetch points + eligible-until projection |
| API-WEB-04 | Role-assignment and delegation APIs (FM/GM/AGM/Captain) | P1 | Todo | Scoped assign/revoke checks enforced server-side |
| API-WEB-05 | Queue ban/unban endpoints | P1 | Todo | Support can ban/unban queue participation with reason trail |
| API-WEB-06 | Result override/NCP endpoints | P1 | Todo | Admin can override result with actor + reason audit |
| API-WEB-07 | Admin rating adjustment endpoint + self-edit guardrail | P2 | Done | Admin can edit others, cannot edit own rating |

## Week F0 Execution Board (2026-02-16 to 2026-02-22)

Week objective: establish the frontend foundation so subsequent role-specific slices can ship safely and quickly.

Definition of done for Week F0:

- Frontend workspace scaffolded under `web/client`
- Type-safe API client + auth/session shell integrated
- Role-aware route guards and navigation placeholders implemented
- Frontend quality gate in CI (lint, typecheck, unit test, build, e2e smoke)
- Onboarding/runbook added for local frontend development

### Day Plan (5 x ~2h sessions)

| Day | Time Budget | Work Items | Deliverables |
| --- | ---: | --- | --- |
| Mon | 2h | Scaffold app + toolchain + directory conventions | `web/client/*`, lockfile, baseline scripts |
| Tue | 2h | Add route tree + auth/session providers + role guards | protected route shell + nav skeleton |
| Wed | 2h | Build API client wrappers and error handling conventions | `src/lib/api/*`, query client, error boundaries |
| Thu | 2h | Add tests and Playwright smoke path | unit test baseline + one e2e route-guard test |
| Fri | 2h | CI wiring + docs + cleanup buffer | updated `.github/workflows/ci.yml`, onboarding docs |

### Ticket Board

| ID | Task | Estimate | Status | Acceptance Criteria |
| --- | --- | ---: | --- | --- |
| WEB-F0-01 | Define frontend conventions and folder structure | 0.5h | Done | doc section finalized and reflected in tree |
| WEB-F0-02 | Scaffold React + TypeScript + Vite app in `web/client` | 1.0h | Done | `npm run dev/build` succeeds |
| WEB-F0-03 | Add ESLint + Prettier + strict TypeScript settings | 0.75h | Done | lint and typecheck scripts pass |
| WEB-F0-04 | Configure test stack (Vitest + Testing Library) | 0.75h | Done | sample component and test pass |
| WEB-F0-05 | Configure Playwright for one auth/guard smoke test | 1.0h | Done | e2e smoke passes locally and in CI |
| WEB-F0-06 | Implement app layout and role-aware nav shell | 0.75h | Done | role-specific nav groups render from auth state |
| WEB-F0-07 | Implement auth/session provider contract and mock mode | 1.0h | Done | app boots with mock principal in local dev |
| WEB-F0-08 | Implement route guard primitives (`public`, `authenticated`, `role-scoped`) | 0.75h | Done | unauthorized routes redirect correctly |
| WEB-F0-09 | Implement API client wrappers + shared error handling | 0.75h | Done | one sample endpoint query/mutation wired |
| WEB-F0-10 | Add frontend CI job and docs/runbook updates | 0.75h | Done | PR CI runs frontend checks + onboarding updated |

### Risks and Mitigations

- Risk: premature design-system work burns schedule.
  - Mitigation: ship only shell primitives in F0; defer deeper components to feature phases.
- Risk: auth uncertainty blocks route implementation.
  - Mitigation: build session adapter interface now with mock/local provider until `API-WEB-01` lands.
- Risk: flaky e2e setup slows CI.
  - Mitigation: keep F0 e2e to one deterministic route-guard scenario.

### Week F0 Start Checklist

- Go backend quality baseline still green: `./tools/quality-gate.sh`
- Latest API service can run locally for integration tests
- Node LTS selected for frontend toolchain (recommend 22.x)
- No unresolved decisions on frontend package manager

## Week F1 Execution Board (2026-02-23 to 2026-03-08)

Week objective: deliver support-operator and platform-operator web workflows first.

Definition of done for Week F1:

- Support operator can triage and resolve inbox tickets in UI.
- Support operator can view active scrims and in-process submissions.
- Platform operator page links to Grafana/GitHub/release ops with exception metrics cards.
- Phase e2e covers inbox triage path.

### Ticket Board

| ID | Task | Estimate | Status | Acceptance Criteria |
| --- | --- | ---: | --- | --- |
| WEB-F1-01 | Build support dashboard layout and filters | 1.5h | Done | operator sees prioritized work queues |
| WEB-F1-02 | Implement operator inbox list/detail UI | 2.0h | Done | list + ticket detail render from `/v1/operator-inbox` |
| WEB-F1-03 | Implement triage/resolve action forms | 1.5h | Done | mutations hit triage/resolve endpoints with optimistic updates |
| WEB-F1-04 | Build active scrims panel | 1.5h | Done | scrim list + state update works |
| WEB-F1-05 | Build submissions-in-process panel | 1.5h | Done | submission list and status chips available |
| WEB-F1-06 | Build platform operations links + metrics card | 1.0h | Done | ops links and exception metrics surfaced |
| WEB-F1-07 | Add F1 e2e and unit coverage | 1.5h | Done | inbox triage path covered in CI |
| WEB-F1-08 | Update runbook/docs for support/operator usage | 0.5h | Done | operator guide committed |

## Week F2 Execution Board (2026-03-09 to 2026-03-22)

Week objective: deliver player scrim core UX.

Definition of done for Week F2:

- Player can join/leave queues and view active scrim details.
- Player can upload replay evidence and move through submission/ratification path.
- Evidence read-only embeds integrated for standings and related views.

### Ticket Board

| ID | Task | Estimate | Status | Acceptance Criteria |
| --- | --- | ---: | --- | --- |
| WEB-F2-01 | Build player home/dashboard shell | 1.0h | Done | key scrim widgets and status cards render |
| WEB-F2-02 | Implement queue join/leave/status interactions | 2.0h | Done | queue flow works against `/v1/queue-entries` |
| WEB-F2-03 | Implement active scrim details and lifecycle display | 1.5h | Done | scrim detail reflects live state |
| WEB-F2-04 | Implement replay evidence upload flow | 2.0h | Done | replay upload and duplicate handling surfaced |
| WEB-F2-05 | Implement submission ratify/reject views | 1.5h | Done | result ratification actions available |
| WEB-F2-06 | Embed Evidence read-only pages | 1.0h | Done | embedded reports load inside UI frame |
| WEB-F2-07 | Add F2 e2e coverage | 1.0h | Done | queue -> scrim -> submit happy path covered |

## Week F3 Execution Board (2026-03-23 to 2026-04-05)

Week objective: deliver league scheduling administration UX.

### Ticket Board

| ID | Task | Estimate | Status | Acceptance Criteria |
| --- | --- | ---: | --- | --- |
| WEB-F3-01 | Build seasons CRUD screens | 2.0h | Done | create/list/edit/delete behaviors validated |
| WEB-F3-02 | Build schedule groups CRUD screens | 1.5h | Done | group flows wired and validated |
| WEB-F3-03 | Build fixtures CRUD screens | 1.5h | Done | fixture lifecycle managed in UI |
| WEB-F3-04 | Build matches CRUD screens | 1.5h | Done | match scheduling updates work |
| WEB-F3-05 | Add table filtering/pagination patterns | 1.0h | Done | lists usable at larger dataset sizes |
| WEB-F3-06 | Add optimistic mutation + rollback handling | 1.0h | Done | failed writes recover cleanly |
| WEB-F3-07 | Add F3 e2e coverage + docs | 1.5h | Done | scheduling admin path covered in CI |

## Week F4 Execution Board (2026-04-06 to 2026-04-19)

Week objective: deliver roster and delegated role management UX.

### Ticket Board

| ID | Task | Estimate | Status | Acceptance Criteria |
| --- | --- | ---: | --- | --- |
| WEB-F4-01 | Build roster explorer and management views | 2.0h | Done | FM/GM/AGM/Captain roster views available |
| WEB-F4-02 | Build assignment/offer/release flows | 2.0h | Done | roster actions complete with validation |
| WEB-F4-03 | Implement scoped role-action visibility in UI | 1.5h | Done | users only see actions they can invoke |
| WEB-F4-04 | Build role-assignment admin views | 1.5h | Done | grant/revoke flows wired to API-WEB-04 |
| WEB-F4-05 | Build roster activity/audit panel | 1.0h | Done | actor/reason/timestamp surfaced |
| WEB-F4-06 | Add F4 e2e coverage + docs | 2.0h | Done | delegated management flows covered |

## Week F5 Execution Board (2026-04-20 to 2026-05-03)

Week objective: deliver account linking, eligibility visibility, and admin override workflows.

### Ticket Board

| ID | Task | Estimate | Status | Acceptance Criteria |
| --- | --- | ---: | --- | --- |
| WEB-F5-01 | Build platform account link/unlink UI | 2.0h | Done | OAuth callback and linked account management works |
| WEB-F5-02 | Build eligibility and rating panels | 1.5h | Done | points/decay/rating cards render from APIs |
| WEB-F5-03 | Build NCP/result override workflow UI | 2.0h | Done | admin override path captures reason and audit fields |
| WEB-F5-04 | Build rating adjustment admin flow | 1.0h | Done | self-edit blocked and explained |
| WEB-F5-05 | Hardening pass across all roles | 1.5h | Done | critical defects and access leaks resolved |
| WEB-F5-06 | End-to-end regression pack + release checklist | 2.0h | Done | phase completion checklist signed off |

## Program Exit Criteria

- All role-critical workflows in `docs/interface-design.md` are covered by either:
  - native UI path, or
  - embedded Evidence path for read-only requirements.
- All privileged actions produce auditable server records.
- CI enforces frontend and backend quality gates on PRs.
