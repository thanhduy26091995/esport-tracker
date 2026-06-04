---
phase: requirements
title: "1v2 Match Type Support"
description: Add a new asymmetric match format where one player faces two players, with special scoring rules
---

# Requirements & Problem Understanding

## Problem Statement

Currently the system only supports two symmetric match types:
- `1v1`: 1 player vs 1 player
- `2v2`: 2 players vs 2 players

There is a real gameplay need for **1 vs 2** matches — where a single strong player takes on two opponents. The scoring must reflect the asymmetry: the solo player is rewarded more heavily for winning against two, and each team-2 player earns a normal win point when they beat the solo.

Current workaround: none — this match type simply cannot be recorded.

## Goals & Objectives

**Primary goals:**
- Add `1v2` as a valid `match_type` value across backend and frontend
- Implement asymmetric scoring: solo wins → +2 pts; duo wins → +1 pt each
- Apply symmetric deductions to losers (zero-sum per side)
- Enforce correct team-size validation (team1 = 1 player, team2 = 2 players)

**Non-goals:**
- Support for `2v1` (team2 is always the duo; the solo is always team1)
- Custom point overrides per-player within a 1v2 match
- Tournament support for 1v2 matches (out of scope for this feature)

## User Stories & Use Cases

- **As a match recorder**, I want to select "1v2" as the match type so that I can record a match where one player faces two opponents.
- **As a match recorder**, I want the system to enforce that team 1 has exactly 1 player and team 2 has exactly 2 players when `1v2` is selected.
- **As a player**, I want my score to increase by 2 when I win a 1v2 match solo, reflecting the difficulty of defeating two opponents.
- **As a player on the duo side**, I want my score to increase by 1 when my team wins the 1v2 match.
- **As a match recorder**, I want validation errors if I try to put 2 players on team1 in a 1v2 match.

**Edge cases:**
- Draw (winner_team = 0): no points change, same as other match types
- Deletion of a 1v2 match must correctly revert asymmetric point changes
- The `points_per_win` override (if provided) acts as the base unit: solo win = 2×base, duo win = 1×base each

## Success Criteria

- `POST /api/v1/matches` accepts `match_type: "1v2"` with `team1: [id]` and `team2: [id1, id2]`
- Score changes after a solo win: team1 player `+2*base`, each team2 player `-1*base`
- Score changes after a duo win: each team2 player `+1*base`, team1 player `-2*base`
- Invalid team sizes return HTTP 400 with error code `INVALID_TEAM_SIZE`
- Frontend match form displays "1v2" option and enforces the correct team slot counts
- Existing `1v1` and `2v2` behaviour is unchanged

## Constraints & Assumptions

- The solo player is always on **team1**, the duo is always on **team2** — this is a convention, not validated by label
- The scoring formula is zero-sum per total points: `solo_win_points = 2 × base`, `duo_each_points = 1 × base`, so both sides move the same total number of points
- The `points_per_win` config/request value serves as the base unit (default 1)
- Losing deductions mirror winning gains symmetrically (zero-sum)
- DB schema change: `match_type` column is `varchar(10)`, "1v2" fits within that limit — **no schema migration needed**

## Questions & Open Items

- Should `1v2` be allowed inside tournament rounds? (Assumed: **No** for now — tournament service validates match types)
- Should the frontend label show "1 vs 2" or "1v2" for clarity?
- Is draw (0-0) a valid outcome for 1v2? (Assumed: **Yes**, same as other types — no points change)
