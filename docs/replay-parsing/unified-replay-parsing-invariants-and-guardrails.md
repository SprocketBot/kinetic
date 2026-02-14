# Unified Replay Parsing Invariants And Guardrails

Date: 2026-02-14  
Status: Draft  
Depends On: `docs/replay-parsing/unified-replay-parsing-core-concepts.md`

## Purpose

Define non-negotiable replay-ingestion rules ("invariants") and measurable guardrails for implementation and operations.

## Invariants

1. Immutable replay evidence
- Raw replay bytes are immutable once ingested.
- Replay evidence identity is content-based (fingerprint/hash), not filename-based.

2. Deterministic parser provenance
- Every parse artifact records parser version and parse configuration.
- The same replay bytes parsed with the same parser version must produce equivalent canonical output.

3. Unique active platform-account ownership
- The tuple `(platform_provider, platform_account_id)` can map to only one active player at a time.
- Ownership changes require explicit, auditable reassignment.

4. Verified account association only
- Automated participant-to-player mapping uses verified platform-account links.
- Display names alone are not an authoritative identity key.

5. Stable historical attribution
- Finalized replay attribution is point-in-time stable.
- Later account-link changes must not silently rewrite historical outcomes without explicit reprocessing.

6. Idempotent replay ingestion
- Submitting the same replay multiple times must not create duplicate game/result records.
- Duplicate submissions should return/attach to the existing replay evidence identity.

7. Context integrity checks
- Replay-to-context attachment (`scrim` or `match`) requires participant/team consistency with the target context.
- Context mismatches must route to review, not auto-finalize.

8. Atomic finalization
- Participant resolution, result derivation, and stat persistence commit atomically.
- Partial writes are invalid and must rollback/retry.

9. Explicit exception states
- Unresolved participants, parser failures, and ownership conflicts must enter explicit non-final states (`needs_review`, `rejected`, etc.).
- The system cannot silently drop, partially apply, or guess unresolved identity mappings.

10. End-to-end observability
- Every ingestion attempt records: replay evidence ID, submitter, parser version, resolution outcome, and terminal/non-terminal state.
- Audit logs must be queryable for dispute and incident review.

## Baseline Operational Guardrails (Initial)

These are initial operating constraints and should be tuned with production data:

- auto-finalization requires successful parsing and deterministic participant resolution for required competitors
- duplicate detection must use replay content fingerprinting before creating new result/stat records
- manual override/reassignment actions require actor identity + reason code
- replay context association must enforce configured time-window and roster-consistency checks

## Success Metrics Guardrails

Initial targets to monitor quality and safety:

- `>= 95%` of ingested replay files parse successfully without manual intervention
- `>= 98%` participant auto-resolution accuracy against verified platform-account links
- `<= 1%` duplicate-ingest defect rate (duplicate result/stat creation)
- `<= 60s` median ingest-to-finalize latency for clean submissions
- `100%` of finalized results are traceable to replay evidence + parser provenance

## Required Test Coverage

1. Unit tests
- replay fingerprint/idempotency behavior
- platform-account uniqueness and reassignment rules
- participant resolution precedence and failure behavior

2. Integration tests
- replay upload -> parse -> resolve -> finalize for both `scrim` and `match` contexts
- duplicate submission convergence behavior
- unresolved-identity and context-mismatch routing to review states

3. Regression tests
- parser-version upgrade reprocessing and output diff checks
- historical-attribution stability under account-link changes
- end-to-end replay audit-trail completeness

## Open Calibration Questions

- How strict should replay-to-scheduled-match time windows be by competition type?
- Which limited override paths are acceptable when a provider does not emit stable participant IDs?
- Should automated reprocessing run on every parser version bump or only on opted-in scopes?

## Change Management Rule

Any change to identity resolution rules, parser-version adoption policy, replay-context validation, or override authority must include:

- updated docs
- migration/config change notes
- before/after validation metrics
- rollback plan
