# Week 13 Onboarding Notes (Handoff + Ops Hardening)

Week 13 focuses on contributor handoff readiness and operational runbook coverage.

## New artifacts

- `CONTRIBUTING.md`
- `docs/adr/018-contributor-handoff-and-operations-hardening.md`
- `docs/runbooks/dev-setup-and-verification.md`
- `docs/runbooks/operations-and-rollback.md`
- `docs/runbooks/replay-ingestion-triage.md`

## Contributor baseline path

```bash
go test ./...
./tools/quality-gate.sh
./tools/week13-smoke.sh
```

## In-cluster verification path

```bash
kubectl config use-context minikube
minikube status
./tools/week13-k8s-smoke.sh
```

## Expected outcomes

- contributor workflow is explicit and repeatable
- rollback procedure is documented and executable
- replay ingestion triage path is documented for common failure classes
