# Kinetic

Kinetic is the next-generation league operations platform for Minor League Esports.
It carries forward the mechanical identity of Sprocket while marking a clean product reboot: a statically typed Go backend, a React/TypeScript web client, Kubernetes-native operations, multi-game workflow foundations, and artifact-backed release gates for the workflows that matter most.

Kinetic is built to make league administration, scrim operations, evidence intake, roster governance, and player-facing workflows reliable enough to ship without the support loop that accumulated in earlier platform iterations.

## Current Status

Core platform slices are complete, including:

- league hierarchy, roster, queue, scrim, and scheduled match foundations
- deterministic queue promotion processing with observability
- result submission + ratification lifecycle
- replay evidence ingestion MVP with parser provenance + dedupe
- local and minikube smoke automation for weekly slices
- browser/API release evidence covering credentialed CORS, session identity, privilege isolation, and replay intake behavior

Handoff hardening and CI quality-gate enforcement are in place. Current active strategy is Pareto Recovery focused on exception automation:

- `plans/pareto-recovery-plan.md`

Planning source of truth:

- `plans/v3-execution-board.md`

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

HTTP handlers are split by route area under `internal/platform/http`. PostgreSQL-backed store wiring is assembled through `internal/platform/db.NewStores`, with capability interfaces declared in `internal/domain/hierarchy/store.go`.

Configuration variables are documented in `docs/runbooks/configuration.md`.

## Working Agreement

- Ship vertical slices weekly.
- Every shipped slice includes code + tests + docs.
- Keep architecture simple: modular monolith first, split later only with evidence.
- Treat release evidence as a required promotion artifact, not an optional smoke test.

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

# release evidence gate (local DB + API + real browser/API-mode checks)
./tools/release-evidence.sh

# full local behavior smoke (latest slice)
./tools/week12-smoke.sh

# full minikube behavior smoke (latest slice)
./tools/week12-k8s-smoke.sh

# handoff + hardening smoke
./tools/week13-smoke.sh

# final roadmap verification
./tools/week14-smoke.sh
```

`tools/release-evidence.sh` writes proof bundles under `artifacts/release-validation/<environment>/<timestamp>/`.
Generated artifacts are ignored by git, but release notes should record the artifact path used for promotion.

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

- `docs/onboarding/week1.md`
- `docs/onboarding/week2.md`
- `docs/onboarding/week3.md`
- `docs/onboarding/week4.md`
- `docs/onboarding/week5.md`
- `docs/onboarding/week6.md`
- `docs/onboarding/week7.md`
- `docs/onboarding/week8.md`
- `docs/onboarding/week9.md`
- `docs/onboarding/week10.md`
- `docs/onboarding/week11.md`
- `docs/onboarding/week12.md`
- `docs/onboarding/week13.md`
- `docs/onboarding/week14.md`
- `docs/onboarding/pareto-p1-p6.md`
- `docs/onboarding/web-phase0.md`
- `docs/onboarding/web-support-ops.md`
- `docs/onboarding/web-player-core.md`
- `docs/onboarding/web-scheduling-admin.md`
- `docs/onboarding/web-roster-and-governance.md`
