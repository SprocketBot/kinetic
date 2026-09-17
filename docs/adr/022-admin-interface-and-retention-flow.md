# ADR-022: Admin Interface and Player Retention Flow

- Status: Proposed
- Date: 2026-09-16
- Owner: Nigel
- Related: ADR-021 (role assignment schema)

## Context

ADR-021 defined the backend schema for role assignments and player retention. This ADR determines:
1. The interface for managing role assignments
2. The player retention submission flow
3. Whether Kinetic needs any Discord integration

## Decision

### 1. No Discord Bot

Kinetic will not have a dedicated Discord bot. The web application serves as the single interface for all administrative and player-facing operations.

**Rationale:**
- Simpler infrastructure (no Node.js service, no bot hosting)
- Single authentication/authorization model
- Easier to maintain and secure
- Web UI can be accessed by Discord-native users without requiring bot integration

### 2. Role Management via Web Admin

Role assignments (FM, PR, GM, AGM, Captain) are managed through the Kinetic web admin interface.

- FMs access franchise-level role management
- GMs access club-level role management  
- Captains access team-level role management
- RBAC enforces that users can only manage roles within their scope

**Rationale:** Even though these users are Discord-native, a web admin interface provides proper authentication, audit logging, and RBAC enforcement that a Discord bot couldn't easily match.

### 3. Player Retention via Web Form

Player retention/intent forms are exposed as authenticated web pages within Kinetic, replacing the current Google Forms workflow.

- Players must be logged in to submit (links to their Player record automatically)
- Form captures: season, intent
- Submissions are recorded in the `player_retentions` table

**Rationale:**
- Single source of truth (no Google Sheets sync)
- Authenticated submission prevents duplicate/impersonation issues
- Player identity is linked via session, no manual input needed

## Non-Goals

- Discord slash commands for role management
- Discord DM-based retention forms
- Real-time Discord role sync (external systems can poll Kinetic APIs)

## External Integrations

External systems (e.g., legacy bots, third-party tools) can consume Kinetic data via the REST API:

- `GET /v1/role-assignments` — Poll for role data
- `GET /v1/players` — Sync player roster
- `GET /v1/player-retentions` — Export retention data

This allows future migration of MLEBB's Discord role sync to an external process that reads from Kinetic APIs rather than the database directly.

## Consequences

### Positive

- Single platform for all operations
- No separate bot infrastructure to maintain
- Proper authZ and audit trail via web app
- Clean API surface for external consumers

### Negative / Tradeoffs

- Discord-native users must use web UI (may have usability friction initially)
- No real-time Discord role updates (external sync required)
- Retention form replaces a workflow players may be accustomed to

## Follow-ups

- Design admin role management pages with scope-based filtering (already exists in admin-home-page.tsx)
- Design player retention form page (authenticated)
- Define external sync process for Discord role updates (optional future work)
