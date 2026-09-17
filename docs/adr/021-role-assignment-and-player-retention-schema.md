# ADR-021: Role Assignment and Player Retention Schema Extensions

- Status: Proposed
- Date: 2026-09-16
- Owner: Nigel
- Related: ADR-003 (league hierarchy), ADR-004 (team-player slice), ADR-006 (roster membership)

## Context

The MLEBB bot (github.com/minor-league-esports/mlebb) provides essential Discord role management and player intent tracking for MLE operations. As part of migrating MLEBB functionality to Kinetic, we need schema extensions to support:

1. **Leadership role assignments** — FM, PR at franchise level; GM, AGM at club level; Captain at team level
2. **Franchise callsigns** — Short identifiers for teams (e.g., "TSM")
3. **Player retention tracking** — Season intent responses (Retain/FA/Former Player)

These extensions must integrate cleanly with existing Kinetic primitives (League → Franchise → Club → Team → Player, RosterMembership, PlatformAccountLink).

**Note:** The frontend interface for managing these entities (web admin vs Discord commands) is covered in a separate ADR.

## Decision

Implement three schema extensions following the existing slice patterns.

### 1. Role Assignments

Create a `role_assignments` table following the `roster_memberships` pattern:

```sql
CREATE TABLE role_assignments (
    id BIGSERIAL PRIMARY KEY,
    player_id BIGINT REFERENCES players(id) NOT NULL,
    
    -- Exactly one entity reference is set based on role level
    franchise_id BIGINT REFERENCES franchises(id),
    club_id BIGINT REFERENCES clubs(id),
    team_id BIGINT REFERENCES teams(id),
    
    role VARCHAR(20) NOT NULL CHECK (role IN ('FM', 'PR', 'GM', 'AGM', 'CAPTAIN')),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT now(),
    
    -- Enforce exactly one entity reference
    CHECK (
        (franchise_id IS NOT NULL)::int +
        (club_id IS NOT NULL)::int +
        (team_id IS NOT NULL)::int = 1
    ),
    
    -- Enforce one active role per player per scope
    UNIQUE (player_id, role, franchise_id) WHERE is_active = true,
    UNIQUE (player_id, role, club_id) WHERE is_active = true,
    UNIQUE (player_id, role, team_id) WHERE is_active = true
);

CREATE INDEX idx_role_assignments_player ON role_assignments(player_id);
CREATE INDEX idx_role_assignments_franchise ON role_assignments(franchise_id);
CREATE INDEX idx_role_assignments_club ON role_assignments(club_id);
CREATE INDEX idx_role_assignments_team ON role_assignments(team_id);
```

**Role-to-entity mapping:**

| Role | Entity Level | Permission Scope |
|------|--------------|------------------|
| `FM` (Franchise Manager) | Franchise | Full franchise control |
| `PR` (PR Support) | Franchise | Franchise communications |
| `GM` (General Manager) | Club | Full club/team control |
| `AGM` (Assistant GM) | Club | Club operations |
| `CAPTAIN` | Team | Team roster management |

**Rationale:** Reuses the established pattern from `roster_memberships` — explicit assignment table with one-active-per-player constraint to prevent split-brain ownership.

### 2. Franchise Callsigns

Add `callsign` field to `franchises` table:

```sql
ALTER TABLE franchises ADD COLUMN callsign VARCHAR(10);
ALTER TABLE franchises ADD CONSTRAINT unique_callsign UNIQUE (callsign);
```

**Validation:** Uppercase alphanumeric, max 10 characters.

**Rationale:** Callsigns are franchise-level identifiers in MLE convention. Adding at franchise level maintains the hierarchy.

### 3. Player Retention

Create `player_retentions` table:

```sql
CREATE TABLE player_retentions (
    id BIGSERIAL PRIMARY KEY,
    player_id BIGINT REFERENCES players(id) NOT NULL,
    season VARCHAR(50) NOT NULL,
    intent VARCHAR(20) NOT NULL CHECK (intent IN ('RETAINABLE', 'FREE_AGENT', 'FORMER_PLAYER')),
    submitted_at TIMESTAMPTZ DEFAULT now(),
    created_at TIMESTAMPTZ DEFAULT now(),
    
    UNIQUE (player_id, season)
);

CREATE INDEX idx_player_retentions_player ON player_retentions(player_id);
CREATE INDEX idx_player_retentions_season ON player_retentions(season);
```

**Intents:**

- `RETAINABLE` — Player wants to be retained by franchise, may become staff if not retained
- `FREE_AGENT` — Player enters free agency
- `FORMER_PLAYER` — Player opts out of the league

**Rationale:** Simple key-value store for season intent. Replaces Google Sheets workflow with durable database tracking. One response per player per season enforced at DB level.

## API Contract

### Role Assignments

```
POST /v1/role-assignments        Create role assignment
GET  /v1/role-assignments        List role assignments
GET  /v1/role-assignments?playerId={id}   Filter by player
GET  /v1/role-assignments?franchiseId={id} Filter by franchise
GET  /v1/role-assignments?clubId={id}     Filter by club
GET  /v1/role-assignments?teamId={id}     Filter by team
PATCH /v1/role-assignments/{id}  Update (e.g., deactivate)
DELETE /v1/role-assignments/{id} Remove assignment
```

Request/response:
```json
{
  "playerId": 123,
  "franchiseId": 45,
  "role": "FM",
  "isActive": true
}
```

### Franchise Callsign

```
PATCH /v1/franchises/{id}
{
  "callsign": "TSM"
}
```

### Player Retention

```
POST /v1/player-retentions      Submit retention response
GET  /v1/player-retentions     List retentions
GET  /v1/player-retentions?season=Season%2016  Filter by season
GET  /v1/player-retentions?playerId={id}      Filter by player
```

Request:
```json
{
  "playerId": 123,
  "season": "Season 16",
  "intent": "RETAINABLE"
}
```

## Consequences

### Positive

- Maintains hierarchy consistency: FM/PR → Franchise, GM/AGM → Club, Captain → Team
- Follows established slice patterns (create/list APIs, DB-level constraints)
- Enables full MLEBB functionality migration to Kinetic
- Retention tracking replaces external Google Sheets dependency

### Negative / Tradeoffs

- Role assignment constraints are complex (three-part unique index with WHERE clause); requires careful migration
- No historical tracking for role changes (role reassignment overwrites, not appends) — deferred to future slice
- Retention `season` is free-form string — consider enum or dedicated season table later

## Non-Goals (Current Scope)

- Role authorization enforcement at API level
- Transfer workflows between entities
- Historical role/retention audit trail
- Bulk assignment endpoints

## Follow-ups

- Add role-based authorization checks (ADR-022)
- Add filtered queries (activeOnly, by role type)
- Consider migration script from MLEBB data
- Design player retention form page (authenticated)
- Note: Admin UI for role management already exists in web client
