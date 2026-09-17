# ADR-023: Scheduled Competition Lifecycle and Standings

- Status: Proposed
- Date: 2026-09-16
- Owner: Nigel
- Related: ADR-008 (scheduled match model), ADR-009 (scheduled competition slice)

## Context

ADR-009 introduced the scheduling hierarchy (Season → ScheduleGroup → Fixture → Match) with create/list APIs only. The data model exists but is unusable for actual league operations.

This ADR fills in the gaps:
1. Update/delete endpoints
2. State transitions
3. Automatic schedule generation
4. Result submission and ratification
5. Standings calculation

## Decision

### 1. Update/Delete Endpoints

Extend all scheduling entities with standard CRUD:

```
PATCH /v1/seasons/{id}           Update season (name, isActive)
DELETE /v1/seasons/{id}           Soft delete (set isActive = false)

PATCH /v1/schedule-groups/{id}    Update schedule group (name, sequence)
DELETE /v1/schedule-groups/{id}  Soft delete

PATCH /v1/fixtures/{id}          Update fixture (homeClubId, awayClubId)
DELETE /v1/fixtures/{id}          Soft delete (cascades to matches)

PATCH /v1/matches/{id}           Update match (homeTeamId, awayTeamId, scheduledFor)
DELETE /v1/matches/{id}         Soft delete
```

**Note:** Soft delete via `is_active = false` maintains referential integrity and audit trail.

### 2. State Transitions

Implement the match lifecycle defined in ADR-008:

```
State: planned → ready → in_progress → submitted → under_review → finalized
                                         ↘→ voided (alternative terminal)
```

**Transition API:**

```
POST /v1/matches/{id}/transition
{
  "action": "schedule" | "start" | "submit" | "review" | "finalize" | "void",
  "payload": { ... }  // depends on action
}
```

**Transition matrix:**

| From | Action | To | Requirements |
|------|--------|-----|--------------|
| `planned` | `schedule` | `ready` | `scheduledFor`, `homeTimeRatifiedAt`, `awayTimeRatifiedAt` |
| `ready` | `start` | `in_progress` | — |
| `in_progress` | `submit` | `submitted` | `resultSubmissionId` |
| `submitted` | `review` | `under_review` | — |
| `under_review` | `finalize` | `finalized` | — |
| any | `void` | `voided` | `reason` |

**Time ratification:**

```
POST /v1/matches/{id}/ratify-time
{
  "ratifiedAt": "2026-09-20T19:00:00Z"
}
```

Both teams must ratify before transition to `ready`.

### 3. Automatic Schedule Generation

Add bulk generation endpoint:

```
POST /v1/schedule-groups/generate
{
  "seasonId": 123,
  "name": "Week 1",
  "sequence": 1,
  "teamIds": [1, 2, 3, 4, 5, 6],
  "format": "round_robin" | "double_round_robin"
}
```

**Algorithm:** Round-robin pairing ensures each team plays each other exactly once (or twice for double round robin).

**Output:** Creates ScheduleGroup + all Fixtures + all Matches in `planned` state.

**Validation:**
- Even number of teams required
- No duplicate fixtures in same season
- Teams must belong to clubs in the league

### 4. Result Submission and Ratification

When a match moves to `submitted`:

```
POST /v1/matches/{id}/submit-result
{
  "homeScore": 3,
  "awayScore": 1,
  "replayEvidenceId": 456,
  "submittedByTeamId": 789
}
```

**Ratification (after submission):**

```
POST /v1/matches/{id}/ratify-result
{
  "ratified": true,
  "teamId": 789
}
```

Both teams must ratify for match to move to `finalized`. If rejected, moves to `under_review`.

### 5. Standings

Calculate standings from finalized matches:

```
GET /v1/standings?seasonId=123
```

**Response:**

```json
{
  "seasonId": 123,
  "standings": [
    {
      "rank": 1,
      "teamId": 1,
      "teamName": "Aviators",
      "wins": 8,
      "losses": 2,
      "draws": 0,
      "points": 24,
      "gamesPlayed": 10,
      "goalsFor": 25,
      "goalsAgainst": 10,
      "goalDifferential": 15
    },
    ...
  ]
}
```

**Points system:**
- Win: 3 points
- Draw: 1 point
- Loss: 0 points

**Sorting:** By points desc, then goal differential, then goals for.

## API Summary

| Endpoint | Description |
|----------|-------------|
| `PATCH/DELETE /v1/seasons/{id}` | Update/delete season |
| `PATCH/DELETE /v1/schedule-groups/{id}` | Update/delete schedule group |
| `PATCH/DELETE /v1/fixtures/{id}` | Update/delete fixture |
| `PATCH/DELETE /v1/matches/{id}` | Update/delete match |
| `POST /v1/matches/{id}/transition` | State transition |
| `POST /v1/matches/{id}/ratify-time` | Ratify scheduled time |
| `POST /v1/matches/{id}/submit-result` | Submit match result |
| `POST /v1/matches/{id}/ratify-result` | Ratify match result |
| `POST /v1/schedule-groups/generate` | Generate schedule |
| `GET /v1/standings` | Get standings by season |

## Consequences

### Positive

- Complete CRUD operations for scheduling hierarchy
- Enables actual league运营 (schedule generation → play → results → standings)
- State machine ensures valid transitions only
- Standings provide immediate visibility into league progress

### Negative / Tradeoffs

- Complex state transition matrix requires careful validation
- Schedule generation assumes round-robin; custom formats need future work
- Standings calculation is read-heavy; consider caching for large seasons

## Non-Goals

- MMR/rating integration from match results (separate ADR)
- Playoff/bracket generation (single elimination, double elimination, etc.)
- Tiebreaker rules beyond point differential
- Automatic scheduling conflicts detection across seasons
- Real-time match tracking (score updates during play)

## Follow-ups

- Add playoffs/bracket generation ADR
- Define rating update pipeline for matches
- Add email/notification triggers for schedule changes
- Consider caching standings for performance
