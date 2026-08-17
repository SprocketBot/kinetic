# Why Kinetic

*For new contributors — what Kinetic is, what it solves, and where it stands.*

## What Is Kinetic?

Kinetic is a complete rewrite of Sprocket in Go, designed from scratch to be
simpler, cheaper, and easier to maintain. It's a single Go binary (backend) plus
a React/TypeScript web app (frontend) — everything else is PostgreSQL.

## How It Solves Sprocket's Problems

### 1. One Binary Instead of Seven Containers

Sprocket needed 7 backend services + 2 clients + Redis + RabbitMQ. Kinetic
is a single Go binary. Everything is compiled into one executable.

| | Sprocket | Kinetic |
|---|---|---|
| Runtime size | ~10 GB RAM across all services | ~50 MB for the server binary |
| Number of services | 7 microservices + 2 clients + Redis + RabbitMQ | 1 server + 1 web client |
| Language | TypeScript (Node.js) | Go (statically compiled) |
| Database | PostgreSQL + Redis | PostgreSQL only |
| Message queue | RabbitMQ | Not needed (Postgres-backed) |
| Monthly infrastructure cost | ~$150 | Can run on a $12–$24 VM |

### 2. No More "Organization" Complexity

Sprocket's data model had an "organization" concept that added complexity to
every query and permission check. In Kinetic, each league gets its own database
instance. This means the organization abstraction is gone — the database
isolation does that job for free. The data model is clean:

```
Game → League → Franchise → Club → Team → Player
```

### 3. Statically Compiled = Fewer Surprises

Go is a compiled language. If the code compiles, common classes of bugs (wrong
types, missing return values, nil pointer dereferences) are caught before
deployment. The binary is self-contained — no separate Node.js runtime, no
package manager, no version compatibility headaches.

### 4. PostgreSQL-Only Architecture

Kinetic uses PostgreSQL for everything:
- Domain data (leagues, teams, players, matches)
- Queue state (no Redis needed)
- Authentication tokens and sessions
- Replay evidence metadata

This means no Redis to manage, no RabbitMQ to keep running, no cache
invalidation bugs. If the database is up, the platform works.

### 5. Designed for Handoff

Kinetic was built with the expectation that new people would need to take over.
That's why it has:

- **Architecture Decision Records** (20 of them) explaining why every major
  design choice was made
- **Onboarding runbooks** (14+ weeks of progressive guides)
- **CI quality gates** that enforce tests, formatting, and linting
- **Release evidence automation** that produces proof bundles for every
  promotion
- **Kubernetes-native deployment** with local minikube for development

## Current State (August 2026)

The core platform is built. Here's what exists:

### Built and Working

- Game, league, franchise, club, team, player hierarchy — create/list/update
- User identity per game (one person can be a player in Rocket League and
  Trackmania)
- Role-based access control with scope-aware enforcement (captains manage their
  own team, FMs manage their franchise)
- Queue enrollment — join/leave/list with FIFO ordering
- Scrim lifecycle — create, check-in, pop timeout, cancel, complete
- Deterministic queue promotion — rating-first with age-based tie-break
- Result submission and ratification — two-party ratify/reject
- Replay evidence ingestion with content-hash dedupe
- Operator exception inbox — ticket triage and resolution for scheduling
  conflicts, no-shows, replay failures
- Organization configuration — tunable policies per league
- Player notifications — in-app inbox for scrim pops, queue bans, skill changes
- Platform account links — Steam, Epic, Xbox, PSN identity binding
- API tokens — machine-client access with scoped permissions
- Skill groups — configurable rating brackets with promotion/demotion
- Rating adjustments — Glicko-2 updates with audit trail
- Player activation workflow — approve/reject new players
- Scheduled competition — seasons, schedule groups, fixtures, matches

### Ready for Web Client Work

All of the above has REST API endpoints. The web client (React/TypeScript) has
a delivery plan with 5 phases:

| Phase | What It Covers |
|---|---|
| Phase 0 (Foundation) | Auth shell, role-based routing, design system |
| Phase 1 (Support + Ops) | Operator inbox UI, player lookup, match administration |
| Phase 2 (Player Scrim Core) | Queue, scrims, lobby credentials, check-in |
| Phase 3 (League Admin) | Season/fixture management, schedule groups, skill groups |
| Phase 4 (Roster + Roles) | Role management, roster transfers, activation approvals |
| Phase 5 (Account + Stats) | Account linking, player stats, rating history, notification inbox |

### Still Needed Before Cutover

From the cutover plan (P0 priority):

1. **Authorization scope enforcement** — The RBAC policy system exists but
   needs the full resource taxonomy (30+ resource types) wired to every endpoint
2. **Real replay pipeline** — Currently accepts pre-parsed JSON. Needs async
   worker that invokes the Rust/Python parser, resolves identities, and
   extracts stat lines
3. **OAuth providers** — Discord login is wired; Steam, Epic, Xbox, PSN callbacks
   need implementation
4. **OpenAPI spec** — The current spec needs to be regenerated from the expanded
   API surface

## Infrastructure Requirements

Kinetic is designed to be simple to deploy:

```
Minimum viable setup:
- 1 VM (2 GB RAM, $12–$24/month)
- 1 PostgreSQL database (can be on the same VM or managed)
- 1 Kinetic server binary
- 1 web client (static files served by Nginx or the Kinetic binary)
- Optionally: Kubernetes (overkill for a single league, but supported)
```

The 14-week onboarding guides walk through every aspect of setup, development,
and deployment. Start with `docs/onboarding/week1.md`.