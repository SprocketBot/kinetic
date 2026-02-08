# ADR-002: Week 2 Auth and RBAC Baseline

- Status: Accepted
- Date: 2026-02-07
- Owner: jacbaile

## Context

Week 2 requires a practical auth and authorization baseline that can be implemented quickly and expanded safely.

The project currently has a small HTTP surface and no identity model in runtime request handling.

## Decision

Implement a minimal, explicit baseline with three parts:

1. Authentication principal model and middleware
- Every request is assigned a principal in context.
- Anonymous requests are represented explicitly.
- Invalid authorization headers are rejected with `401`.

2. Local token format for development
- Accepted header format: `Authorization: Bearer local:<subject>:<role1,role2>`
- `local` is a development prefix, not a production federation mechanism.
- Empty auth header maps to anonymous principal.

3. Static RBAC evaluator with protected endpoint reference
- Authorization decisions use `(roles, resource, action)` checks.
- Week 2 reference policy: `admin` can `read` `admin.ping`.
- Reference protected endpoint: `GET /v1/admin/ping`.

## Data Baseline

Migration `000002_create_authz_baseline.up.sql` adds:

- `roles`
- `policies`
- `user_role_bindings`

Seed baseline:

- roles: `admin`, `observer`
- policy: `admin -> admin.ping/read`

## Consequences

### Positive

- Fast path to enforceable permissions with low complexity.
- Clear testability for allow/deny behavior.
- Smooth upgrade path to JWT or provider-backed auth later.

### Negative / Tradeoffs

- Local token format is intentionally simplistic and not production-grade.
- Static evaluator duplicates policy state at runtime until DB-backed loading is added.

## Follow-ups

- Replace local token validator with signed JWT validation.
- Load role/policy bindings from DB instead of static bootstrap list.
- Introduce scoped permissions once domain entities are in place.
