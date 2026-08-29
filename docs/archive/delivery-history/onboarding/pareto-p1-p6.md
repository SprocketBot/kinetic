# Pareto Track Onboarding (P1-P6)

This guide covers the exception-automation track delivered from P1 through P6.

## Scope

- exception reporting and inbox lifecycle
- rule-based automation for scheduling/no-show/replay-dispute flows
- KPI metrics for operator time reduction

## Key endpoints

- `POST /v1/exceptions/report`
- `GET /v1/operator-inbox`
- `POST /v1/operator-inbox/triage`
- `POST /v1/operator-inbox/resolve`
- `GET /v1/exception-actions`
- `GET /v1/exception-metrics`
- `POST /v1/exception-automations/scheduling`
- `POST /v1/exception-automations/no-show`
- `POST /v1/exception-automations/replay-dispute`

## Verification

```bash
./tools/week12-smoke.sh
./tools/week12-k8s-smoke.sh
./tools/week14-smoke.sh
./tools/week14-k8s-smoke.sh
```

These flows now include exception automation assertions.

## KPI interpretation

`/v1/exception-metrics` returns:

- `adminHoursPerWeek`
- `manualTouchesPerFixture`
- `zeroTouchFixtureRate`
- `timeToCloseHoursP50`

Use these for weekly Pareto-track trend monitoring.
