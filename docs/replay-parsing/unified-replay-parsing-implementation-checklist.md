# Unified Replay Parsing Implementation Checklist

Date: 2026-02-14  
Status: Draft  
Depends On: `docs/adr/012-replay-parsing-and-platform-account-association-model.md`, `docs/adr/013-replay-parsing-invariants-and-guardrails.md`

## Purpose

Map each replay-ingestion invariant to actionable implementation tasks across schema, API, services, operations, and tests.

## How to use

- Treat each checklist block as a vertical slice.
- Do not mark an invariant complete without test coverage and telemetry.
- Keep parser selection, validation windows, and override authority in config, not hardcoded service logic.

## Invariant 1: Immutable Replay Evidence

Schema tasks:

- [ ] Add `replay_evidence` table keyed by content hash (`sha256`) with immutable metadata.
- [ ] Add unique constraint on replay content hash and immutable `stored_object_ref`.
- [ ] Add submitter and received timestamp fields for provenance.

API tasks:

- [ ] Add replay upload endpoint that returns canonical `replayEvidenceId`.
- [ ] Reject mutation of raw replay bytes after ingest.

Service tasks:

- [ ] Compute replay hash on stream ingest before persistence.
- [ ] Deduplicate by hash and reuse existing evidence record when already present.

Testing tasks:

- [ ] Unit test hash identity behavior for same/different bytes.
- [ ] Integration test duplicate upload convergence to one evidence record.

## Invariant 2: Deterministic Parser Provenance

Schema tasks:

- [ ] Add `replay_parse_runs` table with parser name, parser version, config digest, and parse status.
- [ ] Store canonical parse artifact references linked to evidence ID.

API tasks:

- [ ] Expose parser provenance in replay admin/detail payloads.

Service tasks:

- [ ] Route parsing through version-pinned `sprocket-rl-parser` adapter.
- [ ] Persist parser metadata and config digest on every parse attempt.

Testing tasks:

- [ ] Golden tests for deterministic canonical output for fixed bytes + parser version.
- [ ] Regression tests for parser version bump behavior and provenance persistence.

## Invariant 3: Unique Active Platform-Account Ownership

Schema tasks:

- [ ] Add `platform_accounts` table with `(provider, provider_account_id)` identity keys.
- [ ] Add unique active ownership constraint on `(provider, provider_account_id)`.
- [ ] Add auditable ownership reassignment history table.

API tasks:

- [ ] Add account link endpoint for OAuth callback completion.
- [ ] Add conflict response contract when account is already linked to another player.

Service tasks:

- [ ] Implement verified account-link flow from provider identity response.
- [ ] Require explicit reassignment path (privileged or audited user flow).

Testing tasks:

- [ ] Unit tests for uniqueness enforcement and reassignment rules.
- [ ] Integration tests for conflict handling and ownership transfer audit records.

## Invariant 4: Verified Account Association Only

Schema tasks:

- [ ] Add verification metadata fields (`verified_at`, `verification_method`) on account links.

API tasks:

- [ ] Surface participant-resolution confidence and unresolved reason codes.

Service tasks:

- [ ] Resolve replay participants by provider + provider account ID only.
- [ ] Route name-only matches to review state, never automatic finalization.

Testing tasks:

- [ ] Unit tests proving display-name collisions cannot auto-link participants.
- [ ] Integration tests for unresolved participant routing.

## Invariant 5: Stable Historical Attribution

Schema tasks:

- [ ] Add immutable attribution snapshot records attached to finalized replay outputs.
- [ ] Add explicit correction/reprocessing event table with actor and reason.

API tasks:

- [ ] Add admin reprocess endpoint requiring reason code.

Service tasks:

- [ ] Snapshot resolved player mapping at finalization time.
- [ ] Block silent historical rewrites unless reprocessing flow is invoked.

Testing tasks:

- [ ] Regression tests ensuring post-finalization account relinks do not mutate history.
- [ ] Integration tests for audited reprocessing path.

## Invariant 6: Idempotent Replay Ingestion

Schema tasks:

