# Unified Replay Parsing Core Concepts

Date: 2026-02-14  
Status: Draft  
Source Parser: <https://github.com/KineticBot/kinetic-rl-parser>

## Purpose

Define the core concepts behind Kinetic's replay ingestion and identity-resolution model so product, engineering, and operations share the same contract.

## Problem Statement

League operations depend on trustworthy game outcomes and player stats, but raw replay data and user identity are disconnected by default:

- players can use multiple in-game identities across platforms (`Epic`, `Steam`, `Xbox`, `PSN`)
- auth/login identity and in-game identity are different concepts
- replay submissions can be duplicated, incomplete, or misattributed
- downstream standings/statistics become unreliable without deterministic attribution

## Core Philosophy

Replay files are the primary evidence of what happened in-game.

Identity is explicit and layered: user identity, player identity, and platform account identity must be modeled separately and then linked with verifiable association rules.

## Concept Model

### 1. Identity Layers

- `User`: authentication principal (email, Google, Discord, etc.)
- `Player`: competitive entity used by league/team features
- `Platform Account`: game-network identity from replay output (provider + provider account ID)

A single player can own multiple platform accounts, but each active platform account maps to one player at a time.

### 2. Account Association ("Account Reporting")

- Primary account-link path is authenticated OAuth while the user is signed in
- Link records store provider, immutable provider account ID, display metadata, association method, and timestamps
- Users can link multiple platform accounts to one player profile
- Historical replay attribution is evidence-based and should remain auditable when links later change

### 3. Replay Intake and Canonical Parse

- Replay files are uploaded from scrim/match workflows
- Raw bytes are fingerprinted and stored as immutable evidence
- Parsing is delegated to `kinetic-rl-parser`
- Parsed output is normalized into a canonical internal schema (JSON and/or protobuf backed)
- Parser provenance (name/version/config) is persisted for reproducibility

### 4. Participant Resolution

- Parsed participant platform identifiers are resolved against linked platform accounts
- Successful resolution maps replay participants to internal player records
- Name-only matching is non-authoritative and cannot be the primary identity key
- Unresolved or conflicting identities route to review instead of silent attribution

### 5. Competition Context Association

- Each replay is associated to a specific context (`scrim` or scheduled `match`)
- The system tracks submission provenance (who submitted, when, for what context)
- One game/round result should have one authoritative replay evidence record
- Duplicate replay submissions should converge to a single evidence identity

### 6. Result and Statistic Derivation

- Match outcomes and player statistics are derived from canonical replay parse outputs
- Derived records remain traceable to replay evidence + parser version
- Reprocessing is supported when parser versions evolve or corrections are required

### 7. Ingestion Lifecycle and Exception Handling

Replay processing follows explicit lifecycle stages (for example: `received -> parsed -> resolved -> finalized`) with review side paths (`needs_review`, `rejected`) for:

- corrupted/unparseable files
- unresolved participant identities
- context mismatches and duplicate conflicts

### 8. Transparency and Auditability

The system must expose enough detail for operators and users to understand:

- which platform accounts are linked and why
- whether a replay was accepted, rejected, or held for review
- which replay evidence produced each stored result/stat line

## Benefits Expected

- trustworthy automated standings and stat generation
- less manual score reporting and lower operator burden
- consistent identity mapping across queues, scrims, and league matches
- explainable, auditable outcomes for disputes and corrections

## Non-Goals (For This Concept Doc)

- rewriting or replacing `kinetic-rl-parser` internals
- anti-cheat or game integrity enforcement policy
- full UI/UX specifications for all replay-management screens
- final schema-level migration sequencing for every downstream table
