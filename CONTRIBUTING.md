# Contributing To Kinetic v3

This project is optimized for part-time contributors and deterministic delivery.
Use the checklist-driven workflow below so changes are easy to review, verify, and hand off.

## Core workflow

1. Sync and branch.

```bash
git checkout main
git pull
# branch prefix convention for this repo workflow
git checkout -b codex/<short-topic>
```

2. Run baseline verification before edits.

```bash
go test ./...
./tools/week12-smoke.sh
```

3. Make focused changes in a single week slice or bugfix scope.

4. Re-run quality checks.

```bash
./tools/quality-gate.sh
```

5. If your change touches deployed behavior, run minikube smoke.

```bash
./tools/week12-k8s-smoke.sh
```

6. Update docs for behavior changes.

- ADR (`docs/adr`) for design decisions
- onboarding notes (`docs/onboarding`) for contributor execution
- execution board (`plans/v3-execution-board.md`) for schedule/status

7. Commit with explicit scope.

```bash
git add -A
git commit -m "<scope>: <what changed>"
```

## Coding standards

- Keep code paths behavior-first and test-backed.
- Favor additive migrations and backward-compatible API changes.
- Keep error mapping stable (`400` for invalid input, `409` for conflicts/dependencies).
- Prefer explicit domain types over loosely structured maps.

## Testing expectations

- Unit + integration tests for each behavior change.
- Smoke scripts for full service flow on local DB path.
- Minikube smoke for in-cluster runtime validation.

Minimum acceptance for merge:

```bash
./tools/quality-gate.sh
./tools/week12-smoke.sh
```

Recommended for production-affecting changes:

```bash
./tools/week12-k8s-smoke.sh
```

## Commit discipline

- Keep commits logically grouped and reversible.
- Never rewrite history on shared branches.
- Do not mix unrelated refactors with behavior changes.

## Release safety checklist

Before promoting a release candidate:

- quality gate green
- local smoke green
- minikube smoke green
- execution board updated
- onboarding/runbook docs updated for any changed flow
