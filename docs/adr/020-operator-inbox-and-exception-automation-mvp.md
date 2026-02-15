# ADR-020: Operator Inbox And Exception Automation MVP

- Status: Accepted
- Date: 2026-02-15
- Owner: jacbaile
- Related: ADR-019 (CI quality gates and release hardening), Pareto Recovery Plan

## Context

Post-Week14, the highest operational cost comes from exception-heavy workflows:

- scheduling conflicts
- no-shows/forfeits
- replay/dispute triage

These need deterministic backend primitives with measurable time-saving outcomes.

## Decision

Implement an exception automation MVP with an operator inbox and rule-based triage endpoints.

### Core primitives

- `exception_tickets` as the canonical exception state record
- `exception_actions` as append-only operational audit events
- exception metrics endpoint for Pareto KPI tracking

### API MVP

- `POST /v1/exceptions/report`
- `GET /v1/operator-inbox`
- `POST /v1/operator-inbox/triage`
- `POST /v1/operator-inbox/resolve`
- `GET /v1/exception-actions`
- `GET /v1/exception-metrics`

### Rule-based automation endpoints

- `POST /v1/exception-automations/scheduling`
- `POST /v1/exception-automations/no-show`
- `POST /v1/exception-automations/replay-dispute`

## Consequences

### Positive

- centralized queue of unresolved operational work
- explicit actor/action audit trail for every exception
- KPI visibility for time-savings tracking
- deterministic automation behavior with clear override points

### Tradeoffs

- rule coverage is intentionally narrow in MVP and will miss some edge cases
- initial metrics are derived from ticket/action data and need baseline calibration

## Follow-ups

- enrich reason-code taxonomy with league-specific subcategories
- add operator UI on top of inbox APIs once workflow stabilizes
- tune automation thresholds from observed false-positive/false-negative rates
