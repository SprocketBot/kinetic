# Sprocket Functionality Cutover Plan

Date: 2026-07-22

## Scope

This plan closes the Sprocket v1 capability gaps that remain relevant to Kinetic.

The following are intentional omissions and are not migration work:

- MLEDB compatibility or a dual-write bridge.
- Dataset or Evidence replacement. Those remain separate systems whose interface is the database.
- Sprocket's NestJS/GraphQL/RMQ/Redis topology.

Discord integration and report-card image generation are not assumed either way. They get explicit product decisions below. If retained, they should be implemented as Kinetic-native integrations, not copied as Sprocket infrastructure.

## Outcome

Kinetic should be able to run league operations independently of Sprocket and MLEDB:

1. authenticate and authorize users in league/franchise/team context;
2. manage player identities, platform accounts, rosters, and league governance;
3. run queue, scrim, scheduled-match, replay, result, and ratification lifecycles;
4. derive stats, ratings, eligibility, and skill-group transitions from authoritative replay evidence;
5. provide operator recovery paths and auditable state transitions.

## Workstreams and order

### Phase 0 — Freeze contracts and decisions

Create a capability-to-Kinetic contract matrix from the audit and mark every item as `build`, `replace`, or `omit`.

Decide before implementation:

- whether roster-market operations are required for the first Kinetic season;
- whether Discord needs a narrow notification/command integration;
- whether static report cards remain a league requirement;
- whether multi-game/multi-mode administration is in scope.

Lock the REST/OpenAPI shapes, state machines, authorization contexts, and ownership rules before adding UI or workers.

Acceptance evidence: signed-off decision record, updated OpenAPI contract, and no unresolved `DECIDE` items for launch-critical capabilities.

### Phase 1 — Authorization and identity foundations

Complete the security boundary first because every later workflow depends on it.

- Make authorization checks context-aware for league, franchise, club, team, and player resources.
- Enforce FM/GM/AGM/Captain scope at every mutation and sensitive read route.
- Define approval/effective-state semantics for role assignments.
- Finish verified platform-account ownership flows for Steam, Epic, Xbox, and PSN where supported.
- Add player intake, inactive-account review, relink, account-conflict, and bulk-import support workflows.
- Preserve immutable identity attribution when a platform link changes.

Acceptance evidence: negative tests proving cross-team access is rejected, verified account-link tests, and operator tests for identity conflicts.

### Phase 2 — Real replay pipeline and authoritative stats

Replace the current stub parse path with the durable evidence pipeline.

- Accept and hash raw replay files through the ingestion boundary.
- Queue parsing through a worker abstraction that can invoke the Rust/Python parser.
- Persist parser name, version, configuration digest, status, failure reason, and retry history.
- Resolve replay participants using provider/account identifiers and verified platform links.
- Route unresolved or conflicting identities to the operator inbox.
- Extract rounds, player stat lines, team stat lines, and authoritative game outcomes.
- Validate replay participants and teams against the scrim or scheduled-match context.
- Make finalization transactional and idempotent: result, stats, attribution, rating inputs, and terminal replay state must commit together.
- Support retry, reset, reject, review, and operator override paths.
- Keep replay evidence immutable and retain an audit timeline for every decision.

Acceptance evidence: real replay fixtures, parser failure/retry tests, identity mismatch tests, duplicate/context-conflict tests, and an end-to-end finalized result with stats.

### Phase 3 — Results, scheduling, and competition parity

Complete the operational lifecycle around the authoritative replay/result model.

- Add scheduled-time proposal, counterproposal, acceptance, and ratification.
- Compute match readiness and submission status from the actual lifecycle state.
- Support round-level and match-level invalidation/NCP semantics.
- Add resubmission and bulk reprocessing with idempotency and audit records.
- Trigger downstream stat/rating/eligibility work only from valid terminal results.
- Complete scrim transitions: explicit lock/unlock, completion, cancellation, timeout, and no-show handling.
- Add LFS, party/group queueing, and live updates only if Phase 0 confirms they are launch requirements; otherwise record them as post-launch extensions.

Acceptance evidence: browser and API tests for proposal/ratification, invalidation, resubmission, scrim completion, no-show, and replay-to-match finalization.

### Phase 4 — Ratings, eligibility, and competitive governance

Build automated competitive consequences on top of finalized stats.

- Finalize the unified rating algorithm and its configuration contract.
- Apply rating changes automatically after valid ratification/finalization.
- Record every adjustment with actor, reason, previous value, new value, and source submission.
- Evaluate skill-group promotion/demotion with hysteresis.
- Persist skill-group transitions and notify affected players in-app.
- Implement eligibility points, decay, suspension/restriction effects, and match-time eligibility checks.
- Add rankout/no-rankdown behavior if confirmed in Phase 0.
- Keep salary calculation out unless the roster-market decision explicitly reopens it.

Acceptance evidence: deterministic rating fixtures, promotion/demotion boundary tests, replay finalization integration tests, and eligibility decision audits.

### Phase 5 — Roster governance and league-office workflows

If required for launch, add the missing state machines rather than expanding roster CRUD.

- player offers and acceptance;
- release/transfer workflows;
- waiver submission, priority, and resolution;
- RFA handling;
- draft order and draft selections;
- roster locks and effective dates;
- role/roster history and conflict resolution;
- salary caps only if the league still operates them in Kinetic.

Acceptance evidence: concurrent-action tests, authorization tests by role and scope, audit-history checks, and operator recovery flows.

### Phase 6 — Integration replacements and operational completion

Implement only the integrations approved in Phase 0.

- If Discord remains needed, provide a narrow outbound notification/webhook adapter over Kinetic state and an outbox, with no Discord dependency in core transactions.
- If report cards remain needed, build a small renderer consuming Kinetic stats and expose assets through a separate service/storage boundary.
- Add game/mode/platform configuration entities and APIs if multi-game support is confirmed.
- Complete operator dashboards for replay failures, identity mismatches, scheduling conflicts, roster conflicts, and rating corrections.
- Add runbooks, retention rules, feature flags, rollout metrics, and rollback procedures.

Acceptance evidence: release-evidence automation covers each approved integration and all launch-critical workflows have browser-level proof.

## Dependency graph

```text
Phase 0 decisions
        |
        v
Authorization + identity
        |
        v
Replay parsing + participant resolution
        |
        v
Stats + result finalization
       / \
      v   v
Scheduling  Ratings/eligibility
      \   /
       v v
 Roster governance and operator workflows
        |
        v
 Approved integrations and release gate
```

## Recommended priority

`P0`: Phase 0, Phase 1, and Phase 2. Without these, Kinetic cannot safely establish authoritative outcomes.

`P1`: Phase 3 and Phase 4. These complete league competition behavior and automate its consequences.

`P1/P2`: Phase 5, depending on whether transactions and waivers are required for the first season.

`P2`: Phase 6 integrations not required for the hard cutover.

## Definition of done

The cutover is complete when a real player can authenticate, be linked to verified platform identities, enter a scoped roster/queue, participate in a scrim or scheduled match, submit a real replay, have participants resolved, produce durable stats and an authoritative result, ratify or reject it, receive rating and eligibility consequences, and recover failures through the operator workflow—without any runtime dependency on Sprocket or MLEDB.