- [ ] Add unique constraint tying one canonical replay evidence record to one derived game result scope.
- [ ] Add idempotency key support on submission events.

API tasks:

- [ ] Return existing canonical replay/game identity for duplicate submissions.

Service tasks:

- [ ] Implement idempotent ingest workflow keyed by replay hash + context.
- [ ] Make downstream derivation steps safe to retry.

Testing tasks:

- [ ] Integration tests for repeated identical submissions under race conditions.
- [ ] Property tests for idempotent outcome under retry storms.

## Invariant 7: Context Integrity Checks

Schema tasks:

- [ ] Add `replay_context_links` with context type (`scrim` or `match`), target ID, and validation result.
- [ ] Store context mismatch reason codes.

API tasks:

- [ ] Require context reference in replay submission payloads for competition flows.

Service tasks:

- [ ] Validate replay participants/teams against context roster and scheduling constraints.
- [ ] Route mismatches to `needs_review` and prevent auto-finalization.

Testing tasks:

- [ ] Integration tests for valid and invalid context attachments.
- [ ] Regression tests for configured time-window enforcement.

## Invariant 8: Atomic Finalization

Schema tasks:

- [ ] Ensure result rows, stat rows, attribution rows, and replay terminal state updates are transaction-compatible.

API tasks:

- [ ] Expose single terminal outcome for replay processing (`finalized`, `needs_review`, `rejected`).

Service tasks:

- [ ] Commit participant resolution, result derivation, and stat persistence in one DB transaction.
- [ ] Roll back full transaction on any failure.

Testing tasks:

- [ ] Integration test induced failure rollback across all write targets.
- [ ] Retry tests confirming no partial side effects after failure.

## Invariant 9: Explicit Exception States

Schema tasks:

- [ ] Add replay processing state enum/constraint and failure reason catalog.
- [ ] Add transition timestamps (`parsed_at`, `resolved_at`, `finalized_at`, `reviewed_at`).

API tasks:

- [ ] Add replay processing status endpoint with state and reason details.

Service tasks:

- [ ] Implement explicit state machine with no silent-drop paths.
- [ ] Add manual review resolution actions with audit logging.

Testing tasks:

- [ ] Unit tests for valid/invalid lifecycle transitions.
- [ ] Integration tests for parser failure, unresolved identity, and conflict states.

## Invariant 10: End-to-End Observability

Schema tasks:

- [ ] Add `replay_ingestion_events` or equivalent append-only audit/event log.

API tasks:

- [ ] Add operator endpoint for replay evidence timeline and decision trace.

Service tasks:

- [ ] Emit structured logs with replay evidence ID, parser version, resolution outcome, and actor.
- [ ] Emit metrics for ingest latency, parse failures, unresolved rates, and duplicate rate.

Testing tasks:

- [ ] Integration tests verifying event/audit records for every terminal path.
- [ ] Contract tests for operator payload completeness.

## Cross-Cutting Tasks

Configuration and rollout:

- [ ] Centralize provider enablement, context validation windows, and override policy in versioned config.
- [ ] Gate replay auto-finalization behind feature flags during rollout.

Security and privacy:

- [ ] Define retention and redaction policy for replay files and account-link metadata.
- [ ] Restrict privileged reassignment/reprocessing actions by role.

Operations:

- [ ] Build runbook for replay failure triage and manual resolution.
- [ ] Add alerts for parse failure spikes, resolution degradation, and ingest backlog growth.

Docs:

- [ ] Keep `docs/replay-parsing/unified-replay-parsing-core-concepts.md` aligned with production behavior.
- [ ] Keep `docs/replay-parsing/unified-replay-parsing-invariants-and-guardrails.md` and ADR docs synchronized.

## Suggested Delivery Sequence

1. Evidence and parsing foundations (`Invariant 1`, `Invariant 2`, `Invariant 6`)
2. Identity and association foundations (`Invariant 3`, `Invariant 4`, `Invariant 5`)
3. Competition-context validation and lifecycle safety (`Invariant 7`, `Invariant 8`, `Invariant 9`)
4. Observability, operations, and controlled rollout (`Invariant 10`, cross-cutting tasks)
