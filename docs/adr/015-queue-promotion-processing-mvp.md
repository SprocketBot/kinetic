# ADR-015: Queue Promotion Processing MVP

- Status: Accepted
- Date: 2026-02-14
- Owner: jacbaile
- Related: ADR-007 (queue enrollment), ADR-014 (scrim promotion ordering and lifecycle)

## Context

Promotion currently requires targeted calls and manual operator sequencing.
Week 10 needs a simple operational trigger that can process one queue or all active queues with deterministic, safe behavior.

## Decision

Add a synchronous promotion-processing API path for MVP operations.

### Contract

- Endpoint: `POST /v1/scrim-promotions/process`
- Input: `queueId` (`0` means process all active queues; otherwise process one queue)
- Output summary:
  - `processedQueues`
  - `promotionsCreated`
  - `conflicts`

### Safety and idempotency

- Processing reuses transactional queue promotion logic.
- Re-running processing with no eligible entries creates no new scrims.
- Queue conflicts (for example fewer than two active entries) are counted in summary, not treated as fatal process errors.
- Unexpected storage/runtime errors fail the call.

### Out of scope

- Asynchronous workers/schedulers
- Distributed queue partitioning
- Advanced rate control and backoff

## Consequences

### Positive

- Operators and smoke checks can trigger deterministic promotion processing with one call.
- Re-runs are safe and observable.

### Tradeoffs

- Synchronous endpoint may not scale indefinitely.
- Later worker model will still be required for production-grade throughput.

## Follow-ups

- Add async worker trigger path while preserving current processing contract.
- Add processing metrics/dashboard views once submission/replay flows are in place.
