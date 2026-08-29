# Sprocket v1 Capability Audit for Kinetic

Date: 2026-07-02  
Status: First-pass migration audit  
Owner: jacbaile

## Purpose

This audit tracks the current Sprocket v1 platform capabilities and whether Kinetic already has the same capability, has an intentional replacement, or needs an explicit "do not port" decision.

The goal is not one-for-one code parity. The goal is capability parity where the capability still matters, and explicit omission where the old capability is no longer part of the product.

## Source Inventory

Primary sources reviewed:

- Sprocket v1 monorepo: `/Users/jacbaile/Workspace/MLE/RocketLeague/sprocket/main`
- Sprocket v1 extracted inventory: `docs/kinetic-v1-analysis/level1-spec.md`, `level2-system-design.md`, `level3-architecture-intent.md`
- MLEDB repo: `/Users/jacbaile/Workspace/MLE/RocketLeague/MLEDB`
- Public datasets repo: `/Users/jacbaile/Workspace/MLE/datasets`
- Kinetic current API/backend/web: `api/openapi.yaml`, `internal/platform/http/*`, `internal/domain/*`, `web/client/*`
- Kinetic planning and intent docs: `docs/interface-design.md`, `plans/gap-closure-plan-2026-03.md`, `plans/web-client-delivery-plan.md`, `plans/pareto-recovery-plan.md`
- Sprocket evidence/report-card note: `/Users/jacbaile/Workspace/MLE/RocketLeague/sprocket/main/reports/report-cards-evidence-design.md`
- Kinetic release evidence artifact: `artifacts/release-validation/local/20260701T142930Z/summary.md`

Important source caveat: the `docs/kinetic-v1-analysis/level*.md` files are useful as Sprocket v1 raw inventory, but they still describe the old NestJS/Svelte/RMQ platform shape. Current Kinetic in this repo is the Go/React REST platform.

## Status Legend

| Status | Meaning |
|---|---|
| `COVERED` | Kinetic has the capability at a usable MVP level. |
| `PARTIAL` | Kinetic has the core shape but meaningful behavior is missing. |
| `REPLACED` | Kinetic intentionally supports the outcome through a different mechanism. |
| `OMIT` | Do not port unless a new product need reopens it. |
| `DECIDE` | Needs an explicit product/ops decision before implementation. |
| `GAP` | Required capability with no adequate Kinetic equivalent yet. |

## Capability Register

