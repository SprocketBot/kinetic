# ADR-001: v3 Architecture Baseline

- Status: Accepted
- Date: 2026-02-07
- Owner: jacbaile

## Context

Kinetic v1/v2 accumulated complexity from dynamic runtime behavior and distributed infrastructure patterns that were expensive to operate and hard for new contributors to reason about.

The v3 goal is to maximize maintainability, correctness, and handoff readiness under hobby-project constraints (10h/week gross).

## Decision

Kinetic v3 will start as a Go modular monolith, deployed on Kubernetes, with PostgreSQL as the primary system of record.

### Core decisions

1. Runtime and language
- Backend services are implemented in Go.
- Emphasis on compile-time safety and explicit interfaces.

2. Service shape
- Start with a modular monolith in one deployable service.
- Domain boundaries are enforced via package/module boundaries.
- Decompose into multiple services only when metrics show a need.

3. Data and async processing
- PostgreSQL is the source of truth.
- Use migration-driven schema evolution.
- Use an outbox/event-queue table pattern for async workflows.

4. Platform
- Kubernetes is the target runtime for staging/production.
- Keep local developer experience simple and fast.

5. Quality bar
- Every feature ships with unit and behavior-oriented tests.
- CI gates include formatting, linting, tests, and build.

## Consequences

### Positive

- Lower operational complexity than immediate microservices.
- Better onboarding through clear boundaries and fewer moving parts.
- Stronger correctness guarantees from static typing and tests.

### Negative / Tradeoffs

- Initial migration may be slower than extending v2 directly.
- Some v2 components (for example replay parsing) may remain polyglot during transition.
- Kubernetes introduces platform overhead that must be managed with templates and runbooks.

## Non-Goals (for Week 1)

- Full feature parity with v2.
- Premature service decomposition.
- Advanced platform features beyond a minimal deployable baseline.

## Review Trigger

Revisit this ADR if:

- sustained production load indicates a hard scaling bottleneck,
- team size and ownership model require service-level isolation,
- local developer setup exceeds the 30-minute onboarding target.
