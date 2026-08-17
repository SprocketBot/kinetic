# Adding a New Game to Kinetic

*For new contributors — how Kinetic supports multiple games and what it takes
to add a new one like Trackmania.*

## How Kinetic Handles Games

Kinetic treats games as first-class entities in the data model. A "game" is
whatever competition you want to run — Rocket League, Trackmania, or anything
else. The game concept is the root of the hierarchy:

```
Game → League → Franchise → Club → Team → Player
```

A single Kinetic deployment can host multiple games, but each game has its own
leagues, teams, and players. A user (a real person) can be a player in multiple
games — the same person can be a Rocket League player and a Trackmania player,
with separate identities, ratings, and stats in each.

### What's Built Already

The game infrastructure is already in place:

- **`games` table** — games have a name and slug. Rocket League is seeded
  automatically. New games are created via the `CreateGame` API.
- **`user_players` table** — links a user to a player identity within a specific
  game. A user's Trackmania player identity is separate from their Rocket League
  identity.
- **`leagues.game_id`** — each league is scoped to a game. A Trackmania league
  is separate from a Rocket League league.
- **`platform_account_links`** — platform accounts (Steam, Epic, Xbox, PSN) can
  be linked to a player within a game for identity resolution during replay
  parsing.

### What Must Be Built Per Game

While the game entity and identity infrastructure is generic, several things
are game-specific:

| Component | Rocket League | Trackmania (needs work) |
|---|---|---|
| Identity store | Built — `CreateGame("Trackmania")` works | Just needs a DB seed or API call |
| League scoping | Built — leagues scoped to RL game | Create TM leagues scoped to TM game |
| Replay parser | Needs async pipeline (gap closure Theme 4) | Existing parser, needs Kinetic integration |
| Stat model | goals/assists/saves/shots | points/respawns/best-times/CP-times |
| Game-specific validation | Built for RL | Needs TM-specific validation rules |
| Queue + scrim lifecycle | Built (generic) | Should work generically with game context |
| Rating system | Built (Glicko-2, unified pool) | Same system works — different pool per game |
| Eligibility rules | Built | May need TM-specific rules |
| Operator inbox | Built (generic) | Same system works — filter by game |

## Trackmania-Specific Integration Guide

Trackmania has an existing Sprocket-backed setup:
- **`tm-q-bot`** — a Discord bot for queue management (TypeScript, separate
  codebase)
- **Parser** — replay submission and verification (currently Apps Script)
- **Sprocket database** — uses Sprocket tables for fixtures, rosters, eligibility
- **Datasets** — public SQL queries for match data, Elo, player stats

The `tm-q-bot` has partial Sprocket integration: it syncs queue players from
Sprocket profiles, creates local scrim records, and writes results to Sprocket
tables.

### The Decision to Make

Before any integration work starts, someone needs to decide:

> **Is Trackmania going Kinetic-native, or staying on Sprocket until the
> Rocket League cutover is complete?**

**Option A: Trackmania stays on Sprocket first.**
- Lower risk — Kinetic and TM can evolve independently
- The existing `tm-q-bot` + Sprocket integration keeps working
- TM moves to Kinetic only after RL is fully cutover and multi-game support is
  proven
- Downside: TM doesn't benefit from Kinetic improvements until later

**Option B: Trackmania goes Kinetic-native now.**
- TM gets the same benefits: simpler deployment, lower cost, better tooling
- Forces multi-game support to be built sooner
- Higher complexity — need to build TM-specific replay parsing, stat models,
  and validation alongside the RL work
- Downside: more parallel work

### What the Kinetic-Native Integration Would Look Like

If the decision is "go Kinetic-native," here's the work:

**Step 1: Create the Trackmania game entity**

```
POST /v1/games
{
  "name": "Trackmania",
  "slug": "trackmania"
}
```

This creates the Trackmania game in the database, which immediately enables:
- Creating Trackmania leagues, franchises, clubs, teams
- Creating Trackmania player identities for users
- Scoping everything to the Trackmania game

**Step 2: Set up the Trackmania league hierarchy**

Create leagues, schedule groups, fixtures, and teams — the same as Rocket
League, just scoped to game ID 2 (or whatever Trackmania gets).

**Step 3: Connect the replay parser**

Trackmania has an existing parser. It needs to be adapted to Kinetic's replay
pipeline:

- Instead of writing to Sprocket tables, have it POST to Kinetic's replay
  ingestion endpoint
- The parser output format needs to be mapped from Trackmania stats
  (points, respawns, best times, CP times) into Kinetic's stat line model
- Kinetic's identity resolution (platform account → player ID) needs to work
  for Trackmania players

If the parser continues as an external service, it calls the Kinetic API
directly. If you want to run it server-side, it can be a subprocess that
Kinetic invokes (similar to how the RL parser works).

