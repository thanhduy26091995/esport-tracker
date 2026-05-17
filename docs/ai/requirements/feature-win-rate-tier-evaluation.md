---
phase: requirements
title: Win Rate & Tier Evaluation
description: Show each player's win rate calculated from match history and auto-evaluate tier (pro/normal/noob)
---

# Requirements & Problem Understanding

## Problem Statement

Currently players have a `tier` field that is set manually (defaults to `'normal'`). There is no objective, data-driven way to evaluate a player's skill level, and win rate is never surfaced anywhere in the UI. This makes it hard to tell who is improving, who is dominant, and who is struggling — reducing the competitive transparency of the app.

**Affected users:** All players and admins who use the leaderboard or player list to assess relative skill.

**Current workaround:** Admins manually set `tier` or rely on `current_score` alone as a proxy for skill.

## Goals & Objectives

**Primary goals:**
- Calculate each player's win rate from their full match history (all 1v1 + 2v2 matches)
- Automatically set `tier` to `pro`, `normal`, or `noob` based on win rate thresholds
- Display win rate and tier on UsersView (player table) and DashboardView (leaderboard panel)

**Secondary goals:**
- Show "unranked" state for players with fewer than 10 matches to avoid misleading small-sample tiers
- Recalculate and persist tier automatically each time a match is recorded or deleted

**Non-goals:**
- Separate win rates per match type (1v1 vs 2v2)
- Manual tier overrides (tier becomes fully computed)
- Historical win rate trend charts
- Win rate filtering or sorting (may follow in a future feature)

## User Stories & Use Cases

- As a **player**, I want to see my win rate percentage so I can track my performance over time.
- As a **player**, I want to know my tier (pro/normal/noob) so I have a clear label for my skill level.
- As a **viewer of the leaderboard**, I want to see win rates next to each player so I can compare skill beyond just the score ranking.
- As an **admin**, I want tier to be automatically maintained so I don't need to update it manually.

**Edge cases:**
- Player has 0 matches → show `—` for win rate, show `—` for tier (no badge)
- Player has 1–9 matches → show `—` for both win rate and tier (same as 0; insufficient sample)
- Player has exactly 10 matches with 6 wins (60%) → win rate and tier are both revealed: `pro`
- A match is deleted and a player's win rate drops below a threshold → tier downgrades automatically

## Success Criteria

- Win rate (%) is visible in UsersView player table and DashboardView leaderboard
- Tier badge (pro/normal/noob) is displayed per player in both views; players with < 10 matches show `—` for both win rate and tier
- Tier is recalculated and stored automatically on every match create/delete
- Thresholds: pro ≥ 60%, normal 40–59%, noob < 40% (minimum 10 matches)
- On first deploy, all existing users have their tiers backfilled automatically at startup

## Constraints & Assumptions

**Technical:**
- `tier` field already exists (`varchar(10)`, default `'normal'`) — reused, no migration needed for the field itself
- Win rate is computed from `match_participants` JOIN `matches`; no new table required
- Win rate is returned as a computed field in API responses (not stored in DB)

**Business:**
- Tier evaluation is purely statistical — no manual override
- All match types count equally toward win rate

**Assumptions:**
- A "win" = `match_participants.point_change > 0` (winner team)
- A "loss" = `match_participants.point_change < 0` (loser team)
- Draw matches (`WinnerTeam = 0`, `point_change = 0` for all participants) are **excluded** from win rate — win rate = wins / (wins + losses), draws are ignored entirely
- Tournament matches count equally toward win rate (all match types treated the same)
- Inactive users (`is_active = false`) still show win rate (historical data is preserved)
- On first deploy, a one-time backfill runs at startup to recalculate all existing tiers

## Questions & Open Items

- Should the leaderboard API support sorting by win rate in future? *(Non-goal for this feature; open for a follow-up feature)*
