# ADR-019: CI Quality Gates And Release Hardening

- Status: Accepted
- Date: 2026-02-15
- Owner: jacbaile
- Related: ADR-001 (architecture), ADR-018 (contributor handoff and operations hardening)

## Context

After Week 13, contributor and operational runbooks exist, but CI enforcement still only guarantees fmt/vet/test/build.
The project needs stronger automated quality gates to reduce regressions and support handoff confidence.

## Decision

Adopt stricter CI quality gates and a scripted release-verification workflow.

### CI quality gates

- run local quality gate script (`fmt`, `vet`, `test`, `build`)
- run `staticcheck` across repository
- run race-enabled test suite (`go test -race ./...`)
- run full local smoke script in CI (`week14-smoke.sh`)

### Release verification posture

- release readiness requires CI verification script and full local/minikube smokes
- release promotion requires artifact-backed release evidence from `./tools/release-evidence.sh`
- release evidence must prove credentialed CORS, browser API-mode auth identity, privilege isolation, and replay intake/dedupe/context rejection through production-shaped HTTP boundaries
- release/rollback checklist is documented in runbooks

## Consequences

### Positive

- better detection of concurrency and static-analysis defects
- higher confidence in handoff to new maintainers
- deterministic, script-driven release validation
- better protection against browser-only auth/CORS regressions and replay-intake failures

### Tradeoffs

- longer CI runtime
- higher resource usage on CI runners

## Follow-ups

- consider optional nightly minikube smoke in CI if runtime budget allows
- add PR template checklist tied to release/runbook gates
- add a hosted-environment variant of release evidence once dedicated harness identities and stable replay fixtures are available
