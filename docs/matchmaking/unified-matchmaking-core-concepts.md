# Unified Matchmaking Core Concepts

Date: 2026-02-14  
Status: Draft  
Source Proposal: <https://minor-league-esports.github.io/knowledgeBase/departments/development/features-and-designs/sprocket-v2-unified-matchmaking-proposal/>

## Purpose

Define the core concepts behind Sprocket's unified matchmaking and rating design so product, engineering, and operations share the same model.

## Problem Statement

The legacy model fragments matchmaking and rating by skill group, which causes:

- isolated rating pools that cannot be compared meaningfully
- queue fragmentation and longer waits
- unstable promotion/demotion behavior for boundary ("bubble") players

## Core Philosophy

Skill is continuous, not discrete.

The system uses one continuous rating spectrum for all players. Skill groups remain important for organization and identity, but they are overlays on that spectrum, not hard barriers for matchmaking.

## Concept Model

### 1. Unified Rating Pool

- Every player has one primary rating value (`unified_elo_rating`)
- Rating is comparable across the full ecosystem
- Skill group membership is derived from rating range + hysteresis thresholds

### 2. Skill Group Ranges (Overlapping)

These ranges define display/organization bands and transition zones:

- Foundation: `0-800`
- Academy: `700-1200`
- Champion: `1100-1600`
- Master: `1500-2000`
- Premier: `1900+`

Overlap zones enable fair cross-group matching near boundaries while preserving group identities.

### 3. Hysteresis-Based Transitions

Promotion and demotion must use separate thresholds (buffered) to prevent oscillation.

Example from proposal:

- Foundation -> Academy promotion above `820`
- Academy -> Foundation demotion below `680`

This keeps boundary players from flipping groups due to small match-to-match variance.

### 4. Single Master Queue

- Players can queue for multiple supported scrim types simultaneously
- Queue entries remain per mode/type
- Matchmaking logic evaluates one combined eligible player pool per game/mode, sorted by rating proximity and queue age

### 5. Time-Based Search Expansion

Search tolerance expands with queue time:

1. `0-2 min`: same group preference, `+/-100`
2. `2-5 min`: `+/-150`, allow cross-group near boundaries (within `50`)
3. `5+ min`: `+/-250`, prioritize match quality over label boundaries
4. `10+ min`: `+/-400` emergency expansion

This yields fast matches when population is low while preserving quality-first behavior early.

### 6. Rating + Uncertainty

Rating updates are not based on outcome alone.

Each player tracks:

- rating value
- rating uncertainty (`elo_uncertainty`, Glicko-like RD)

Higher uncertainty permits larger movement and broader early matching; lower uncertainty stabilizes established ratings.

### 7. Transparent Player UX

The system should explain:

- current rating and skill group
- promotion and demotion thresholds
- cross-group eligibility and active search band
- expected wait-time and why a match was formed

## Benefits Expected

- lower average queue times through larger effective pools
- better balance at boundaries via rating-first matching
- improved skill mobility without artificial silos
- clearer player understanding of progression

## Non-Goals (For This Concept Doc)

- exact database schema and migration plan
- final parameter tuning for all queue types
- full UI specification
- rollout sequencing and ops playbooks