| ID | Capability | Sprocket v1 capability summary | Kinetic status | Audit decision | Next action |
|---|---|---|---|---|---|
| CAP-001 | Runtime topology and API contracts | Nest core, GraphQL, RMQ services, Redis, Celery, MinIO, Influx, Discord bot, image generation service | `REPLACED` | Accept Go modular monolith + REST as the Kinetic shape | Keep old topology as reference only |
| CAP-002 | Identity, sessions, and RBAC | Discord OAuth/JWT, refresh, org-team guards, admin login-as-user | `PARTIAL` | Keep Kinetic session/API-token/RBAC model | Finish scope-aware route enforcement and decide on impersonation |
| CAP-003 | League, org, franchise, club, team, player hierarchy | Organizations, members, franchises, game skill groups, teams, member profiles | `PARTIAL` | Port core hierarchy; omit Discord guild mapping as product dependency | Fill profile/branding gaps only if native UI needs them |
| CAP-004 | Games, modes, features, platforms | Game/game-mode GraphQL, feature toggles, platform lookup | `DECIDE` | Kinetic currently models queues and skill groups, not games/modes as first-class API | Decide whether multi-game/mode admin remains in scope |
| CAP-005 | Player intake and platform account association | Register user, create/intake players, report/relink platform accounts, MLEDB mirrors | `PARTIAL` | Keep player activation + platform links; replace legacy repair scripts with support workflows | Add verified OAuth platform linking if still required |
| CAP-006 | Roster governance, transactions, waivers, RFA, salary caps | MLEDB services and bot commands for offers, sign/release, waivers, RFA, roster lock, salary caps | `DECIDE` | Kinetic covers direct roster/role mutation, not league-market workflows | Explicitly accept omission or add a later roster-market epic |
| CAP-007 | Queue and scrim lifecycle | Create/LFS/join/leave/check-in/cancel/lock/disable scrims, queue bans, subscriptions | `PARTIAL` | Keep Kinetic queue/scrim core; replace Discord delivery with web/API | Decide LFS, lock, disable toggle, group queueing, and live updates |
| CAP-008 | Replay upload, parsing, validation, ratification, finalization | GraphQL file upload, MinIO, Redis submission state, Celery parser, validation, reset, force-save, finalization to stats/MLEDB | `PARTIAL` | Keep Kinetic replay evidence model but it is not parser-equivalent yet | Integrate real parser/object storage/participant resolution |
| CAP-009 | Results, scheduled matches, NCP/overrides | Schedule groups, fixtures, matches, report-card triggers, NCP series/round mutation, resubmit jobs | `PARTIAL` | Kinetic result submissions and overrides cover NCP MVP | Add full match scheduling lifecycle and round-level invalidation only if needed |
| CAP-010 | Ratings, Elo, eligibility, skill groups | SprocketRating, Elo service, MLEDB EloData, rankouts, salary updates, eligibility points | `PARTIAL` | Keep Kinetic ratings/skill groups/eligibility, do not assume Elo parity | Decide salary/rankout continuation; add automated rating updates |
| CAP-011 | Stats, standings, and read-only reporting | Sprocket/MLEDB stat tables, public datasets, Evidence pages | `REPLACED` / `PARTIAL` | Evidence remains source for read-heavy analytics | Plan dataset migration to Kinetic schema before Sprocket DB retirement |
| CAP-012 | Report card image generation | Image generation microservice/frontend, MinIO report-card assets, datasets integration | `OMIT` / `DECIDE` | Current Kinetic does not port image-generation service | Confirm whether static report-card PNGs are still a league requirement |
| CAP-013 | Discord bot, webhooks, role sync, notifications | Discord command marshal, DMs, guild/webhook messages, role/nickname sync | `REPLACED` | Discord bot removed by design; Kinetic uses web/API and player notifications | Keep Discord out unless community ops explicitly reverses decision |
| CAP-014 | Support and operator workflows | Admin scrim/submission/restriction tools, Discord/MLEDB support commands | `COVERED` / `PARTIAL` | Kinetic operator inbox is a stronger direction for high-friction ops | Backfill missing support actions from actual support incidents |
| CAP-015 | MLEDB bridge and legacy compatibility | Embedded MLEDB models, bridge tables, core services, standalone MLEDB app/jobs/bot | `DECIDE` | Kinetic currently has no MLEDB bridge | Decide data migration boundary: hard cutover vs compatibility bridge |
| CAP-016 | Public datasets and Evidence publishing | `sprocket-public-datasets`, Sprocket/MLEDB SQL queries, `sprocketdb` DuckDB bundle | `PARTIAL` | Kinetic web embeds Evidence, but datasets still target old schemas | Add Kinetic-compatible datasets or compatibility views |
| CAP-017 | Web client role surfaces | Svelte admin/scrims/league/submit routes with GraphQL live stores | `PARTIAL` | React role dashboards cover MVP actions but not all v1 UX | Keep building role flows from Kinetic API, not by cloning Svelte |
| CAP-018 | Operations, CI, release evidence | Docker/Nest health, local scripts, ad hoc checks | `COVERED` | Kinetic has stronger health, k8s, quality gates, release evidence | Maintain release evidence as promotion gate |
| CAP-019 | Analytics and event fan-out | RMQ event bus, server analytics service, Influx points | `PARTIAL` / `OMIT` | Kinetic favors DB state and release evidence over service event mesh | Add event/outbox only when a real async consumer exists |

## Detailed Capability Notes

### CAP-001 Runtime Topology and API Contracts

Sprocket v1 evidence:

- `core/src/app.module.ts` wires GraphQL, Bull, many feature modules, and service connectors.
- `common/src/service-connectors/core/core.types.ts` and `matchmaking.types.ts` define RMQ endpoint contracts.
- `docs/kinetic-v1-analysis/level2-system-design.md` identifies GraphQL over HTTP/WebSocket, RMQ RPC/pub-sub, Bull/Redis, Celery, MinIO, and InfluxDB.

Kinetic evidence:

- `README.md` and `docs/adr/001-architecture.md` define a Go modular monolith with PostgreSQL.
- `internal/platform/http/route_registrar.go` registers REST route groups.
- `api/openapi.yaml` documents the current REST surface.

