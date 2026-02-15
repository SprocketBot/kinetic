# Operations And Rollback Runbook

## Deploying current main to minikube

```bash
./deploy/scripts/apply-local.sh
```

Then verify rollout:

```bash
kubectl -n sprocket-v3 rollout status deploy/sprocket-v3-api --timeout=180s
curl -sSf http://localhost:8080/healthz
```

## Mandatory post-deploy verification

```bash
./tools/week12-k8s-smoke.sh
```

## Rollback strategy

Primary rollback command:

```bash
kubectl -n sprocket-v3 rollout undo deploy/sprocket-v3-api
kubectl -n sprocket-v3 rollout status deploy/sprocket-v3-api --timeout=180s
```

If rollback is needed due to migration-related runtime breakage:

1. Stop traffic to the broken version (rollout undo above).
2. Confirm API health.
3. Assess migration compatibility before re-attempt.

## Incident triage checklist

1. Capture failing endpoint and timestamp.
2. Capture current deployment revision and image.
3. Gather API logs and event stream:

```bash
kubectl -n sprocket-v3 logs deploy/sprocket-v3-api --tail=300
kubectl -n sprocket-v3 get events --sort-by=.lastTimestamp | tail -n 50
```

4. Reproduce with local smoke (`week12-smoke.sh`) if possible.
5. If user-impacting, rollback first and investigate second.
