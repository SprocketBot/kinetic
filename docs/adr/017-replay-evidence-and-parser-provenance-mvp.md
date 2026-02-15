# ADR-017: Replay Evidence And Parser Provenance MVP

- Status: Accepted
- Date: 2026-02-15
- Owner: jacbaile
- Related: ADR-012 (replay parsing and platform account association model), ADR-013 (replay parsing invariants and guardrails), ADR-016 (result submission and ratification MVP)

## Context

Week 11 delivered submission/ratification lifecycle but left replay evidence and parser provenance as out-of-band metadata.
Week 12 needs a concrete, testable ingestion slice that enforces idempotent replay evidence identity and provides auditable parser run records.

## Decision

Adopt a Week 12 replay-ingestion MVP with three persistence primitives and one command endpoint.

### Persistence primitives

- `replay_evidence`
  - immutable evidence metadata keyed by `replay_sha256`
  - explicit context (`scrim` or `match`) and submitter team provenance
- `replay_parse_runs`
  - append-only parser provenance records linked to evidence ID
  - parser name/version/config digest and normalized parser output JSON
- `result_submission_replay_links`
  - explicit link between replay evidence and result submissions

### API MVP

- `POST /v1/replay-evidence`
- `GET /v1/replay-evidence`
- `GET /v1/replay-parse-runs`
- `GET /v1/result-submission-replay-links`

### Behavioral rules

- replay evidence identity is SHA-256 over submitted replay body
- duplicate ingest of identical replay body returns existing evidence (`duplicate=true`) and does not create a second evidence row
- each ingest attempt records a parser provenance row
- submission linking requires context match and participant-team ownership

## Consequences

### Positive

- establishes auditable evidence + parser provenance baseline
- delivers idempotent replay ingest behavior for functional tests and local/k8s smoke coverage
- keeps Week 12 scope bounded while aligning to ADR-012/013 invariants

### Tradeoffs

- replay bytes are represented as MVP body payload, not external object storage yet
- parse runs are synchronous and parser status is simplified to MVP success path
- participant identity resolution to player/platform-account entities remains for a later slice

## Follow-ups

- add platform-account association entities and resolution workflow
- move replay body storage from inline payload to object store with immutable object refs
- extend parser run states and review workflow (`failed`, `needs_review`) with operator actions