Decision:

`REPLACED`. Do not port Sprocket's distributed runtime by default. Preserve the old contracts as capability evidence, not as implementation architecture.

### CAP-002 Identity, Sessions, and RBAC

Sprocket v1 evidence:

- `core/src/identity/auth/oauth/oauth.controller.ts` provides Discord login/refresh.
- `core/src/identity/user/user.resolver.ts` includes `me`, `registerUser`, `getUserByAuthAccount`, and admin `loginAsUser`.
- `core/src/mledb/mledb-player/mle-organization-team.guard.ts` gates admin/league-ops workflows by MLE organization teams.

Kinetic evidence:

- `internal/platform/http/routes_auth.go` provides `/v1/auth/providers`, `/v1/auth/login`, `/v1/auth/callback`, `/v1/auth/logout`, and `/v1/session`.
- `internal/platform/http/routes_self_tokens.go` provides `/v1/api-tokens`.
- `internal/domain/authz/resources.go` defines the resource/action taxonomy.
- `internal/domain/authz/evaluator.go` defines `AllowedInContext`, but current `checkPermission` calls it with `GlobalContext()` and no scoped roles.

Decision:

`PARTIAL`. Kinetic has a cleaner session/API-token model and a first-class resource taxonomy, but scoped authorization is not fully enforced at every route boundary. Admin impersonation is not ported and should be a deliberate decision because it is high risk.

### CAP-003 League and Organization Hierarchy

Sprocket v1 evidence:

- `core/src/organization/*`, `core/src/franchise/*`, and `core/src/database/mledb/*` represent organizations, members, franchises, game skill groups, teams, and MLE legacy entities.
- `CoreEndpoint` includes organization profile, guild mapping, franchise profile, game skill group profile, and transaction webhook lookups.

Kinetic evidence:

- `internal/domain/hierarchy/models.go` includes `League`, `Franchise`, `Club`, `Team`, `Player`, `RosterMembership`, and `RoleAssignment`.
- `internal/platform/http/routes_hierarchy.go` exposes `/v1/leagues`, `/v1/franchises`, `/v1/clubs`, `/v1/teams`, `/v1/players`, `/v1/roster-memberships`, and `/v1/role-assignments`.
- `internal/domain/orgconfig/models.go` and `/v1/leagues/{id}/config` provide league-scoped config values.

Decision:

`PARTIAL`. The structural model is present. Sprocket's organization profile, Discord guild mapping, photos, profile metadata, and webhook routing are not ported. That is acceptable if Discord and rich public profile management are not Kinetic-native requirements.

### CAP-004 Games, Modes, Features, and Platforms

Sprocket v1 evidence:

- `core/src/game/*` exposes games, game modes, platforms, and feature toggles.
- Web queries include `GamesAndModes.store.ts` and game feature mutations.
- MLEDB and datasets encode mode, platform, league, and mode preference.

Kinetic evidence:

- Kinetic has queues, skill groups, platform account links, and rating context keys.
- There is no current first-class `/v1/games`, `/v1/game-modes`, or feature-toggle API.

Decision:

`DECIDE`. If Kinetic remains Rocket League only for the next launch, this can stay omitted. If multi-game operations are a real Kinetic promise, this is a gap and should become a compact game/mode/config slice.

### CAP-005 Player Intake and Platform Account Association

Sprocket v1 evidence:

- `PlayerResolver` supports bulk skill-group changes, player creation, bulk intake, account swaps, force-team, and name changes.
- `MemberModResolver` supports `reportMemberPlatformAccount` and `relinkPlatformAccount`, mirroring to MLEDB player accounts.
- MLEDB has `PlayerService`, `PlayerAccountService`, platform enums, mode preference, timezone, and suspension data.

Kinetic evidence:

- `/v1/players` and `/v1/players/{id}/activate|deactivate` cover player creation and support activation.
- `/v1/platform-accounts`, `/v1/platform-accounts/link`, and `/v1/platform-accounts/unlink` cover manual platform account links.
- `/v1/eligibility` and `/v1/me/eligibility` expose eligibility status.

Decision:

