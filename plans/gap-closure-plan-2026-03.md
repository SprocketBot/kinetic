# Kinetic v3 Gap Closure Plan

Date: 2026-03-01
Status: Draft
Inputs: `docs/kinetic-v1-analysis/`, `/Users/jacbaile/Workspace/MLE/RocketLeague/kinetic_dev`, `plans/web-client-delivery-plan.md`, gap analysis (2026-03-01)

---

## Framing

Two architectural decisions shape this plan:

1. **Discord bot and notification service are out of scope for v3.** The community Discord integration is being removed as a platform dependency. Information that v1 delivered via Discord DM or webhook (lobby credentials, queue ban notices, submission ratification prompts, skill group changes) must instead be surfaceable via the web client and queryable via the API.

2. **A complete, fully-featured RBAC system is a first-class v3 requirement**, modeled on the v2 design: a resource-and-action taxonomy covering the full data model, scope-aware enforcement for organizational roles, and an API token system for machine clients.

---

## Theme 1: RBAC Completeness

### Context

v3 has the structural skeleton (roles, policies, user_role_bindings, role_assignments tables; a static evaluator) but the resource taxonomy is nearly empty and the evaluator ignores the organizational scope stored in `role_assignments`. v2 had 33 explicitly enumerated resources, scope-aware policy enforcement via Casbin domains, an approval workflow for restricted role grants, and API tokens with per-scope permission sets.

### Work Items

**1.1 — Define the v3 resource taxonomy**

Enumerate every resource that the API protects. Align with v3's domain entities (not v1's or v2's naming directly). Suggested starting taxonomy:

| Resource | Example actions |
|---|---|
| `league` | read, create, update, delete |
| `franchise` | read, create, update, delete |
| `club` | read, create, update, delete |
| `team` | read, create, update, delete |
| `player` | read, create, update, delete |
| `roster_membership` | read, create, delete |
| `role_assignment` | read, create, revoke |
| `queue` | read, create, update, delete |
| `queue_entry` | read, create, delete |
| `queue_ban` | read, create, lift |
| `scrim` | read, create, update |
| `scrim_promotion` | create |
| `result_submission` | read, create, ratify, reject, reset |
| `result_override` | read, create |
| `replay_evidence` | read, create |
| `platform_account_link` | read, create, delete |
| `player_rating` | read |
| `rating_adjustment` | read, create |
| `matchmaking_decision` | read |
| `exception_ticket` | read, create, triage, resolve |
| `season` | read, create, update, delete |
| `schedule_group` | read, create, update, delete |
| `fixture` | read, create, update, delete |
| `match` | read, create, update, delete |
| `skill_group` | read, create, update, delete |
| `skill_group_transition` | read |
| `eligibility` | read |
| `player_stat` | read |
| `organization_config` | read, update |
| `api_token` | read, create, revoke |

Deliver this as: a Go `const` or `type` block in `internal/domain/authz/` defining resource and action string constants, and a migration seeding the default policy set for the `admin` and `observer` roles.

**1.2 — Scope-aware authorization enforcement**

The evaluator currently checks `(role, resource, action)` globally. For organizational roles (fm, gm, agm, captain), access to resources should be scoped: a captain can manage their own team's roster memberships and result submissions, not another team's.

Design:
- Add an `AuthzContext` struct carrying the requesting player's ID plus optional `franchise_id`, `club_id`, `team_id` identifiers for the resource being acted on.
- Extend `Allowed(roles, resource, action)` to `AllowedInContext(principal, resource, action, ctx AuthzContext)`.
- For scoped roles, the evaluator checks `role_assignments` for a row where `role = 'captain'`, `team_id = ctx.team_id`, and `is_active = true`. Global policies (`user_role_bindings`) continue to override.
- Update handler middleware to build `AuthzContext` from route parameters before calling `RequirePermission`.

This eliminates the gap where organizational-scope roles grant blanket access (v2 had the same gap; v3 should close it properly).

**1.3 — Role assignment approval workflow**

