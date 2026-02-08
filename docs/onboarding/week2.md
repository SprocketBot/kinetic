# Week 2 Onboarding Notes (Auth/RBAC Baseline)

This guide explains how to exercise the Week 2 auth/RBAC foundation.

## Token Format (development only)

Use:

`Authorization: Bearer local:<subject>:<role1,role2>`

Examples:

- admin request: `Bearer local:alice:admin`
- observer request: `Bearer local:bob:observer`
- anonymous request: no Authorization header

## Protected Endpoint

- `GET /v1/admin/ping`
- Permission required: `admin.ping` + `read`

Behavior:

- no auth -> `401`
- invalid header -> `401`
- non-admin role -> `403`
- admin role -> `200`

## Policy Source

- If `REQUIRE_DATABASE=true` (or startup DB mode is enabled), authz policies are loaded from the `policies` table at startup.
- If DB is not required, the API falls back to a static in-memory baseline policy set for local development.

## Quick Manual Check

Start API:

```bash
go run ./cmd/api
```

In another terminal:

```bash
# 401
curl -i http://localhost:8080/v1/admin/ping

# 403
curl -i -H 'Authorization: Bearer local:bob:observer' http://localhost:8080/v1/admin/ping

# 200
curl -i -H 'Authorization: Bearer local:alice:admin' http://localhost:8080/v1/admin/ping
```

## Test Commands

```bash
go test ./...

# auth tests only
go test ./internal/platform/auth ./internal/domain/authz ./internal/platform/http
```
