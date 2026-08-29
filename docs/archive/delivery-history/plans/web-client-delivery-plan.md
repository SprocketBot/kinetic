# Kinetic v3 Web Client Delivery Plan

Date: 2026-02-15  
Status: Draft  
Inputs: `docs/interface-design.md`, `internal/platform/http/server.go`, `docs/adr/012-replay-parsing-and-platform-account-association-model.md`

## Goals

- Deliver a single web client that serves all four roles:
  - User
  - League administrator
  - League support operator
  - Platform operator
- Reuse existing backend APIs first; avoid frontend-driven backend churn.
- Reuse existing Evidence views for read-heavy league/player data instead of duplicating reporting UI.

## Non-Goals (Initial Web Client Track)

- Rebuilding Grafana/GitHub Actions dashboards inside Kinetic.
- Full replay parser admin consoles beyond existing ingestion/review flows.
- Replacing the current static Evidence publishing workflow.

## Proposed Client Architecture

1. App shape
- One SPA with role-aware route guards and role-specific navigation.
- Shared design system and component library to keep admin/support/user flows consistent.

2. Suggested stack
- React + TypeScript + Vite
- React Router (route-level auth and role guards)
- TanStack Query (server state and mutation workflows)
- Zod for runtime validation of API payloads
- Playwright for end-to-end regression coverage of critical workflows

3. API integration pattern
- Start with hand-authored API client wrappers around `/v1/*` endpoints.
- Add an API contract artifact (OpenAPI or equivalent) once endpoint surface stabilizes.
- Standardize frontend handling for `401/403`, domain validation errors, and conflict responses.

4. Auth model (UI)
- OAuth/OIDC login-only (no local password path).
- Store session token securely (httpOnly cookie preferred).
- Enforce role-based route access in both navigation and server calls.

5. Evidence integration
- Treat Evidence as source of truth for read-only analytics pages.
- Embed Evidence pages in web client for standings, static reports, and historical dashboards.
- Build native Kinetic UI only for actions/CRUD that Evidence does not support.

## Role Surface Plan

## User

MVP capabilities:
- Sign in and onboarding state visibility (`inactive` vs active player).
- Queue/scrim participation flows:
  - view queue status
  - join/leave queue
  - view active scrims and assignments
- Result submission and ratification:
  - submit result package
  - upload replay evidence
  - ratify/reject outcomes
- Read-only rating visibility.

Requires backend additions:
- OAuth provider callback/link APIs for platform account association.
- Eligibility points endpoint + decay schedule endpoint.
- User-scoped endpoint variants for cleaner client queries (optional but strongly recommended).

## League Administrator

MVP capabilities:
- CRUD:
  - seasons
  - schedule groups
  - fixtures
  - matches
- Roster/role assignment management:
  - top-level role grants/revokes
  - roster membership adjustments
- Override/NCP flow for invalid or revised results.

Requires backend additions:
- Explicit role-assignment endpoints and policy model for FM/GM/AGM/Captain authority scopes.
- Result override endpoints with audit reason and actor metadata.
- Rating adjustment mutation endpoint with self-edit prohibition rules.

## League Support Operator

MVP capabilities:
- Live scrim operations:
  - view active scrims
  - cancel/update scrim state
- Submission operations:
  - view in-flight submissions
  - submit replay on behalf of players
- Operator inbox workflows:
  - triage
  - resolve
- Add/activate players post-review.

Requires backend additions:
- Ban/unban queueing endpoint(s).
- Player activation workflow endpoint(s).
- Ticket ingestion endpoint for user support requests (if external ticketing is not the source).

## Platform Operator

MVP capabilities:
- Single "Operations" page linking out to:
  - Grafana dashboards
  - alerting system
  - GitHub Actions / release rollouts
- In-app operator inbox summary and exception metrics panel.

Requires backend additions:
- None mandatory for link-out model.
- Optional aggregated health/ops summary endpoint.

## Current API Readiness Snapshot