Some role grants (fm, admin) require approval by a higher-authority actor. Extend `role_assignments`:
- Add `status` column: `pending | active | rejected`.
- Add `reviewed_by_actor_player_id`, `review_reason`, `reviewed_at` columns.
- Add endpoints: `POST /v1/role-assignments/approve`, `POST /v1/role-assignments/reject`.
- Pending assignments are not effective in authorization checks until approved.
- Gate approval authority: only `admin` can approve `fm` grants; `fm` can approve `gm`/`agm`; `gm` can approve `captain`.

**1.4 — API token system**

Add a machine-client token system for integration use cases (stat pipelines, external tooling, CI smoke tests):

Schema additions:
```sql
CREATE TABLE api_tokens (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name        text NOT NULL,
    token_hash  text NOT NULL UNIQUE,        -- bcrypt of the token
    subject     text NOT NULL,               -- player or service identity
    scopes      text[] NOT NULL DEFAULT '{}', -- resource:action pairs
    created_by  uuid REFERENCES players(id),
    expires_at  timestamptz,
    revoked_at  timestamptz,
    last_used_at timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now()
);
```

- Token generated on creation (returned once, never stored in plain text).
- Presented as `Authorization: Bearer spr_<token>`.
- Auth middleware checks api_tokens table as fallback when no session cookie is present.
- Scopes constrain permissions below the subject's role-granted ceiling (additive roles, subtractive scopes).

---

## Theme 2: Discord → API/Web Migration

### Context

v1 used Discord DMs and webhooks as the primary notification surface for player-facing events. Since v3 removes the Discord bot, all information previously delivered via Discord must be queryable via the API so the web client can display it. This does not require a new push notification system — it requires making the existing state machine observable and adding the missing API endpoints that Discord previously made unnecessary.

### Information previously delivered by Discord → required API surface

| v1 Discord event | Required API endpoint |
|---|---|
| Scrim popped: lobby name + password DM | `GET /v1/me/scrim` — returns current scrim with lobby credentials for authenticated participant |
| Queue ban notice | `GET /v1/me/queue-bans` — active bans for session player |
| Submission needs ratification DM | `GET /v1/me/result-submissions?pending_ratification=true` |
| Skill group change notification | `GET /v1/me/eligibility` — eligibility + current rating context |
| Scrim report card | `GET /v1/result-submissions/{id}/stats` — stat lines from finalized submission |
| Player team changed | Surfaced via `GET /v1/me/roster-membership` |

All of these are `GET` queries against existing data. No event push infrastructure is required. The web client uses TanStack Query polling on the relevant endpoints to show the player their current state.

### Work Items

**2.1 — Me-scoped endpoint group (`/v1/me/*`)**

Add a set of session-scoped convenience endpoints that return data filtered to the authenticated player's context. These avoid requiring the client to know their own player ID before making useful queries.

Endpoints:
- `GET /v1/me` — session player profile, active roster membership, active role assignments
- `GET /v1/me/queue-entry` — current queue enrollment (null if not queued)
- `GET /v1/me/scrim` — current scrim with lobby credentials if participant, null if not in a scrim
- `GET /v1/me/result-submissions` — submissions player is a party to, accepts `?state=pending_ratification`
- `GET /v1/me/queue-bans` — active queue bans on session player
- `GET /v1/me/eligibility` — eligibility status, current rating, projected transitions
- `GET /v1/me/roster-membership` — active roster membership

**2.2 — Lobby credential generation**

v1 generated a random lobby name and password when a scrim was popped and delivered them via DM. v3 needs to generate and store these on the scrim record so the web client can display them.

Schema addition to `scrims`:
```sql
ALTER TABLE scrims
    ADD COLUMN lobby_name     text,
    ADD COLUMN lobby_password text,
    ADD COLUMN popped_at      timestamptz;
```

- `lobby_name` and `lobby_password` populated when scrim state transitions to `popped` (new state — see Theme 3).
- `GET /v1/me/scrim` returns lobby credentials only if the requesting player is a participant in the scrim.
- Non-participant reads (operator view) also include credentials.

**2.3 — Player notification inbox (lightweight)**

To replace Discord DM notification awareness without building a full push system, add a simple in-app notification inbox: a list of unread events that the player can acknowledge. This is a pull model, not push.

