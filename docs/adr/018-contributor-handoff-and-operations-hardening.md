# ADR-018: Contributor Handoff And Operations Hardening

- Status: Accepted
- Date: 2026-02-15
- Owner: jacbaile
- Related: ADR-001 (architecture), ADR-017 (replay evidence and parser provenance MVP)

## Context

By Week 12, core behavior slices are implemented and verified through smoke automation.
The remaining roadmap risk is operational and contributor onboarding friction, not missing core endpoint behavior.

## Decision

Treat handoff and operations hardening as a first-class delivery slice.

### Required artifacts

- contributor workflow guide (`CONTRIBUTING.md`)
- dev setup and verification runbook
- deployment/rollback runbook
- replay-ingestion triage runbook

### Process guardrails

- behavior-changing work must include corresponding onboarding/runbook updates
- release readiness requires local and minikube smoke validation
- rollback guidance is explicit and executable with non-interactive commands

## Consequences

### Positive

- lower ramp-up cost for inexperienced contributors
- faster incident recovery with codified runbooks
- more predictable weekly delivery continuity

### Tradeoffs

- documentation overhead per shipped behavior change
- requires discipline to keep runbooks current with code paths

## Follow-ups

- add CI checks to verify docs and quality gates are enforced (Week 14)
- keep runbooks synchronized as replay pipeline complexity increases
