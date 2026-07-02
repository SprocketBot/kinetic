# Web Client Onboarding (Phase 0)

Date: 2026-02-15

## Purpose

Run and verify the Kinetic v3 web client foundation (`web/client`) locally.

## Prerequisites

- Node.js 22.x
- npm 10+
- backend API available at `http://localhost:8080` for API mode (optional in Phase 0)

## Install

```bash
cd /Users/jacbaile/Kinetic-v3/web/client
npm ci
```

## Run Development Server

```bash
npm run dev
```

Default Vite URL: `http://localhost:5173`

## Auth Modes

Phase 0 ships with a session adapter and mock mode.

### Mock mode (default)

Set in `web/client/.env` (or shell env):

```bash
VITE_AUTH_MODE=mock
VITE_API_BASE_URL=http://localhost:8080
VITE_MOCK_PRINCIPAL_JSON={"subject":"dev-1","displayName":"Dev User","roles":["league_support"]}
```

Behavior:

- route guards work without backend auth endpoints
- login page exposes a mock login button for local navigation

### API mode (future-ready)

```bash
VITE_AUTH_MODE=api
VITE_API_BASE_URL=http://localhost:8080
```

Expected endpoints (planned in `API-WEB-01`):

- `GET /v1/session`
- `GET /v1/auth/login`
- `POST /v1/auth/logout`

## Verification Commands

From `web/client`:

```bash
npm run lint
npm run typecheck
npm run test
npx playwright install --with-deps chromium
npm run test:e2e
npm run build
```

Or from repo root:

```bash
cd /Users/jacbaile/Kinetic-v3
./tools/web-quality-gate.sh
```

## Troubleshooting

- If Playwright tests fail due to missing browser binaries, run:
  - `cd web/client && npx playwright install --with-deps chromium`
- If API requests fail in mock mode, confirm `VITE_AUTH_MODE=mock`.
- If CORS issues appear in API mode, verify backend CORS/session configuration for the Vite origin.
