# Configuration

Sprocket v3 reads process configuration from environment variables in `internal/platform/config`.

## API runtime

| Variable | Default | Secret | Notes |
| --- | --- | --- | --- |
| `DEPLOYMENT_ENV` | `local` | No | Runtime lane. `staging`, `prod`, and `production` enable startup safety checks. |
| `PORT` | `8080` | No | HTTP listen port inside the process. |
| `LOG_LEVEL` | `info` | No | Supported levels are `debug`, `info`, `warn`, and `error`; unknown values behave like `info`. |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/sprocket?sslmode=disable` | Yes | PostgreSQL URL used by the API and migrator. |
| `MIGRATIONS_DIR` | `migrations` | No | Directory containing `*.up.sql` migration files. Use `/app/migrations` in container images. |
| `RUN_MIGRATIONS_ON_START` | `false` | No | When `true`, the API applies pending migrations during startup. |
| `REQUIRE_DATABASE` | `false` | No | When `true`, startup fails unless database connectivity, authz loading, and store initialization succeed. |

## Auth and web

| Variable | Default | Secret | Notes |
| --- | --- | --- | --- |
| `AUTH_SESSION_SECRET` | `dev-insecure-session-secret` | Yes | HMAC secret for browser session cookies. Override outside local development. |
| `AUTH_SESSION_COOKIE` | `sprocket_session` | No | Browser session cookie name. |
| `AUTH_SESSION_TTL` | `12h` | No | Parsed with Go duration syntax. Invalid or non-positive values fall back to `12h`. |
| `AUTH_LOCAL_LOGIN_ENABLED` | `true` | No | Enables local query-parameter login. Must be `false` in staging and production. |
| `WEB_BASE_URL` | `http://localhost:5173` | No | Allowed web origin and default post-login redirect base URL. |
| `CORS_ALLOWED_ORIGINS` | unset | No | Comma-separated credentialed CORS origins. Defaults to `WEB_BASE_URL`. Do not use `*` in production-like lanes. |
| `DISCORD_CLIENT_ID` | unset | No | Enables Discord OAuth only when both client ID and secret are set. |
| `DISCORD_CLIENT_SECRET` | unset | Yes | Discord OAuth client secret. |
| `DISCORD_REDIRECT_URL` | unset | No | Optional explicit Discord OAuth callback URL. If unset, the API derives one from `WEB_BASE_URL`. |

## Production safety checks

When `DEPLOYMENT_ENV` is `staging`, `prod`, or `production`, API startup fails unless:

- `AUTH_LOCAL_LOGIN_ENABLED=false`
- `AUTH_SESSION_SECRET` is non-empty and not the local default
- `WEB_BASE_URL` is not a local development URL
- `CORS_ALLOWED_ORIGINS` does not contain `*`

## Local and Kubernetes defaults

The supported local Kubernetes path uses `deploy/k8s-local-dev/configmap.yaml`, which sets `REQUIRE_DATABASE=true`, `RUN_MIGRATIONS_ON_START=true`, `MIGRATIONS_DIR=/app/migrations`, and an in-cluster `DATABASE_URL`.

The generic `deploy/k8s/configmap.yaml` keeps database-required startup disabled by default. Production-like deployments should provide a real `DATABASE_URL`, set `REQUIRE_DATABASE=true`, and inject secrets through Kubernetes Secret objects rather than ConfigMaps.
