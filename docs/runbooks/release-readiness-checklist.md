# Release Readiness Checklist

Use this checklist before cutting or promoting any release candidate.

## 1) Preflight

- working tree clean
- branch synced with `main`
- migrations reviewed for additive/backward-compatible behavior

## 2) Mandatory automated checks

```bash
./tools/ci-verify.sh
./tools/week14-smoke.sh
./tools/week14-k8s-smoke.sh
./tools/release-evidence.sh
```

All four must pass. `release-evidence.sh` writes an artifact-backed proof bundle under `artifacts/release-validation/<environment>/<timestamp>/`.

## 3) Operational readiness

- rollback command verified:

```bash
kubectl -n kinetic-v3 rollout undo deploy/kinetic-v3-api
```

- latest runbooks reflect current endpoint and smoke behavior
- execution board updated with shipped status

## 4) Release note minimums

- key behaviors shipped
- migration IDs introduced
- validation commands used
- release evidence artifact path
- known limitations and next-slice follow-ups

## 5) Post-release validation

- API health/readiness checks return `200`
- smoke target endpoints remain functional
- no critical errors in deployment logs

If any gate fails, do not promote. Fix or rollback first.
