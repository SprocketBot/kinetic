# Why Sprocket Is Unsustainable

*For new contributors — the short version of why we're building a replacement.*

## What Sprocket Does

Sprocket is the platform that runs Rocket League esports for Minor League
Esports (MLE). It handles queueing, scrims, match scheduling, replay parsing,
Elo ratings, roster management, and Discord integration. It has been in
production for several seasons and is generally stable.

## The Problems

### 1. It's Expensive to Run

Sprocket requires about **10 GB of RAM** just to keep all its services running.
That forces us onto a $120/month VM (the smallest DigitalOcean tier with enough
RAM is 16 GB). On top of that, we pay ~$30/month for a managed PostgreSQL
database. That's **~$150/month** for a platform that manages a single esports
league.

| Resource | Monthly Cost |
|---|---|
| 16 GB VM | ~$120 |
| Managed PostgreSQL | ~$30 |
| **Total** | **~$150** |

### 2. It's a Leaky Balloon of Services

Sprocket was never designed from scratch. It was built incrementally over years,
with each new feature bolted on as a separate service. Right now it's:

- **7 backend services**, each with its own Docker image, dependencies, and
  runtime
- **2 frontend clients** (web + Discord bot)
- **Plus** a replay parser, a Redis cache, a RabbitMQ message broker, and an
  image generation service

Every one of these is a thing that can break, needs updates, and consumes
memory. The NestJS framework they're built on adds startup time and runtime
overhead that makes everything heavier than it needs to be.

### 3. The Data Model Has a Bad Assumption

Sprocket's database has a concept called "organizations" that was inherited from
an earlier design. It was meant to let one platform serve multiple leagues, but
in practice every MLE league shares one organization, and the org abstraction
adds complexity to every single database query and permission check. Untangling
it would require a full data model rewrite — which is effectively what Kinetic
is.

### 4. It's Tied to a Legacy Database

Sprocket still reads and writes to the old MLEDB schema for some operations.
This means there are two sources of truth for roster, match, and player data.
When something doesn't match between the two, it produces bugs that are
difficult to diagnose.

### 5. The Runtime Is a Problem

Sprocket runs on Node.js and TypeScript — a dynamically-typed language compiled
to JavaScript. This means:

- Runtime errors that a compiled language would catch at build time instead
  show up in production
- Startup time for each service is slow (seconds per container)
- Memory usage is high because of the JIT compiler and garbage collector
- TypeScript's type system catches some bugs, but not all — and the type
  definitions drift from reality over time

### 6. It Can't Be Handed Off

Because of all the above, Sprocket has become a system that only one or two
people fully understand. The service boundaries, the data model quirks, the
deployment dance — none of it is simple enough to hand to new contributors and
say "go run this."

## The Bottom Line

Sprocket works today, and it will keep working for a while. But every new
feature or bug fix costs more than it should, the infrastructure bill is
unnecessary, and the complexity means nobody new can maintain it. We could keep
patching it, but the right move is to replace it with something simpler.

That something is **Kinetic**.