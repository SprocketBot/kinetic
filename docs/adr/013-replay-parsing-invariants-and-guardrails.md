# ADR-013: Replay Parsing Invariants And Guardrails

- Status: Accepted
- Date: 2026-02-14
- Owner: jacbaile
- Related: ADR-012 (replay parsing and platform account association model)
- Source Parser: <https://github.com/SprocketBot/sprocket-rl-parser>
- Detailed References: `docs/replay-parsing/unified-replay-parsing-invariants-and-guardrails.md`, `docs/replay-parsing/unified-replay-parsing-implementation-checklist.md`

## Context

Replay ingestion and identity resolution are foundational for league integrity.
Without explicit invariants, implementation drift can cause silent misattribution, duplicate result records, and non-auditable corrections.

The system needs stable non-negotiables and measurable operating guardrails before broad rollout.

## Decision

Define and enforce invariant rules and operating guardrails for replay parsing and replay-derived result/stat processing.

### Invariants

1. Immutable replay evidence
- Replay bytes are immutable after ingest.
- Evidence identity is content-based, not filename-based.

2. Deterministic parser provenance
- Every parse stores parser version and parse config fingerprint.
- Equivalent bytes plus parser version produce equivalent canonical output.

3. Unique active platform-account ownership
- `(provider, provider_account_id)` has one active player owner at a time.
- Reassignment requires explicit audited action.

4. Verified association only
- Participant auto-resolution uses verified platform-account links only.
- Display names are never authoritative identity keys.

5. Stable historical attribution
- Finalized attribution is point-in-time stable.
- Historical rewrites require explicit reprocessing workflow.

6. Idempotent ingestion
- Duplicate replay submissions cannot create duplicate result/stat records.
- Duplicate paths converge on existing canonical evidence identity.

7. Context integrity checks
- Replay-to-context attachment requires participant/team consistency.
- Mismatches route to review, not auto-finalization.

8. Atomic finalization
- Participant resolution, result derivation, and stat writes commit atomically.
- Partial updates are invalid.

9. Explicit exception states
- Parser failures, unresolved identities, and conflicts move to explicit non-final states.
- Silent drop/guessing behavior is prohibited.

10. End-to-end observability
- Every ingestion attempt records submitter, evidence ID, parser provenance, and outcome.
- Audit records must support dispute and incident reconstruction.

### Baseline operational guardrails (initial)

- Auto-finalization requires successful parse and deterministic identity resolution for required competitors.
- Duplicate detection runs on replay content fingerprint before result/stat writes.
- Override and reassignment actions require actor identity and reason code.
- Context validation enforces roster-consistency and scheduling-window rules.

### Success metrics guardrails (initial)

- `>= 95%` replay parse success without manual intervention
- `>= 98%` participant auto-resolution accuracy against verified links
- `<= 1%` duplicate-ingest defect rate
- `<= 60s` median ingest-to-finalize latency for clean submissions
- `100%` finalized result/stat records traceable to evidence and parser provenance

### Change management rule

Any change to participant resolution rules, parser adoption policy, replay context validation, or override authority must include:

- docs update
- migration/config notes
- before/after validation metrics
- rollback plan

## Consequences

### Positive

- Reduces integrity risk for replay-derived results and statistics.
- Improves auditability and operational confidence.
- Creates clear quality gates for scaling replay automation.

### Negative / Tradeoffs

- Adds process overhead to parser upgrades and identity-policy changes.
- Requires telemetry and manual review tooling earlier in delivery.

## Follow-ups

- Build invariant-focused integration/regression suite for replay ingestion.
- Implement dashboards and alerting aligned to guardrail metrics.
- Periodically calibrate context-validation thresholds with live data.
