# Pareto Baseline Report

Date: 2026-02-15

## Data source

- `GET /v1/exception-metrics`
- Week 12 local and k8s smoke behavior flows

## Baseline posture

Initial baseline is from first exception-automation MVP execution data. Values should be treated as seed metrics for trend direction, not final production benchmark numbers.

## Baseline KPIs (seed)

- `admin_hours_per_week`: captured via `adminHoursPerWeek`
- `manual_touches_per_fixture`: captured via `manualTouchesPerFixture`
- `zero_touch_fixture_rate`: captured via `zeroTouchFixtureRate`
- `time_to_close_exception`: captured via `timeToCloseHoursP50`

## Notes

- Baseline currently reflects synthetic/integration smoke traffic, not full live-season volume.
- Next weekly report should compare live operational data deltas against this seed baseline.
