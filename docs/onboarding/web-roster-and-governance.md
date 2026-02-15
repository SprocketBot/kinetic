# Web Roster and Governance (F4/F5)

Date: 2026-02-15

## Scope

This runbook covers roster/governance surfaces in `/app/admin` and `/app/player`.

## Admin Surfaces

- Roster membership explorer and assignment (`/v1/roster-memberships`)
- Role delegation matrix with scope visibility
- Audit feed (`/v1/exception-actions`)
- Result override (NCP) form (API-pending)
- Rating administration form (API-pending)

## Player Surfaces

- Ratings panel (`/v1/player-ratings`)
- Account linking form (API-pending)
- Eligibility panel (API-pending)

## Backend Gap Notes

The following remain backend-dependent:

- `API-WEB-02`: account link/unlink endpoints
- `API-WEB-03`: eligibility points/decay endpoint
- `API-WEB-04`: delegated role grant/revoke endpoints
- `API-WEB-06`: result override/NCP submit endpoint
- `API-WEB-07`: admin rating adjustment mutation endpoint

Current UI surfaces these as explicit disabled actions with blocker copy.

## Verification

```bash
cd /Users/jacbaile/Sprocket-v3/web/client
npm run lint
npm run typecheck
npm run test
npm run build
npm run test:e2e
```
