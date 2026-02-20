# Web Roster and Governance (F4/F5)

Date: 2026-02-15

## Scope

This runbook covers roster/governance surfaces in `/app/admin` and `/app/player`.

## Admin Surfaces

- Roster membership explorer and assignment (`/v1/roster-memberships`)
- Role delegation matrix with scope visibility
- Audit feed (`/v1/exception-actions`)
- Result override (NCP) form (`/v1/result-overrides`)
- Rating administration form (`/v1/player-ratings/adjust`)

## Player Surfaces

- Ratings panel (`/v1/player-ratings`)
- Account linking form (`/v1/platform-accounts`, `/v1/platform-accounts/link`, `/v1/platform-accounts/unlink`)
- Eligibility panel (`/v1/eligibility`)

## Backend Gap Notes

No open backend blockers remain for the current roster/governance web scope.

## Verification

```bash
cd /Users/jacbaile/Sprocket-v3/web/client
npm run lint
npm run typecheck
npm run test
npm run build
npm run test:e2e
```