Schema:
```sql
CREATE TABLE player_notifications (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id   uuid NOT NULL REFERENCES players(id),
    category    text NOT NULL,   -- scrim_popped, queue_ban, ratification_needed, skill_group_change
    context_type text,           -- scrims, result_submissions, etc.
    context_id   uuid,
    message     text NOT NULL,
    read_at     timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now()
);
```

Endpoints:
- `GET /v1/me/notifications` — unread-first list, paginated
- `POST /v1/me/notifications/mark-read` — accepts `{ ids: [...] }` or `{ all: true }`

Notifications are written as a side-effect of state transitions in scrim promotion, queue-ban creation, submission lifecycle, and eligibility evaluation. No external event bus needed — written in the same transaction as the triggering operation.

---

## Theme 3: Scrim State Machine Completion

### Context

v3 has `created → in_progress → closed/voided`. v1 had a richer state machine: `created (queued) → popped (pop timeout begins, participants notified) → in_progress (all checked in, lobby active) → complete/cancelled`. The check-in phase + pop timeout are operationally important: they prevent phantom scrims where one team doesn't show up.

### Work Items

**3.1 — Add check-in states to scrim**

Extend `scrims.state` enum:
```
created → popped → in_progress → closed | voided | cancelled
```

- `popped`: scrim has been promoted from the queue. Participants have N minutes to check in. Lobby credentials generated and stored. Player notifications written.
- `in_progress`: all participants have checked in. Match can begin.
- `cancelled`: pop timeout expired before all participants checked in. Queue ban applied to non-checking-in party.

Add `checked_in_teams` as a tracked set (e.g., `home_checked_in_at timestamptz`, `away_checked_in_at timestamptz`).

**3.2 — Check-in endpoint**

```
POST /v1/scrims/{id}/check-in
```

- Requires session player to be on the home or away team of the scrim.
- Sets `home_checked_in_at` or `away_checked_in_at`.
- If both teams checked in, transitions state to `in_progress`.
- Returns updated scrim.

**3.3 — Pop timeout background worker**

Add a lightweight background goroutine (or `time.AfterFunc` triggered at promotion time) that:
- Fires after `pop_timeout_minutes` (from organization config, defaulting to 5).
- If scrim is still `popped` at expiry, cancels the scrim and writes a queue ban for the non-checking-in team (or both if neither checked in).
- Writes a `player_notification` for affected players.
- Writes an exception ticket of category `no_show` to the operator inbox.

Design note: This does not require a separate task queue. A goroutine launched from the HTTP handler (or the `process` endpoint) with a timer is sufficient for MVP. The scrim state machine is the source of truth; the timer is just a trigger.

**3.4 — Scrim metrics endpoint**

```
GET /v1/scrim-metrics
```

Response: `{ players_queued, teams_in_scrim, open_scrims, scrims_closed_today, avg_wait_seconds_p50 }`

Computed from live queue_entries and scrims tables. No caching required at this scale.

---

## Theme 4: Replay Parsing Pipeline

### Context

v3 accepts pre-parsed JSON but has no pipeline that actually parses replays. Identity resolution (platform account → player) is designed (ADR-012) but not implemented. Stat extraction (PlayerStatLine, TeamStatLine, Round) is absent. Without these, automated result attribution and standings computation are impossible.

Replay bytes are **not persisted** after stat extraction is complete. The raw bytes serve as input to the parser only; once `player_stat_lines` and `team_stat_lines` are written, the bytes are discarded. The `replay_evidence.raw_bytes` column therefore does not need to store bytes long-term — it holds them transiently during parse, then is nulled out. This avoids any object storage dependency.

### Work Items

**4.1 — Async replay parse worker**

The replay evidence endpoint currently accepts `parse_output_json` inline. Change this:

- `POST /v1/replay-evidence` accepts raw replay bytes (multipart), buffers them in memory (or a temp column), sets `parse_status = queued`.
- A worker goroutine processes `queued` evidence by calling `kinetic-rl-parser` as a subprocess or HTTP service.
- On parse success: write stat lines (4.3), null out `raw_bytes`, set `parse_status = parsed`.
- On failure: set `parse_status = failed`, write error JSON, raise exception ticket.
- `GET /v1/replay-evidence/{id}` returns current parse status.

Keep the existing pre-parsed ingest path (`parse_output_json` inline) as an override for automated pipelines.

