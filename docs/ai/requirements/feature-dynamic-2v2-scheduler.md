---
phase: requirements
title: Requirements – Dynamic 2v2 Scheduler
feature: dynamic-2v2-scheduler
---

# Requirements & Problem Understanding

## Problem Statement

The current 2v2 tournament scheduler assigns players into **fixed teams** at creation time, then runs a round-robin between those teams. For 6 players this produces only **3 matches** (3 fixed teams × C(3,2)=3 matchups). This means:

- Players always have the same teammate — no variety
- Only 3 opponent pairs are explored out of C(6,2)=15 possible
- Sit-outs are unbalanced (same team always sits the same round)
- The "Pro paired with Noop" rule is locked in once and not re-applied

**Who is affected:** All tournament participants — they expect fair, varied matchups across all rounds.

**Current workaround:** None; users manually track missing matchups outside the app.

## Goals & Objectives

### Primary goals
- Each round re-forms 2 teams of 2 from 4 of the N players (others sit out)
- Generate rounds automatically until **every pair of players has been opponents at least once** (covers all C(N,2) pairs)
- Rotate sit-outs as evenly as possible
- Minimize repeated teammates and repeated opponents across rounds
- Tier-balance each round: Pro players should be paired with Noop/Normal partners when possible

### Secondary goals
- Per-player standings (W/D/L/Points) since teams are no longer fixed
- All existing result-recording, handicap calculation, and score-integration flows remain unchanged

### Non-goals
- Optimising for "minimal number of rounds" (greedy coverage is sufficient)
- Changing 1v1 tournament scheduling (unaffected)
- Support for bracket/elimination formats

## User Stories & Use Cases

| # | Story |
|---|-------|
| US-1 | As an organizer, I create a 2v2 tournament with 6 players and the system generates a full schedule where every pair has faced each other as opponents at least once. |
| US-2 | As a player, I see which round I sit out so I can plan accordingly. |
| US-3 | As an organizer, I record each round's result and the individual standings update automatically. |
| US-4 | As a player, I can see my personal W/D/L record in the tournament standings. |

### Key workflow
1. Organiser selects players (3–16)
2. Selects match type = `2v2`
3. System generates schedule (rounds until all C(N,2) opponent pairs covered)
4. Each round shows: Team1 (player1 + player2) vs Team2 (player3 + player4), sit-out players
5. Organiser records actual score per round
6. Standings show per-player stats

## Success Criteria

- For 6 players: ≤ 10 rounds, all 15 opponent pairs covered
- For N players: all C(N,2) opponent pairs covered
- No player sits out more than `ceil(rounds / 3)` rounds (for 6-player games)
- Pro players are never paired as teammates when a non-Pro is available in the same round
- All 46 existing tests still pass; new scheduler tests ≥ 10 cases

## Constraints & Assumptions

- Min 4 players (need 4 to play 2v2), max 16
- 2v2 requires even number of players (enforced)
- Tier constraint applies per-round (not globally fixed at creation)
- `HandicapRate` for a team = min of team members (existing rule, unchanged)
- If N is not divisible by 4, some rounds have sit-outs (e.g. 6 players → 2 sit out per round)
- Backend only change — no DB model changes needed
- Frontend standings table needs to switch from team-based to player-based view

## Questions & Open Items

- ✅ Number of rounds: auto (cover all opponent pairs)
- ✅ Standings: per-player W/D/L
- ✅ Tier balancing: per-round, Pro paired with Noop/Normal preferred
- ❓ If N=5 (odd, should be blocked for 2v2), skip — already validated upstream
- ❓ For large N (12–16 players): multiple matches per round? → **No**, keep 1 match per round for simplicity (2 teams of 2 play, rest sit out)
