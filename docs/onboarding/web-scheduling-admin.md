# Web Scheduling Admin (F3)

Date: 2026-02-15

## Scope

This runbook covers `/app/admin` scheduling workflows:

- seasons
- schedule groups
- fixtures
- matches

## Endpoints Used

- `GET /v1/seasons`
- `POST /v1/seasons`
- `GET /v1/schedule-groups`
- `POST /v1/schedule-groups`
- `GET /v1/fixtures`
- `POST /v1/fixtures`
- `GET /v1/matches`
- `POST /v1/matches`

## Notes

- Current backend support is create/list for this slice.
- UI callouts document update/delete as pending backend support.
- Create mutations are optimistic and roll back on error.

## Verification

```bash
cd /Users/jacbaile/Sprocket-v3/web/client
npm run lint
npm run typecheck
npm run test
npm run build
npm run test:e2e
```