**Step 4: TM-specific stat model**

The current `player_stat_lines` table has Rocket League-specific columns
(goals, assists, saves, shots). Trackmania needs different stats. Options:

1. Add TM-specific columns to the existing table (nullable, game-filtered)
2. Use a JSONB column for game-specific stats with game-aware queries
3. Create a game-specific stat line table

Option 2 (JSONB) is the simplest approach — it lets each game define its own
stat schema without schema changes for every new game.

**Step 5: TM-specific replay validation**

Kinetic's replay validation is currently Rocket League-specific. Trackmania
needs its own validation rules:
- Correct number of players per team
- Map-level results consistent across replays
- Winner determination from cumulative match score

These would be added as a game-specific validation plugin that dispatches based
on the game ID.

**Step 6: Queue + scrim operations**

The queue and scrim lifecycle in Kinetic is game-agnostic — it operates on
teams and players without caring what game they're playing. A Trackmania queue
is just a queue scoped to a Trackmania league. This should work with minimal
changes.

**Step 7: Ratings and eligibility**

Kinetic's Glicko-2 rating system is a single continuous pool per game.
Trackmania players would have their own ratings, separate from Rocket League.
The same rating update, skill group promotion/demotion, and eligibility logic
works — just scoped by game.

## What the TM-to-Kinetic Cutover Looks Like

Once Kinetic supports Trackmania, the migration involves:

1. **Set up a Kinetic Trackmania instance** — new database, deploy Kinetic
   binary, create Trackmania game + league hierarchy
2. **Export TM data from Sprocket** — player identities, ratings, roster
   assignments, match history (mostly in `trackmania.*` tables, some in
   Sprocket)
3. **Import into Kinetic** — via the API or direct DB migration
4. **Switch the TM bot** — instead of talking to Sprocket APIs, the `tm-q-bot`
   talks to Kinetic APIs (or gets replaced by the Kinetic web client)
5. **Update datasets** — point the public SQL queries at Kinetic's database
   instead of the Sprocket/Trackmania database

A detailed Trackmania implementation plan exists in the slipbox
(`archives/trackmania/trackmania-platform-implementation-plan-2026-05-10.md`)
with 7 phases covering identity fixes, fixture management, result submission,
Elo validation, roster ops, and a test run. That plan targets Sprocket, but the
same phases apply to a Kinetic target — just replace "write to Sprocket tables"
with "call Kinetic API."

## Key Questions to Answer Before Starting

These are listed as Phase 0 decisions in the cutover plan:

1. **Multi-game** — Is Trackmania integration happening before or after the RL
   cutover is complete?
2. **TM bot** — Does Trackmania keep its Discord bot for queue/scrim operations,
   or does it use the Kinetic web client?
3. **Parser** — Does the existing parser stay external, or does it move to
   Kinetic's worker model?
4. **Elo** — Does Trackmania keep its local Elo calculation, or does it use
   Kinetic's Glicko-2 system?
5. **Datasets** — Does Trackmania share the same dataset namespace as RL, or
   does it have its own?

## Reference Documents

- Kinetic ADR-012: Replay parsing and platform account association model
  (`docs/adr/012-*`)
- Sprocket cutover plan (`plans/sprocket-functionality-cutover-plan.md`)
- Kinetic identity store integration test (`internal/platform/db/
  hierarchy_identity_store_integration_test.go` — shows CreateGame + CreateUserPlayer)
- Trackmania dataset queries (in the `datasets` repo: `queries/public/trackmania/`)

## Demo: Creating a Trackmania Player in Code

The integration test at `internal/platform/db/
hierarchy_identity_store_integration_test.go` demonstrates the multi-game
identity model:

```go
// Create a Trackmania game
trackmania, err := store.CreateGame(ctx, hierarchy.CreateGameInput{
    Name: "Trackmania",
    Slug: "trackmania",
})

// List all games (includes seeded Rocket League)
games, err := store.ListGames(ctx)

// Create a user
user, err := store.UpsertUser(ctx, hierarchy.UpsertUserInput{
    Subject:     "discord:12345",
    DisplayName: "Jake",
})

// Create a Rocket League player for that user
rlPlayer, err := store.CreateUserPlayer(ctx, hierarchy.CreateUserPlayerInput{
    UserID: user.ID,
    GameID: games[0].ID,  // Rocket League
    DisplayName: "Rocket League Jake",
    Slug: "rocket-league-jake",
})

// Create a Trackmania player for the same user
tmPlayer, err := store.CreateUserPlayer(ctx, hierarchy.CreateUserPlayerInput{
    UserID: user.ID,
    GameID: trackmania.ID,  // Trackmania
    DisplayName: "Trackmania Jake",
    Slug: "trackmania-jake",
})
```