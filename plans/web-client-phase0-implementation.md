# Web Client Phase 0 Implementation Spec

Date: 2026-02-15  
Scope: `WEB-F0-*` tickets in `/Users/jacbaile/Kinetic-v3/plans/web-client-execution-board.md`

## Phase 0 Goal

Stand up a production-grade frontend foundation in `web/client` with route/auth skeleton, API client primitives, and CI quality checks so feature phases can ship without rework.

## Technical Decisions

1. Package manager: `npm` (lowest friction for contributors and CI runners).
2. Runtime target: Node 22 LTS.
3. Framework/tooling: React + TypeScript + Vite.
4. Testing:
- Unit/component: Vitest + React Testing Library.
- End-to-end: Playwright (single deterministic smoke in F0).
5. State/data:
- TanStack Query for API-backed server state.
- Minimal local state via React context for session principal.

## Proposed Directory Layout

```text
web/client/
  package.json
  package-lock.json
  tsconfig.json
  tsconfig.node.json
  vite.config.ts
  vitest.config.ts
  playwright.config.ts
  .eslintrc.cjs
  .prettierrc
  .env.example
  index.html
  src/
    main.tsx
    app/
      App.tsx
      providers.tsx
      routes.tsx
    auth/
      session-context.tsx
      auth-guard.tsx
      role-guard.tsx
      auth-service.ts
    components/
      layout/
        app-shell.tsx
        nav.tsx
      feedback/
        error-boundary.tsx
        loading-state.tsx
    features/
      support/
        pages/
          support-dashboard-page.tsx
      player/
        pages/
          player-home-page.tsx
      admin/
        pages/
          admin-home-page.tsx
      platform/
        pages/
          platform-ops-page.tsx
    lib/
      api/
        client.ts
        errors.ts
        schemas.ts
      config/
        env.ts
      utils/
        permissions.ts
    test/
      setup.ts
      msw/
        handlers.ts
        server.ts
  tests/
    e2e/
      auth-guard.spec.ts
```

## App Routing Skeleton (F0)

Public routes:
- `/login`
- `/unauthorized`

Authenticated routes:
- `/app` (role-aware landing redirect)

Role-scoped placeholders:
- `/app/player`
- `/app/support`
- `/app/admin`
- `/app/platform`

Guard behavior:
- Unauthenticated -> redirect to `/login`.
- Authenticated without required role -> redirect to `/unauthorized`.
- Authenticated with role -> allow route and render app shell.

## Session/Auth Contract (F0)

Because OAuth APIs are not finalized yet (`API-WEB-01`), implement an adapter boundary:

`auth-service.ts`:
- `getSession(): Promise<SessionPrincipal | null>`
- `login(): Promise<void>` (stub redirect or mock mode action)
- `logout(): Promise<void>`

`SessionPrincipal` shape:

```ts
type Role = "player" | "league_admin" | "league_support" | "platform_operator";

type SessionPrincipal = {
  subject: string;
  displayName: string;
  roles: Role[];
};
```

Local development mock mode:
- Controlled by `VITE_AUTH_MODE=mock|api`.
- `mock` mode returns a configurable principal from env.
- `api` mode uses real backend session endpoints once available.

## API Client Baseline (F0)

Implement one shared HTTP client wrapper:
- base URL from `VITE_API_BASE_URL` (default `http://localhost:8080`).
- JSON request/response helpers.
- standardized error type with:
  - `status`
  - `code` (if provided)
  - `message`
  - `requestId` (if returned by backend later)

Add one sample endpoint integration to validate stack:
- `GET /v1/operator-inbox`

F0 only requires read display and loading/error states.

## Scripts and Commands

Inside `web/client`:

```bash
npm run dev
npm run build
npm run lint
npm run typecheck
npm run test
npm run test:e2e
```

Expected `package.json` scripts:
- `dev`: vite dev server
- `build`: tsc + vite build
- `lint`: eslint src --max-warnings=0
- `typecheck`: tsc --noEmit
- `test`: vitest run
- `test:e2e`: playwright test

## CI Integration Plan

Add new job to `/Users/jacbaile/Kinetic-v3/.github/workflows/ci.yml`:

- Job name: `web-client`
- Runs on: `ubuntu-latest`
- Steps:
  1. checkout
  2. setup-node (`22.x`)
  3. `npm ci` in `web/client`
  4. `npm run lint`
  5. `npm run typecheck`
  6. `npm run test`
  7. `npx playwright install --with-deps chromium`
  8. `npm run test:e2e`
  9. `npm run build`

Wire dependency:
- `smoke-local` should require both `quality` and `web-client` once frontend exists.

## Frontend Quality Gate Script

Add `/Users/jacbaile/Kinetic-v3/tools/web-quality-gate.sh`:
- runs lint, typecheck, unit tests, and build in `web/client`.
- Keep this separate from Go quality gate to preserve backend iteration speed.

Usage:

```bash
./tools/web-quality-gate.sh
```

## Onboarding Updates

Update `/Users/jacbaile/Kinetic-v3/README.md` and add `/Users/jacbaile/Kinetic-v3/docs/onboarding/web-phase0.md` with:
- Node/npm prerequisites
- install/run commands
- auth mock mode usage
- how to run unit and e2e tests
- troubleshooting notes (Playwright browser install, CORS/base URL mismatches)

## Acceptance Checklist (WEB-F0)

- [ ] `web/client` scaffold compiles and runs locally.
- [ ] Route guards enforce unauthenticated and unauthorized redirects.
- [ ] Role-aware nav shell renders with mock principal.
- [ ] API client successfully calls one backend endpoint with loading/error states.
- [ ] `npm run lint typecheck test test:e2e build` pass locally.
- [ ] CI has a dedicated `web-client` job and passes on PR.
- [ ] Onboarding docs cover full local setup.

## Out-of-Scope for F0

- Detailed UI styling and brand-perfect visual system.
- Feature-complete role pages.
- OAuth provider-specific callback implementation.
- Complex caching/invalidation strategies beyond basic query invalidation.

## Recommended Execution Order

1. `WEB-F0-02` + `WEB-F0-03` + `WEB-F0-04` (scaffold/tooling).
2. `WEB-F0-07` + `WEB-F0-08` + `WEB-F0-06` (auth/session + route shell).
3. `WEB-F0-09` (API client and one real query).
4. `WEB-F0-05` (Playwright smoke).
5. `WEB-F0-10` (CI wiring + onboarding docs).
