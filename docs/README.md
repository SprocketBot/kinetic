# Kinetic Documentation

This directory is the single home for product documentation and planning.
Material outside `archive/` is either:

- descriptive: it defines the product, architecture, behavior, or operation of
  Kinetic; or
- prescriptive: it defines work that still remains before product cutover.

## Start Here

- Product overview: `handoff/why-kinetic.md`
- User and operator surfaces: `interface-design.md`
- Current work: `plans/product-cutover-plan.md`
- Local setup and verification: `runbooks/dev-setup-and-verification.md`
- Release readiness: `runbooks/release-readiness-checklist.md`

## Live Documentation

| Area | Purpose |
| --- | --- |
| `adr/` | Accepted architecture and behavior decisions |
| `handoff/` | Product overview and game-integration guidance |
| `matchmaking/` | Current matchmaking concepts and invariants |
| `replay-parsing/` | Current replay-processing concepts and invariants |
| `onboarding/` | Current, feature-oriented contributor guides |
| `runbooks/` | Setup, operations, release, and incident procedures |
| `plans/` | Work that remains; the product cutover plan is the plan of record |
| `archive/` | Completed, superseded, historical, or research material |

The repository root `README.md` and `CONTRIBUTING.md` remain entry-point
documents. `api/openapi.yaml` remains outside this folder because it is the
machine-readable API contract, not a planning artifact. The client-specific
setup guide remains at `web/client/README.md` beside the client code.

## Archive Policy

Archived documents are retained for provenance. They may describe decisions,
paths, and status that are no longer current and must not be treated as backlog
or implementation truth. See `archive/README.md` for the archive map.

When a plan is completed or superseded, move it into the appropriate archive
category and update this index and any live references in the same change.