`PARTIAL`. Kinetic has the MVP support path, but it does not yet have OAuth-verified platform account ownership, application review, bulk CSV intake, account repair workflows, or MLEDB mirroring. The old relink/force-change tools should not be copied blindly; each should become a support workflow only when the support team still needs it.

### CAP-006 Roster Governance, Transactions, Waivers, RFA, and Salary Caps

Sprocket v1 and MLEDB evidence:

- Standalone MLEDB includes `TransactionService`, `WaiverService`, `SalaryCapService`, `RosterLockService`, `DraftOrderService`, and RFA/waiver workers.
- MLEDB bot commands include `Transactions.command.ts`, `Waviers.command.ts`, `Rankouts.command.ts`, `ScheduleMatch.command.ts`, and `PreventRankout.command.ts`.
- Sprocket core player models include salary and game skill group salary caps.

Kinetic evidence:

- Kinetic supports roster membership create/list and role assign/revoke.
- React admin surfaces include roster membership and role delegation workflows.
- There is no offer, acceptance, waiver priority, RFA, draft-order, roster-lock, or salary-cap state machine.

Decision:

`DECIDE`. Current Kinetic covers direct admin mutation, not league-market automation. This is probably acceptable for MVP only if MLE can run transactions/waivers outside Kinetic or defer them. If the league relies on these workflows in season operations, create a dedicated roster-market epic instead of hiding them inside roster CRUD.

### CAP-007 Queue and Scrim Lifecycle

Sprocket v1 evidence:

- `ScrimModuleResolver` supports scrim metrics, all/available/LFS/current scrims, create scrim, create LFS scrim, join, leave, check-in, cancel, lock, and subscriptions.
- `ScrimToggleResolver` supports global scrims-disabled state.
- `MemberRestrictionResolver` supports queue bans/restrictions.
- `MatchmakingEndpoint` includes `CreateLFSScrim`, `CreateScrim`, `JoinScrim`, `LeaveScrim`, `CheckInToScrim`, `CompleteScrim`, `CancelScrim`, `SetScrimLocked`, and `UpdateLFSScrimPlayers`.

Kinetic evidence:

- `routes_queue_scrim.go` exposes `/v1/queues`, `/v1/queue-entries`, `/v1/queue-bans`, `/v1/queue-bans/lift`, `/v1/scrims`, `/v1/scrim-promotions`, `/v1/scrim-promotions/process`, `/v1/promotion-processing-runs`, `/v1/player-ratings`, and `/v1/matchmaking-decisions`.
- `routes_admin_mutations.go` exposes `GET /v1/scrims/{id}`.
- OpenAPI includes `POST /v1/scrims/{id}/check-in` and `/v1/scrim-metrics`.
- `plans/gap-closure-plan-2026-03.md` documents replacing Discord scrim notifications with `/v1/me/scrim`, `/v1/me/queue-bans`, and player notifications.

Decision:

`PARTIAL`. Kinetic preserves the important queue/scrim lifecycle, promotion observability, queue bans, pop timeout, and check-in direction. Missing v1 features: LFS scrims, group/party queueing, `leaveAfter`, global scrim disable, explicit lock/unlock, complete endpoint, and live GraphQL/WebSocket subscriptions. Polling plus notifications is acceptable unless live scrim UX requires push.

### CAP-008 Replay Upload, Parsing, Validation, Ratification, and Finalization

Sprocket v1 evidence:

- `ReplayParseModResolver` supports `getSubmission`, `parseReplays`, `mockCompletion`, `resetSubmission`, `forceSubmissionSave`, `ratifySubmission`, `rejectSubmission`, `validateSubmission`, and `followSubmission`.
- `microservices/replay-parse-service` includes the Python parser service and a real replay fixture `17A7C1084017DFA7DBE66D9C66D81CBD.replay`.
- Sprocket uses MinIO for replay objects, Redis for submission state, Celery for parsing, and finalization subscribers for durable stats and Elo follow-up.

Kinetic evidence:

- `/v1/replay-evidence`, `/v1/replay-parse-runs`, and `/v1/result-submission-replay-links` model evidence, parser provenance, and submission links.
- `ReplayStore.TriggerReplayParse` is explicitly described as a background stub parse.
- `docs/adr/017-replay-evidence-and-parser-provenance-mvp.md` states replay bytes are an MVP body payload, not external object storage, and parse status is simplified.
- Release evidence covers replay intake, dedupe, and cross-context rejection.