Ready now (existing `/v1/*` routes can back immediate UI):
- hierarchy entities (league/franchise/club/team/player)
- roster memberships
- queues, queue entries, scrims, promotions, promotion runs
- player ratings (read-only)
- seasons/schedule groups/fixtures/matches
- result submissions + ratify/reject
- replay evidence + replay parse runs
- operator inbox + triage/resolve + exception metrics

Not yet explicit in current routes:
- OAuth/OIDC login + callback exchange
- platform account link/unlink flows
- eligibility points/decay API
- ban/unban queue controls
- role-assignment hierarchy management APIs (FM/GM/AGM/Captain delegation)
- NCP/result override APIs
- rating adjustment mutation API (admin-only with guardrails)

## Delivery Phases (8h/week pace)

## Phase 0: Foundation (Week 1)

- Scaffold frontend app, CI, lint/test baseline, env config.
- Implement auth shell, route structure, role-aware nav placeholders.
- Build API client layer and error boundary patterns.

Exit criteria:
- App boots in local + CI.
- Protected vs public routes enforced.
- One integration smoke test in CI.

## Phase 1: Support + Ops First (Weeks 2-3)

- Deliver support operator console:
  - active scrims table + state update
  - submissions queue view
  - operator inbox triage/resolve
- Deliver platform operator page with external integrations + exception metrics widgets.

Exit criteria:
- Support team can process inbox end-to-end inside UI.
- No manual API calls required for triage/resolve workflows.

## Phase 2: Player Scrim Core (Weeks 4-5)

- Deliver player dashboard:
  - queue join/leave/status
  - active scrim details
  - replay evidence upload + result submission ratification
- Embed Evidence pages for read-only standings/summary where applicable.

Exit criteria:
- Player can complete queue -> scrim -> submission path in UI.
- Critical flow covered by end-to-end tests.

## Phase 3: League Scheduling Admin (Weeks 6-7)

- Deliver season/schedule/fixture/match CRUD screens.
- Add basic filters, pagination, optimistic updates with rollback.

Exit criteria:
- Admin can run weekly scheduling lifecycle without direct DB/API tooling.

## Phase 4: Roster + Role Delegation (Weeks 8-9)

- Deliver roster management UX for FM/GM/AGM/Captain scopes.
- Add assignment/offer/release workflows and audit activity log.

Dependency:
- Backend role-assignment and delegation APIs.

Exit criteria:
- Captain/GM/FM actions enforced correctly by role scope.

## Phase 5: Account Association + Eligibility + Overrides (Weeks 10-11)

- Deliver platform account linking management UI.
- Deliver eligibility/rating panels.
- Deliver NCP/result override admin workflow.

Dependencies:
- OAuth link callbacks and account association APIs.
- Eligibility and override endpoints.

Exit criteria:
- Account linking and override workflows are fully UI-driven and audited.

## Cross-Cutting Standards

- Accessibility: keyboard navigation and semantic labeling on all action-critical forms.
- Auditability: all privileged mutations show actor, timestamp, and reason where supported by API.
- Observability: frontend emits structured client telemetry for failed mutations and high-latency calls.
- Quality gate: each phase ships with unit tests + Playwright path tests + docs update.

## Backend Prerequisite Backlog (for planning board)

1. Auth/OIDC endpoints and session strategy for web client.
2. Platform account link/unlink APIs (ADR-012 follow-up).
3. Eligibility points + decay projection read endpoint.
4. Role-assignment/delegation CRUD and authority evaluation endpoints.
5. Queue ban/unban endpoint(s).
6. Result override/NCP endpoint(s) with audit log.
7. Admin rating adjustment endpoint with self-edit guardrail.

## Suggested Immediate Next Step

- Convert this plan into ticketized boards:
  - `WEB-F0-*` (frontend foundation)
  - `WEB-F1-*` ... `WEB-F5-*`
  - `API-WEB-*` for prerequisite backend endpoints
