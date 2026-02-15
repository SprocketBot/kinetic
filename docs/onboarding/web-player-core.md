# Web Player Core (F2)

Date: 2026-02-15

## Scope

This runbook covers the player-facing core workflows in `/app/player`:

- queue join/leave
- active scrim visibility
- submission ratify/reject actions
- replay evidence upload
- embedded read-only Evidence views

## Endpoints Used

- `GET /v1/queue-entries`
- `POST /v1/queue-entries`
- `DELETE /v1/queue-entries`
- `GET /v1/scrims`
- `GET /v1/result-submissions`
- `POST /v1/result-submission-ratifications`
- `POST /v1/result-submission-rejections`
- `POST /v1/replay-evidence`

## Local Validation

```bash
cd /Users/jacbaile/Sprocket-v3/web/client
npm run lint
npm run typecheck
npm run test
npm run build
npm run test:e2e
```

## Evidence Integration

`/app/player` embeds Evidence views via iframe, controlled by:

- `VITE_EVIDENCE_BASE_URL` (default `https://evidence.sprocket.gg`)

Views included in F2:

- standings
- ratings
- eligibility
