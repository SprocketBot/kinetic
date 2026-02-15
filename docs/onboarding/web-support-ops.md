# Web Support and Platform Ops (F1)

Date: 2026-02-15

## Scope

This runbook covers the Week F1 web surfaces:

- League support inbox workspace (`/app/support`)
- Platform operations metrics + links (`/app/platform`)

## Support Operator Flow

1. Sign in with a principal that includes `league_support`.
2. Open `/app/support`.
3. Check live cards:
   - `Active scrims`
   - `Submissions in process`
4. Filter the inbox by severity/state as needed.
5. Select a ticket from the inbox list.
6. Use the `Triage` form to set reason/severity/suggested action.
7. Use the `Resolve` form when the case is closed.

Backed endpoints:

- `GET /v1/operator-inbox`
- `POST /v1/operator-inbox/triage`
- `POST /v1/operator-inbox/resolve`
- `GET /v1/scrims`
- `GET /v1/result-submissions`

## Platform Operator Flow

1. Sign in with a principal that includes `platform_operator`.
2. Open `/app/platform`.
3. Use quick links for Grafana/GitHub/GHCR.
4. Review metrics cards sourced from `GET /v1/exception-metrics`.

## Verification

```bash
cd /Users/jacbaile/Sprocket-v3/web/client
npm run lint
npm run typecheck
npm run test
npm run build
npm run test:e2e
```
