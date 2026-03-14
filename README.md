# Sprocket v3

Sprocket v3 is a reimagining of the platform as a statically typed, compiled backend (Go) with Kubernetes-native operations.

## Current Status

Weeks 1-12 are complete, including:

- league hierarchy, roster, queue, scrim, and scheduled match foundations
- deterministic queue promotion processing with observability
- result submission + ratification lifecycle
- replay evidence ingestion MVP with parser provenance + dedupe
- local and minikube smoke automation for weekly slices

Weeks 13-14 complete handoff hardening and CI quality-gate enforcement.
Current active strategy is Pareto Recovery focused on exception automation:

- `/Users/jacbaile/Sprocket-v3/plans/pareto-recovery-plan.md`

Planning source of truth:

- `/Users/jacbaile/Sprocket-v3/plans/v3-execution-board.md`

## Repository Layout

- `cmd/`: entrypoints (binaries)
- `internal/`: app-private code (domain + platform)
- `pkg/`: optional reusable packages intended for external use
- `web/`: web client workspace (React/TypeScript)
- `migrations/`: database migration files
- `deploy/`: Kubernetes manifests/charts and deployment scripts
- `docs/`: ADRs, runbooks, onboarding
- `test/`: integration/functional test assets
- `tools/`: development tooling scripts

## Working Agreement

- Ship vertical slices weekly.
- Every shipped slice includes code + tests + docs.
- Keep architecture simple: modular monolith first, split later only with evidence.

## Quick Commands

```bash
# start the supported local k8s-backed dev path
./tools/start-dev.sh

# start local k8s-backed dev path and web client together
./tools/start-dev.sh --with-web

# seed sample data after the API is reachable on localhost:8080
./tools/seed-dev.sh

# quality gate (fmt + vet + test + build)
./tools/quality-gate.sh

# web quality gate (lint + typecheck + test + build)
./tools/web-quality-gate.sh

# full local behavior smoke (latest slice)
./tools/week12-smoke.sh

# full minikube behavior smoke (latest slice)
./tools/week12-k8s-smoke.sh

# handoff + hardening smoke
./tools/week13-smoke.sh

# final roadmap verification
./tools/week14-smoke.sh
```

## Local K8s Dev

Supported local testing path:

1. `minikube start`
2. `kubectl config use-context minikube`
3. `./tools/start-dev.sh`
4. In a second shell, `./tools/seed-dev.sh`
5. Optional web UI: `./tools/start-dev.sh --with-web`

Notes:

- `tools/start-dev.sh` deploys the API and an in-cluster Postgres to Minikube using `deploy/k8s-local-dev`.
- The API is port-forwarded to `http://localhost:8080`.
- The web client still runs on the host via Vite; it is not deployed to Kubernetes in this repo.
- Clean up Minikube dev resources with `kubectl delete -k deploy/k8s-local-dev`.

Current onboarding guides:

- `/Users/jacbaile/Sprocket-v3/docs/onboarding/week1.md`
- `/Users/jacbaile/Sprocket-v3/docs/onboarding/week2.md`
- `/Users/jacbaile/Sprocket-v3/docs/onboarding/week3.md`
- `/Users/jacbaile/Sprocket-v3/docs/onboarding/week4.md`
- `/Users/jacbaile/Sprocket-v3/docs/onboarding/week5.md`
- `/Users/jacbaile/Sprocket-v3/docs/onboarding/week6.md`
- `/Users/jacbaile/Sprocket-v3/docs/onboarding/week7.md`
- `/Users/jacbaile/Sprocket-v3/docs/onboarding/week8.md`
- `/Users/jacbaile/Sprocket-v3/docs/onboarding/week9.md`
- `/Users/jacbaile/Sprocket-v3/docs/onboarding/week10.md`
- `/Users/jacbaile/Sprocket-v3/docs/onboarding/week11.md`
- `/Users/jacbaile/Sprocket-v3/docs/onboarding/week12.md`
- `/Users/jacbaile/Sprocket-v3/docs/onboarding/week13.md`
- `/Users/jacbaile/Sprocket-v3/docs/onboarding/week14.md`
- `/Users/jacbaile/Sprocket-v3/docs/onboarding/pareto-p1-p6.md`
- `/Users/jacbaile/Sprocket-v3/docs/onboarding/web-phase0.md`
- `/Users/jacbaile/Sprocket-v3/docs/onboarding/web-support-ops.md`
- `/Users/jacbaile/Sprocket-v3/docs/onboarding/web-player-core.md`
- `/Users/jacbaile/Sprocket-v3/docs/onboarding/web-scheduling-admin.md`
- `/Users/jacbaile/Sprocket-v3/docs/onboarding/web-roster-and-governance.md`
