# ADR-012: Replay Parsing And Platform Account Association Model

- Status: Accepted
- Date: 2026-02-14
- Owner: jacbaile
- Related: ADR-002 (auth and RBAC baseline), ADR-008 (scheduled match vs scrim), ADR-009 (scheduled competition slice)
- Source Parser: <https://github.com/KineticBot/kinetic-rl-parser>
- Detailed Reference: `docs/replay-parsing/unified-replay-parsing-core-concepts.md`

## Context

Kinetic's league and scrim workflows depend on reliable game results and player statistics derived from replay files.
The same person can appear under multiple in-game platform identities (`Epic`, `Steam`, `Xbox`, `PSN`), while platform login identity and in-game identity are separate concerns.

Without a unified model for account association and replay attribution, replay-driven automation becomes error-prone and difficult to audit.

## Decision

Adopt a replay-centric ingestion and identity-resolution model with explicit platform-account association.

### 1. Layered identity model

- `User` is the authentication principal.
- `Player` is the competition entity.
- `Platform Account` is provider-specific in-game identity (`provider + provider_account_id`).

### 2. OAuth-first account reporting

- Users link platform accounts while authenticated.
- Account links are created from verified provider identity responses.
- A player may own multiple linked platform accounts.

### 3. Replay evidence as source of truth

- Replay bytes are ingested as immutable evidence.
- Parsing is delegated to `kinetic-rl-parser`.
- Parsed outputs are normalized into canonical internal structures with parser provenance.

### 4. Participant resolution contract

- Replay participants are resolved through verified platform-account links.
- Name-only matching is non-authoritative.
- Unresolved or conflicting identities route to review, not silent attribution.

### 5. Competition-context association

- Replays must attach to an explicit competition context (`scrim` or scheduled `match`) for automated result application.
- Submission provenance and evidence traceability are retained end-to-end.

### 6. Derivation and reprocessing

- Results and stats are derived from canonical parse outputs.
- Reprocessing is supported for parser/version corrections with explicit audit trails.

### 7. Explicit processing lifecycle

- Replay processing uses explicit lifecycle states (for example `received`, `parsed`, `resolved`, `finalized`) with non-terminal review states.

## Consequences

### Positive

- Improves trustworthiness of automated standings/stats.
- Reduces manual score-reporting burden.
- Preserves clear evidence trails for disputes and corrections.

### Negative / Tradeoffs

- Requires careful provider-linking and reassignment controls.
- Increases need for robust ingestion observability and review tooling.
- Introduces schema/service complexity around identity and replay lifecycle state.

## Non-Goals (Current Decision Scope)

- Replacing `kinetic-rl-parser` internals.
- Defining full UI/UX flows for every replay-management screen.
- Final retention policy and cost optimization strategy for replay storage.

## Follow-ups

- Define invariant-level guardrails and change controls (ADR-013).
- Implement replay ingestion checklist by invariant.
- Add operational dashboards and runbooks for replay ingestion quality.
