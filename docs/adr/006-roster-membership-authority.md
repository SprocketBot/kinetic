# ADR-006: Make Roster Membership the Team Ownership Source

- Status: Accepted
- Date: 2026-02-09
- Owner: jacbaile
- Supersedes: ADR-005 follow-up on `player.team_id`

## Context

Week 5 introduced `roster_memberships` while `players.team_id` still existed.
That duplicated ownership semantics and allowed data drift (player row and membership rows disagreeing).

## Decision

Make `roster_memberships` authoritative for player-to-team assignment.

### Contract changes

- Remove `players.team_id` from schema and API model.
- Player creation no longer requires `teamId`.
- Active assignment is represented only by `roster_memberships`.
- Enforce one active roster membership per player with a partial unique index on `roster_memberships(player_id)` where `is_active=true`.

### API impact

- `POST /v1/players` payload removes `teamId`.
- Team assignment is performed through `POST /v1/roster-memberships`.

## Consequences

### Positive

- Eliminates split-brain ownership for team assignment.
- Clarifies write path for transfers and future roster lifecycle logic.
- Keeps assignment constraints in one place.

### Negative / Tradeoffs

- Existing assumptions that player creation implies team ownership must be updated.
- Historical assignment tracking remains deferred (only active-state enforcement is introduced here).

## Follow-ups

- Add explicit deactivation/transfer endpoints for roster memberships.
- Add filtered roster queries by team/player/active state.
