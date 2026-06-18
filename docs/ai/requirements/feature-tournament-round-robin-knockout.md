---
phase: requirements
title: Round Robin + Top 4 Knockout Tournament Format — Requirements
description: Requirements for a new tournament format with a fixed 5-team round-robin group stage and a top-4 knockout semifinal bracket
feature: tournament-round-robin-knockout
created: 2026-06-18
---

# Requirements & Problem Understanding

## Problem Statement

The current tournament system only supports a dynamic round-robin where players are paired into different teams each match (no fixed teams). For groups of **10 players (5 fixed pairs)**, there is no way to run a proper league-style format where the same 5 teams always compete together and the best teams advance to a playoff.

**Who is affected:** Admin and 10-player friend groups that want a more competitive, structured tournament experience.

**Current workaround:** None — admins must track standings and knockout pairings manually outside the app.

## Goals & Objectives

**Primary goals:**
- Add a "Round Robin + Top 4 Knockout" (`round_robin_top4`) tournament format for exactly 10 players organized into 5 fixed 2-person teams
- Auto-generate all 10 group-stage matches (each team plays every other team once)
- Track live team standings (points, wins, draws, losses, GF, GA, GD) after each group result is recorded
- Let admin trigger knockout generation once all group matches are done — system seeds 1st vs 4th and 2nd vs 3rd in the semis
- Auto-create the Final and 3rd-place match once both semifinal results are recorded

**Secondary goals:**
- Support the existing tier-balanced auto-assignment when no explicit teams are provided
- Allow manual team composition at creation time

**Non-goals:**
- Support for player counts other than 10 (format is specifically 5 teams of 2)
- Multiple groups / group stages
- Penalty-shootout tiebreaker logic
- Head-to-head tiebreaker (v1 uses points → GD → GF only)
- Changes to the existing classic round-robin format

## User Stories & Use Cases

1. As an admin, I want to create a "Round Robin + Top 4" tournament with 10 players so that all teams are guaranteed at least 4 group-stage matches before anyone is eliminated.
2. As an admin, I want to assign players to fixed teams so that the same 2 people always play together throughout the tournament.
3. As an admin, I want to see a live standings table after each group match is recorded so I can track who is in the top 4.
4. As an admin, I want to click "Generate Knockouts" once all 10 group matches are done so the correct top-4 seeding is applied automatically.
5. As a player, I want to see my team's position in the standings (points, GD, goals scored) so I understand how we're performing.
6. As an admin, I want to record semifinal and final results in the same tournament detail page and see the tournament champion declared.

## Success Criteria

- [ ] Creating a tournament with `format="round_robin_top4"` and 10 players auto-generates exactly 10 group-stage matches
- [ ] Each of the 5 teams plays exactly 4 group matches (against each of the other 4 teams, once)
- [ ] Standings correctly compute: Win=3 pts, Draw=1 pt, Loss=0 pt; sorted by pts → GD → GF (descending)
- [ ] Top 4 teams in standings are visually highlighted
- [ ] "Generate Knockouts" button appears only after all 10 group matches are completed
- [ ] "Generate Knockouts" is blocked with a clear error if any two teams on the top-4 boundary are tied on pts + GD + GF
- [ ] Group match results are locked (non-editable) once knockout matches have been generated
- [ ] Knockout bracket: seed1 vs seed4, seed2 vs seed3 in semis; winners meet in the Final, losers in the 3rd-place match
- [ ] Recording both semifinal results auto-creates the Final and 3rd-place match
- [ ] A champion banner with both player names is displayed at the top of the tournament detail page once the Final result is recorded
- [ ] `entry_fee` field behaves identically to the classic format
- [ ] `affects_score` flag applies uniformly to all stages (group + knockout)
- [ ] Existing classic tournament format and all existing data are unaffected

## Constraints & Assumptions

- **Player count**: Exactly 10 players required (5 teams × 2); validation error if not met
- **Match type**: Always 2v2 (team-based)
- **Knockout structure**: Single elimination — 2 semis, 1 final, 1 third-place match (4 extra matches total)
- **Effective winner**: Handicap is still applied to determine match result; standings W/D/L are based on the effective winner
- **Affects score**: Same flag as classic — tournament matches may or may not feed the main leaderboard
- **Points system**: Standard football points (W=3, D=1, L=0)
- **Tiebreaker v1**: GD then GF; head-to-head deferred
- **Backend**: Go/Gin/GORM/PostgreSQL following existing patterns
- **Frontend**: Vue 3/TypeScript/Pinia following existing patterns; all UI strings via vue-i18n

## Questions & Open Items

- [x] Should teams have custom names?
  → **Decision**: Teams identified by their two player names; no separate team name needed in v1
- [x] Is the 3rd-place match mandatory?
  → **Decision**: Always included (standard 4-team knockout)
- [x] What happens if a knockout match is a draw?
  → **Decision**: Admin records the winner manually (no automatic resolution required by the system)
- [x] Should the schedule include round numbers for group matches?
  → **Decision**: Yes, group matches are organized into rounds (5 rounds × 2 matches) so each round has a natural play session
- [x] What happens when teams 4 and 5 tie on all criteria (pts + GD + GF)?
  → **Decision**: Block "Generate Knockouts" and show an error message; admin manually resolves (head-to-head deferred to v2)
- [x] Is handicap applied in knockout matches?
  → **Decision**: Yes — handicap applies uniformly to all stages (group, semi, final, 3rd-place); consistent with existing match logic
- [x] Can a group result be corrected after knockouts are generated?
  → **Decision**: No — group results are locked once knockout matches exist; admin must delete the tournament and recreate if a correction is needed
- [x] How is the tournament champion displayed?
  → **Decision**: A prominent champion banner showing both player names appears at the top of the tournament detail page once the Final result is recorded
- [x] Does this format support entry fee?
  → **Decision**: Yes — same `entry_fee` field as classic format (0 = free, >0 = VND amount)
- [x] Should `affects_score` be configurable per stage?
  → **Decision**: No — single `affects_score` flag applies uniformly to all stages (group + knockout)
