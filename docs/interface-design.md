# Designing our Interfaces for v3

Now that we have a core backend fully functional, it's time to work out the
other pieces of what this new version of the platform experience will look like.
We have four different points of view to consider:

- User
- League administrator
- League support operator
- Platform operator

Let's begin with the easiest to define, the platform operator:

## Platform Operator

The platform operator will need an interface through which to ensure that the
platform is running as expected, and to handle issues as they come up. They'll
need a central place to

- view system logs, metrics, and alerts (Grafana works well for this)
- field tickets from administrators for issues, e.g.
  - we're not getting notifications
  - the admin panel for matches seems broken
  - monitor code changes/builds/tests/rollouts (github/GHCR is the likely
    candidate).

## League Administrator

The league administrator will need an interface through which they can define
the overall operations of the league:

- Define seasons, schedule groups, fixtures, and matches (CRUD)
- CRUD for games, including scrim and match formats, parser tools (This is a
  secondary feature, experimental/we'll work on it later)
- Override invalid or revised match results (we refer to this colloquially as
  "NCP")
- CRUD on role-assignments for top level roles, like Franchise Manager or Club
  Manager/GM

## League Support Operator

The league support operator will need an interface to:

- See all players currently active in scrims
- See all active scrims and their data, and perform basic CRUD (including
  cancellation)
- See all submissions in process and their data
- Submit replay data on behalf of players
- Ban/unban players from queueing
- Field tickets from users for support requests
- Add players to the league's roster after account creation

## User

Finally, the user/player interface will be the most complex:

### Account management

- Authentication/login. We should only support OAuth/SSO, we don't want to
  manage passwords
- Signup.
  - When first accessing the app, users create an account.
  - This account will be _inactive_ at first. Unable to scrim or participate in
    matches, or really do much at all.
  - League support, after review the application (handled separately) will add
    the player to the league roster/activate the account.
- Platform account association.
  - Again, via OAuth/SSO/OIDC.
  - Once signed in to app, link your Steam/Xbox(MSN)/PSN/Epic platform accounts.
  - Manage accounts, unlink, etc.

### Scrim management

- All actions that are present in
  ./matchmaking/unified-matchmaking-core-concepts.md
  - View scrim activity
  - Queue for scrims in available game modes (multi-queueing supported)
  - Get generated data like lobby info and team assignments per game
  - Upload replays
  - Review parsed results
  - Ratify results

### Roster management

Some players will be assigned roles either by league administration (FM/GM) or
by franchise/club management. These roles include:

- Franchise Manager (manages Franchises, hires/fires GMs)
- General Manager (manages Clubs, hires/fires AGMs and Captains)
- Assistant General Manager (same as GM)
- Captain (manages Teams, hires/fires players)
- These roles are proper subsets: AGM has all Captain perms for the whole club,
  GM has all AGM perms, FM has all GM perms.
- These players should have CRUD access to their associated rosters, e.g. team
  captains should be able to offer roster slots on their teams to free agents,
  and to release/waive players, GMs should be able to offer AGM or Team Captain
  slots to players, etc.
- All players should be able to _view_ rosters and management roles of all
  clubs/teams.

### Match management

- All players should be able to _view_ their club's match schedule, including:
  - Results of matches that have already been submitted
  - Scheduled times for upcoming matches (once ratified)
  - Submission status of played, but not submitted, matches
- Captains should be able to propose, accept, and ratify scheduled times for
  their matches.
- Captains should be able to upload, ratify, and submit replay data for their
  matches.

### Elo and Eligibility

- All players should be able to _view_ their elo rating/bracket.
- No player should be able to change their rating
- League administration should be able to change any player's rating other than
  their own
- All players should be able to _view_ their current eligibility points, as well
  as the decay schedule for them and an "eligible until" date.

## A word on Evidence (captial E, the platform Evidence)

Our league already provides most of the _viewing_ capability for all of the
user/player data I have just outlined via a static site we publish once per hour
via a framework called Evidence (https://evidence.dev/). We do not want to
duplicate any of this capability. Ideally, we would be able to embed various
views from the static site into the web client for the Sprocket app, and simply
provide whatever additional interface is necessary to support the rest of the
CRUD capabilities, as applicable.

## A word on styling

The sprocket platform already has a pretty recognizable brand in terms of
styling and art. You can get a feel for it by checking out the web client code
in github.com/SprocketBot/sprocket. The v1 (current) UI is on the `main` branch,
and the v2 UI on `dev`.