**4.2 — Identity resolution**

Implement the resolution step from ADR-012:

Given a parsed replay with player identifiers (platform + account ID), resolve to `player_id` by:
1. Look up `platform_account_links` for matching `(provider, provider_account_id)`.
2. If a verified link exists, map to `player_id`.
3. If no link, create a `replay_identity_mismatch` exception ticket for operator triage.

Add `resolved_player_id uuid REFERENCES players(id)` to the replay participant record (either as a JSONB field in `parse_output_json` or a new join table `replay_participants`).

**4.3 — Stat extraction**

After identity resolution, extract per-player and per-team stat lines from the parse output:

Schema additions:
```sql
CREATE TABLE rounds (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id   uuid NOT NULL REFERENCES result_submissions(id),
    round_number    int NOT NULL,
    duration_seconds int,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE player_stat_lines (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    round_id        uuid NOT NULL REFERENCES rounds(id),
    player_id       uuid REFERENCES players(id),  -- null if unresolved
    replay_identity text NOT NULL,                -- raw identity from replay
    goals           int NOT NULL DEFAULT 0,
    assists         int NOT NULL DEFAULT 0,
    saves           int NOT NULL DEFAULT 0,
    shots           int NOT NULL DEFAULT 0,
    score           int NOT NULL DEFAULT 0,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE team_stat_lines (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    round_id        uuid NOT NULL REFERENCES rounds(id),
    team_id         uuid REFERENCES teams(id),
    goals           int NOT NULL DEFAULT 0,
    shots           int NOT NULL DEFAULT 0,
    saves           int NOT NULL DEFAULT 0,
    created_at      timestamptz NOT NULL DEFAULT now()
);
```

Endpoints:
- `GET /v1/result-submissions/{id}/stats` — returns rounds + player_stat_lines + team_stat_lines for a finalized submission.

**4.4 — KineticRating calculation**

Add `opi`, `dpi`, `gpi` computed columns (or a derived endpoint) from stat lines, using the v1 formula:
- `opi = (goals + 0.75 * assists + 0.3 * shots) / rounds_played`
- `dpi = saves / rounds_played`
- `gpi = opi + dpi`

Expose via `GET /v1/player-ratings?include_kinetic_rating=true` or as fields on the rating response.

---

## Theme 5: Unified Rating and Skill Group System

### Context

