# Replay Ingestion Triage Runbook

This runbook covers Week 12 replay evidence ingestion behavior.

## Common failure classes

- invalid replay ingest payload (`400`)
- context/team mismatch (`409`)
- duplicate replay hash submitted to different context (`409`)
- submission link mismatch (`409`)

## Investigate replay evidence

```bash
curl -s http://localhost:8080/v1/replay-evidence
curl -s http://localhost:8080/v1/replay-parse-runs
curl -s http://localhost:8080/v1/result-submission-replay-links
```

## Expected dedupe behavior

- first ingest of replay body -> `201`, `duplicate=false`
- repeated ingest with same replay body and same context -> `200`, `duplicate=true`
- repeated ingest with same replay body and different context -> `409`

## Fast repro command

```bash
./tools/week12-smoke.sh
```

## Recovery approach

1. Validate context/scrim/match IDs and participant team IDs.
2. Verify linked submission context matches replay context.
3. Re-ingest only after correcting payload mismatch.
4. If behavior regressed, bisect using previous known-good commit and compare `replay_evidence` + `replay_parse_runs` rows.