Decision:

`PARTIAL`. Kinetic has the correct evidence/provenance model but not v1 parser parity. Real parity requires file/object storage, real replay parser integration, parse failure states, participant identity resolution, validation, retry/reset semantics, and atomic finalization.

### CAP-009 Results, Scheduled Matches, and NCP/Overrides

Sprocket v1 evidence:

- `MatchResolver` supports schedule/match lookup, report-card triggering, match reprocessing, series NCP, round NCP, and match submission status computation.
- `SubmissionManagementResolver` supports admin active submissions and admin reset.
- Sprocket's match finalization maps Sprocket match state to MLEDB series/replay state.

Kinetic evidence:

- `/v1/seasons`, `/v1/schedule-groups`, `/v1/fixtures`, `/v1/fixtures/{id}`, and `/v1/matches` provide scheduling CRUD/read MVP.
- `/v1/result-submissions`, `/v1/result-submission-ratifications`, `/v1/result-submission-rejections`, `/v1/result-submissions/{id}/reset`, and `/v1/result-overrides` provide result lifecycle and NCP-style override.
- `ResultOverride` stores previous/new winners, state, actor, and reason.

Decision:

`PARTIAL`. Kinetic has the right MVP result/override model. It does not yet cover full scheduled-time proposal/acceptance, match status computation from submission service, round-level invalidation, report-card event triggering, or bulk match reprocessing.

### CAP-010 Ratings, Elo, Eligibility, and Skill Groups

Sprocket v1 and MLEDB evidence:

- Sprocket has `SprocketRatingService` for OPI/DPI/GPI.
- Core Elo resolver and connector trigger external Elo jobs and process manual skill-group changes.
- MLEDB has `EloData`, `EloService`, `EloWorkerService`, rankout processing, salary updates, no-rankdown lists, and eligibility data.

Kinetic evidence:

- `/v1/player-ratings`, `/v1/player-ratings/adjust`, `/v1/rating-adjustments`, and `/v1/matchmaking-decisions` exist.
- `/v1/leagues/{id}/skill-groups`, `/v1/skill-groups/{id}`, and `/v1/players/{id}/skill-group-transitions` exist.
- `ReplayStatsStore` computes OPI/DPI/GPI from persisted stat lines.
- `/v1/eligibility` and `/v1/me/eligibility` expose eligibility projection.

Decision:

`PARTIAL`. Kinetic has a cleaner rating/skill-group foundation and manual audited adjustments. It is not equivalent to MLEDB Elo/rankout/salary automation. Decide whether salary and rankout remain product concepts before implementing them.

### CAP-011 Stats, Standings, and Read-only Reporting

Sprocket v1 and datasets evidence:

- Sprocket and MLEDB persist player/team stat lines, rounds, scrims, series, standings, eligibility, salary, role usage, and bridge identifiers.
- `/Users/jacbaile/Workspace/MLE/datasets/queries/public` contains SQL for players, members, teams, leagues, standings, scrim stats, eligibility data, report cards, and season-specific stats.
- `docs/interface-design.md` and `plans/web-client-delivery-plan.md` explicitly state that Evidence remains the source of truth for read-heavy views.

Kinetic evidence:

- `web/client/src/features/player` embeds Evidence views.
- `web/client/tests/e2e/cuj-stats-history.spec.ts` checks Evidence views for standings, ratings, and eligibility.
- Kinetic has `/v1/result-submissions/{id}/stats` and `/v1/players/{id}/career-stats`.

Decision:

`REPLACED` for read-heavy UI, `PARTIAL` for data production. Kinetic should not duplicate every Evidence page in React, but the datasets currently query Sprocket/MLEDB schemas. Before Sprocket DB retirement, either migrate datasets to Kinetic tables or provide compatibility views.

### CAP-012 Report Card Image Generation

Sprocket v1 evidence:

- `clients/image-generation-frontend` provides report/template operator tooling.
- `microservices/image-generation-service` renders SVG/PNG assets.
- `reports/report-cards-evidence-design.md` defines `report_card_asset`, public bucket URLs, and datasets for scrim/match report cards.
- `CoreEndpoint.GenerateReportCard` and `UpsertReportCardAsset` are part of the old service contract.