The [v2 unified matchmaking proposal](https://minor-league-esports.github.io/knowledgeBase/departments/development/features-and-designs/kinetic-v2-unified-matchmaking-proposal/) defines the rating model v3 should implement. The core principle: **all players share one continuous Elo pool (0–3000+) regardless of which league they compete in**. Skill groups are configurable labeled brackets along this continuum, not isolated rating pools. ADR-010 in v3 already commits to this model; Theme 5 is its implementation.

### 5A: Skill Group Entities

Skill groups are currently implicit (referenced as strings in `player_ratings.context_key`) but have no first-class entity. They need to become configurable per-league records.

**Schema:**
```sql
CREATE TABLE skill_groups (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    league_id           uuid NOT NULL REFERENCES leagues(id),
    name                text NOT NULL,           -- e.g. "Foundation", "Academy", "Champion"
    display_order       int  NOT NULL,           -- ascending: lower number = lower tier
    rating_floor        numeric NOT NULL,        -- inclusive lower bound on unified rating scale
    rating_ceiling      numeric NOT NULL,        -- inclusive upper bound
    promotion_threshold numeric NOT NULL,        -- must exceed this to be eligible for promotion (> rating_ceiling - hysteresis_buffer)
    demotion_threshold  numeric NOT NULL,        -- must fall below this to be demoted (< rating_floor + hysteresis_buffer)
    is_active           bool NOT NULL DEFAULT true,
    created_at          timestamptz NOT NULL DEFAULT now(),
    UNIQUE (league_id, name)
);
```

The `promotion_threshold` and `demotion_threshold` encode the hysteresis buffers. Example for the reference ranges from the v2 proposal:

| Skill group | Floor | Ceiling | Promotion threshold | Demotion threshold |
|---|---|---|---|---|
| Foundation | 0 | 800 | 820 | — |
| Academy | 700 | 1200 | 1220 | 680 |
| Champion | 1100 | 1600 | 1620 | 1080 |
| Master | 1500 | 2000 | 2020 | 1480 |
| Premier | 1900 | 3000+ | — | 1880 |

The 140-point spread between adjacent thresholds (e.g., Academy promotes at 1220, demotes at 1080) prevents oscillation. These are seeded values, overridable per-league via the config API.

**Endpoints:**
- `GET /v1/leagues/{id}/skill-groups` — list skill groups for league
- `POST /v1/leagues/{id}/skill-groups` — create skill group (admin)
- `PATCH /v1/skill-groups/{id}` — update thresholds (admin)
- `DELETE /v1/skill-groups/{id}` — deactivate skill group (admin)

Replace `player_ratings.context_key` (text) with `skill_group_id uuid REFERENCES skill_groups(id)` in a migration. The `context_key` string was a placeholder for this entity.

### 5B: Automated Glicko-2 Rating Update on Finalization

After a `result_submission` transitions to `ratified`, trigger rating recalculation synchronously in the same transaction:

- Apply a Glicko-2 style update using the existing `player_ratings.rating` and `player_ratings.uncertainty` columns.
- Expected score formula: `E = 1 / (1 + 10^((opponent_rating - player_rating) / 400))`.
- Rating delta: `Δ = K * (actual_score - E)` where K is derived from `uncertainty` (higher uncertainty → larger K → faster convergence).
- Uncertainty decay: reduce `uncertainty` by a factor each match; floor at `MIN_UNCERTAINTY` (configurable, default 50).
- Update `player_ratings` for all participants; write `rating_adjustments` rows (automated: `actor_player_id = NULL`, reason = `scrim_result` | `league_result`).

No background worker needed — Glicko-2 math is O(n) on participants and completes in microseconds.

### 5C: Skill Group Boundary Evaluation with Hysteresis

After each rating update, evaluate whether the player has crossed a skill group boundary:

1. Load the player's current `skill_group_id` and the corresponding `skill_groups` record.
2. Check if `new_rating >= skill_group.promotion_threshold` → eligible for promotion.
3. Check if `new_rating <= skill_group.demotion_threshold` → eligible for demotion.
4. Find the target skill group (adjacent by `display_order`).
5. If a transition is warranted:
   - Update `player_ratings.skill_group_id` to the new group.
   - Insert into `skill_group_transitions`.
   - Write a `skill_group_change` player notification.

```sql
CREATE TABLE skill_group_transitions (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id           uuid NOT NULL REFERENCES players(id),
    from_skill_group_id uuid REFERENCES skill_groups(id),
    to_skill_group_id   uuid NOT NULL REFERENCES skill_groups(id),
    rating_at_transition numeric NOT NULL,
    direction           text NOT NULL CHECK (direction IN ('promotion', 'demotion')),
    effective_at        timestamptz NOT NULL DEFAULT now()
);
```

### 5D: Queue Expansion Using Unified Rating

Update the matchmaking algorithm (in `scrim-promotions/process`) to use the unified rating for expansion instead of a per-skill-group window:

| Queue wait time | Search radius | Cross-group behavior |
|---|---|---|
| 0–2 min | ±100 rating points | Same skill group preferred |
| 2–5 min | ±150 rating points | Cross-group allowed within 50 pts of boundary |
| 5–10 min | ±250 rating points | Cross-group unrestricted |
| 10+ min | ±400 rating points | Emergency — match best available |

Cross-group matches are semantically valid because all players share the same rating scale. A Foundation player at 795 vs. an Academy player at 805 is a 10-point spread — a fairer match than two same-group players 200 points apart.

Record the expansion stage and cross-group flag on `matchmaking_decisions` (both columns already exist from the v3 foundation work).

### 5E: KineticRating performance indices

After stat lines are extracted (Theme 4.3), compute per-player performance indices from the v1 formula:

- `opi = (goals + 0.75 * assists + 0.3 * shots) / rounds_played`
- `dpi = saves / rounds_played`
- `gpi = opi + dpi`

Store on `player_stat_lines` as computed columns or a derived summary table. Expose via `GET /v1/result-submissions/{id}/stats` and a player career aggregate via `GET /v1/players/{id}/stats`.

---

## Theme 6: Organization Configuration

### Context

v1 had `OrganizationConfigurationKeyCode` enum with runtime-tunable policy values. v3 currently hardcodes all policy values. At minimum, pop timeout, queue-ban durations, and ratification thresholds need to be configurable without code changes.

### Work Items

**6.1 — Organization config table + API**

Schema:
```sql
CREATE TABLE organization_configs (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    league_id   uuid NOT NULL REFERENCES leagues(id),
    key         text NOT NULL,
    value       text NOT NULL,
    updated_by  uuid REFERENCES players(id),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE (league_id, key)
);
```

Initial config keys:
- `scrim.pop_timeout_minutes` (default: 5)
- `scrim.queue_ban_initial_duration_minutes` (default: 60)
- `scrim.queue_ban_duration_modifier` (default: 2.0 — doubles on each subsequent ban)
- `scrim.queue_ban_modifier_falloff_days` (default: 30)
- `scrim.required_ratifications` (default: 2 — home + away)
- `queue.expansion_interval_seconds` (default: 30)

Endpoints:
- `GET /v1/leagues/{id}/config` — read config for league (admin/fm only)
- `PATCH /v1/leagues/{id}/config` — update one or more keys (admin only)

Config is loaded at startup and cached in memory with a 60-second TTL. Hot reload via cache invalidation on PATCH.

---

## Theme 7: API Surface Completeness

Fill in the missing REST endpoints identified in the gap analysis that are required by the web client delivery plan but don't yet exist.

### Work Items

**7.1 — Single-resource lookups**

- `GET /v1/fixtures/{id}` — single fixture with teams and scheduling metadata
- `GET /v1/result-submissions/{id}` — single submission with full state
- `GET /v1/scrims/{id}` — single scrim

**7.2 — Submission lifecycle completions**

- `POST /v1/result-submissions/{id}/reset` — re-open a rejected submission to `pending` state (admin/operator only); writes audit row
- `GET /v1/result-submissions?state=pending_ratification&team_id={id}` — filter by ratification-pending state for a team

**7.3 — Player activation workflow**

- `POST /v1/players/{id}/activate` — set `is_active = true`; requires `player.update` permission
- `POST /v1/players/{id}/deactivate` — set `is_active = false`; requires `player.update` permission

**7.4 — OAuth/OIDC completion**

The `/v1/auth/login` and `/v1/auth/callback` endpoints are stubbed. Complete them with a multi-provider OAuth2 implementation. Discord credentials are available immediately; Google, Microsoft/Xbox, PSN, Steam, and Epic credentials can be procured as needed.

Provider support matrix:

| Provider | Use case | Priority |
|---|---|---|
| Discord | Community identity; already used in v1/v2 | P0 — launch |
| Steam | PC platform account linking for replay identity | P1 — replay pipeline |
| Epic Games | PC platform account linking | P1 — replay pipeline |
| Microsoft / Xbox | Console platform account linking | P1 — replay pipeline |
| PSN | Console platform account linking | P1 — replay pipeline |
| Google | Alternative login for non-Discord users | P2 — post-launch |

Design:
- `/v1/auth/login?provider={name}` — redirect to provider's OAuth authorization URL.
- `/v1/auth/callback?provider={name}` — exchange code, upsert `platform_account_links` for the provider identity, issue session cookie.
- Session identifies the player via `platform_account_links.player_id`; if no player record exists yet, create an inactive player pending admin review.
- Steam, Epic, Xbox, and PSN callbacks additionally fulfill platform account linking (not just login) — the same callback flow serves both auth and `platform_account_links` population.
- `GET /v1/auth/providers` — list configured and enabled OAuth providers (for the web client login page).

**7.5 — Eligibility endpoint**

- `GET /v1/eligibility?player_id={id}` — returns eligibility status, current skill group, active queue bans, pending result submissions, rating, and projected skill group transition threshold.

**7.6 — OpenAPI specification**

Generate or hand-author an OpenAPI 3.1 spec for all `/v1/*` routes. This:
- Enables auto-generated TypeScript client for the web client (replaces hand-authored wrappers).
- Enforces a contract artifact that changes require updating (ADR candidate).
- Enables testing tools and third-party integrations without Discord dependency.

---

## Sequencing and Dependencies

The themes have the following dependency ordering:

```
Theme 1 (RBAC) ─────────────────────────────────────────────► all themes depend on this
Theme 5A (skill group entities) ────────────────────────────► Theme 5B/5C/5D (rating uses skill_group_id)
Theme 7.4 (OAuth) ──────────────────────────────────────────► Web client auth shell
Theme 2 (me/* endpoints, lobby credentials) ───────────────► Web Phase 2 (player scrim core)
Theme 3 (scrim state machine) ─────────────────────────────► Theme 2 (me/scrim)
Theme 6 (org config) ──────────────────────────────────────► Theme 3 (pop timeout config), Theme 5A (threshold seeds)
Theme 5B/5C/5D (Glicko-2 + transitions + expansion) ──────► Theme 5A prereq
Theme 4 (replay parsing + stat extraction) ────────────────► Theme 5E (KineticRating uses stat lines)
Theme 7.1–7.6 (API completeness) ──────────────────────────► Web Phase 1–5
```

### Recommended phasing alongside the web client delivery plan

| Web Phase | Backend prerequisites from this plan |
|---|---|
| Phase 0 (Foundation) | Theme 1.1 (resource taxonomy), Theme 7.4 (OAuth) |
| Phase 1 (Support + Ops) | Theme 7.1 (single-resource lookups), Theme 7.2 (submission lifecycle) |
| Phase 2 (Player Scrim Core) | Theme 2 (me/* endpoints), Theme 3 (scrim state machine + check-in), Theme 5A (skill groups visible to players) |
| Phase 3 (League Admin) | Theme 6 (org config + skill group thresholds), Theme 7.5 (eligibility with skill group context) |
| Phase 4 (Roster + Roles) | Theme 1.2–1.3 (scope enforcement + approval), Theme 5B/5C/5D (automated rating + transitions), Theme 7.3 (activation) |
| Phase 5 (Account + Eligibility + Overrides) | Theme 4 (replay pipeline + stat extraction), Theme 5E (KineticRating), Theme 1.4 (API tokens), Theme 7.6 (OpenAPI) |

---

## Explicitly Out of Scope

- Discord bot and notification service (removed by design)
- RabbitMQ / event bus
- Redis for hot ephemeral state
- Object storage for replay bytes (bytes discarded after parse — no persistence needed)
- InfluxDB / analytics telemetry pipeline
- MLEDB compatibility bridge
- GraphQL (REST-only by architectural decision)
- Image / report card generation service
- Salary calculation engine
- LFS (looking-for-scrim) mode — defer until post-launch signal
- ModePreference (PREFER_2S / PREFER_3S) — defer until multi-mode queue is a priority
- Google OAuth — P2, not blocking launch
- Chatwoot / customer support widget

---

## Open Questions

1. **Pop timeout worker architecture**: The MVP goroutine-per-scrim approach works at low scrim volume. At high scrim volume a periodic DB sweeper is safer (avoids goroutine leak on crash). Suggested inflection point: switch to a ticker-based sweeper when average concurrent scrims exceeds ~50. Should the sweeper run as an internal goroutine on the same binary or as a separate process?

2. **OpenAPI tooling**: Code-generate the spec from handler annotations, or hand-author and validate in CI? Given v3's plain `net/http`, auto-generation requires tooling integration (e.g., `swaggo/swag` with comment annotations). Hand-authored + `kin-openapi` validation in CI may be simpler and more maintainable. Decision needed before Phase 5.

3. **Skill group initial seeding**: The reference ranges from the v2 proposal (Foundation 0–800, Academy 700–1200, etc.) are reasonable defaults. Should these be seeded in a migration as league-specific records, or loaded from the organization config system? Recommendation: seed them as `skill_groups` rows in an initial data migration, overridable via the API.

4. **Rating initialization**: New players have no rating. Options: (a) start all new players at a fixed midpoint (e.g., 800) with high uncertainty so the system converges quickly, or (b) run calibration matches before assigning a skill group. Option (a) is simpler; option (b) was discussed in the v2 proposal's seasonal recalibration section. Decision affects the new-player onboarding flow.