Kinetic evidence:

- Kinetic has no image-generation service, image-generation frontend, report-card asset table, or report-card trigger.
- Kinetic exposes stat lines/career stats and embeds Evidence views.
- `plans/gap-closure-plan-2026-03.md` lists image/report-card generation service as explicitly out of scope.

Decision:

`OMIT` unless the league still depends on static report-card PNGs. If report cards are required, implement a new small pipeline around Kinetic result/stat data; do not port the old full image-generation service without a narrowed requirement.

### CAP-013 Discord Bot, Webhooks, Role Sync, and Notifications

Sprocket v1 and MLEDB evidence:

- `clients/discord-bot` handles commands, notifications, guild/member sync, role/nickname updates, DMs, and webhooks.
- `BotEndpoint` supports guild text, direct message, and webhook message delivery.
- MLEDB bot commands cover administration, transactions, waivers, scrims, replays, rankouts, scheduling, streams, and player/team info.

Kinetic evidence:

- `plans/gap-closure-plan-2026-03.md` explicitly removes Discord bot and notification service as v3 dependencies.
- `/v1/me/*` endpoints and `player_notifications` replace Discord-delivered state with web/API-visible state.
- `web/client` role dashboards consume API state instead of Discord commands.

Decision:

`REPLACED`. Kinetic should not port the Discord bot by default. The one risk is operational reality: if Discord remains the actual place league users act, this decision must be revisited with a narrow command/notification list.

### CAP-014 Support and Operator Workflows

Sprocket v1 evidence:

- Admin web routes and Svelte components expose scrim management, active submissions, restricted players, and game features.
- Sprocket/MLEDB Discord commands provide many support repair paths.

Kinetic evidence:

- `/v1/operator-inbox`, `/v1/operator-inbox/triage`, `/v1/operator-inbox/resolve`, `/v1/exception-actions`, `/v1/exception-metrics`, and exception automation endpoints exist.
- `plans/pareto-recovery-plan.md` focuses Kinetic on reducing weekly operator burden.
- `web/client/src/features/support` and `platform` pages provide support/operator dashboards.

Decision:

`COVERED` for the strategic direction, `PARTIAL` for specific historical repair commands. Add missing support actions only when backed by actual current incidents. Avoid porting every old Discord repair command.

### CAP-015 MLEDB Bridge and Legacy Compatibility

Sprocket v1 and MLEDB evidence:

- Sprocket embeds MLEDB models and bridge tables such as player-to-player, team-to-franchise, fixture-to-fixture, series-to-match-parent, league-to-skill-group, and season-to-schedule-group.
- Sprocket finalization writes/reads MLEDB series, scrims, players, stats, NCP state, role usage, and stakeholders.
- Standalone MLEDB remains a large operational system with GraphQL, bot commands, workers, and domain services.

Kinetic evidence:

- Kinetic has no MLEDB bridge tables or MLEDB write path.
- `plans/gap-closure-plan-2026-03.md` lists the MLEDB compatibility bridge as out of scope.

Decision:

`DECIDE`. This is the largest strategic migration boundary. If Kinetic is a hard cutover, omission is acceptable but datasets/Evidence and any MLEDB-only operations must be retired or migrated. If Sprocket/MLEDB must coexist during a season, a small compatibility bridge is required.

### CAP-016 Public Datasets and Evidence Publishing

Datasets evidence:

- `/Users/jacbaile/Workspace/MLE/datasets/queries/public` includes public SQL datasets for members, players, teams, leagues, standings, schedules, scrim stats, eligibility, role usages, and report cards.
- The datasets repo publishes `sprocket-public-datasets` assets and a `sprocketdb` DuckDB bundle.
- Queries currently depend heavily on `sprocket.*`, `mledb.*`, and `mledb_bridge.*` schemas.

Kinetic evidence:

- Web client embeds Evidence views, but Kinetic does not yet generate compatible public datasets.
- Kinetic has core API endpoints and stat tables that can support a future dataset layer.

Decision:

`PARTIAL`. Evidence integration is accepted. Dataset production is not migrated. Add a dataset migration track before Sprocket/MLEDB schemas stop being available.

### CAP-017 Web Client Role Surfaces

Sprocket v1 evidence:

- Svelte routes include `/admin`, `/scrims`, `/league`, `/league/[fixtureId]`, `/league/scrim/[submissionId]`, and `/league/submit/[submissionId]`.
- GraphQL stores and subscriptions drive live scrim/submission views.

Kinetic evidence:

- React routes cover login, player, support, admin, platform operator, and unauthorized pages.
- E2E tests cover login landing, scrim flow, league match state, rosters, support operations, player skill/rating, stats history, auth guard, and release evidence.

Decision:

`PARTIAL`. Kinetic covers the main role surfaces but not all v1 UX depth. Keep building from Kinetic's current API and role model; do not clone the Svelte implementation structure.

### CAP-018 Operations, CI, and Release Evidence

Sprocket v1 evidence:

- Sprocket has Docker/Nest service health checks and local scripts, but operational proof historically depended on ad hoc service checks and hosted CI.

Kinetic evidence:

- `README.md` documents `quality-gate.sh`, `web-quality-gate.sh`, `release-evidence.sh`, smoke scripts, and k8s local dev.
- `artifacts/release-validation/local/20260701T142930Z/summary.md` shows passing credentialed CORS, session isolation, privileged negative control, replay intake/dedupe/cross-context rejection, and browser API-mode login.
- `docs/runbooks/*` cover config, dev setup, operations/rollback, and release readiness.

Decision:

`COVERED`. This is a Kinetic improvement over v1. Keep release evidence as a required promotion artifact.

### CAP-019 Analytics and Event Fan-out

Sprocket v1 evidence:

- Sprocket uses RMQ events, analytics service, Influx points, notification-service, and image/report-card event triggers.
- Core operations call analytics for account-link/reporting workflows.

Kinetic evidence:

- Kinetic stores operational state and exception metrics in PostgreSQL and exposes API/read models.
- No event bus, Influx, notification-service, or analytics microservice is currently present.

Decision:

`PARTIAL` / `OMIT`. Kinetic does not need Sprocket's event mesh until there is a real async consumer. If future needs include outbound notifications, report-card generation, dataset materialization, or audit streaming, add an outbox/event table first rather than reintroducing RMQ by default.

## Priority Follow-up List

These are the highest-leverage unresolved items from the audit.

| Priority | Item | Capability IDs | Recommendation |
|---|---|---|---|
| P0 | Decide MLEDB/datasets migration boundary | CAP-011, CAP-015, CAP-016 | Choose hard cutover, compatibility views, or a bridge before Sprocket DB is retired. |
| P0 | Close real replay parser parity | CAP-008, CAP-010, CAP-011 | Integrate parser/object storage/participant resolution and atomic stat/rating finalization. |
| P1 | Finish scoped authorization enforcement | CAP-002, CAP-006, CAP-009 | Route-level scope checks must match FM/GM/AGM/Captain authority. |
| P1 | Decide roster-market scope | CAP-006, CAP-010 | Explicitly omit or ticket transactions, waivers, RFA, roster lock, salary caps, and rankouts. |
| P1 | Decide LFS/live scrim features | CAP-007, CAP-017 | Either omit LFS/lock/disable/live subscriptions or add Kinetic-native endpoints. |
| P2 | Decide report-card image future | CAP-012, CAP-016 | If static PNG report cards remain required, build a narrower Kinetic report-card pipeline. |
| P2 | Decide games/modes API | CAP-004 | Needed only if Kinetic launch requires real multi-game/mode administration. |

## Audit Conclusion

Kinetic is not missing the core platform skeleton. It already covers the main modernized direction: Go REST API, role-based web surfaces, league hierarchy, roster and role mutation, queue/scrim lifecycle, result submission and overrides, replay evidence MVP, operator inbox, player notifications, API tokens, k8s/runbooks, and release evidence.

The risky gaps are not generic CRUD. The risky gaps are legacy operational dependencies:

1. MLEDB and public datasets still own a large amount of read/reporting and league-office behavior.
2. Replay parsing/finalization is not yet equivalent to Sprocket v1's real parser pipeline.
3. Roster-market operations such as transactions, waivers, RFA, salary caps, and rankouts are not represented.
4. Discord and report-card image generation are intentionally omitted, but those decisions should be revalidated against actual league operations.

If those four areas are explicitly accepted or planned, the rest of the audit is manageable incremental closure rather than a surprise platform rewrite.
