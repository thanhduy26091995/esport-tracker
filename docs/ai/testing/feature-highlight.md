---
phase: testing
title: Player Highlight Feed — Testing Strategy
description: Test scope, coverage targets, and validation criteria for all 8 highlight categories
---

# Testing Strategy

## Scope

- **Unit tests:** `HighlightService` rule engine — all 8 categories, thresholds, section assignment, priority scoring, social variant selection
- **Integration tests:** `HighlightRepository` — all 9 query methods against a real test DB
- **Manual E2E:** Record matches to trigger each highlight type; verify all appear in the correct dashboard section

## Test Files

| File | Package/Layer | Coverage Target |
|---|---|---|
| `backend/internal/service/highlight_service_test.go` | Service / rule engine | ≥ 90% |
| `backend/internal/repository/highlight_repository_test.go` | Repository / SQL | ≥ 80% |

## Unit Tests

### Category 1 — Streaks
- Win streak 3 → `streak_win`, section `trending`, priority 60
- Win streak 5 → priority 80
- Win streak 10 → priority 95
- Lose streak 5 → `streak_lose`, section `daily_recap`
- Unbeaten streak 8 (wins only, no losses) → `streak_unbeaten`
- Streak broken by opponent → `streak_broken_win` emitted for opponent
- Player 0 matches → no streak highlights

### Category 2 — Rank / Point Movement
- pointsGainedToday = 120 → `points_gained_today`, section `daily_recap`
- pointsGainedToday = 0 (no matches) → no highlight
- pointsLostToday = -95 → `points_lost_today`
- Rank moved up 3 positions → `rank_climbed`, section `competitive`
- Rank dropped 4 positions → `rank_dropped`
- pointsGainedToday highest among all users → `fastest_climber_today`, section `trending`

### Category 3 — Recent Form
- last10WinRate = 0.80 → `form_hot`, section `trending`
- last10WinRate = 0.40 → `form_cold`, section `daily_recap`
- last10 stddev low (≤ 1 flip) → `form_stable`, section `social`
- < 10 total decisive matches → no form highlight

### Category 4 — Activity
- matchesToday highest → `most_active_today`, section `trending`
- matchesToday ≥ 20 → `marathon`, section `social`
- consecutiveActiveDays ≥ 5 → `active_streak_days`, section `social`
- matchesToday = 0 → no activity highlight

### Category 5 — Hot / Cold
- last10WinRate = 0.80, matchesToday = 5 → `hot_player` generated
- last10WinRate = 0.80, matchesToday = 4 → no hot_player (below threshold)
- last10WinRate = 0.50, matchesToday = 10 → no hot_player
- loseStreak = 5 → `cold_player` generated
- loseStreak = 4 → no cold_player
- < 10 total matches → neither hot nor cold

### Category 6 — Fast Climb / Collapse
- delta1h = +50 → `fast_climb_hour`, section `trending`
- delta1h = -40 → `fast_collapse_hour`, section `competitive`
- delta1h = 0 (no matches in last hour) → no highlight

### Category 7 — Milestones
- totalWins == 100 → `milestone_wins` value=100, section `social`
- totalWins == 101 → no milestone (already past)
- totalWins == 500 → `milestone_wins` value=500
- currentScore == 1000 → `milestone_points`
- totalMatches == 500 → `milestone_matches`
- rank enters top 10 (was 11, now 10) → `milestone_top10`, section `competitive`

### Category 8 — Social / Fun
- hot_player conditions met → `social_cooking` emitted
- loseStreak ≥ 3 AND matchesToday ≥ 5 → `social_tilt`
- matchesToday ≥ 10 → `social_tryhard`
- matchesToday ≥ 15 → `social_grinder`
- rankClimbedToday ≥ 2 AND matchesToday ≤ 5 → `social_sneaky`
- Social messages: calling `buildMessage(social_cooking, "Player X")` returns one of the known variants (test all variants exist in the pool)

### Priority + Section Cap
- 10 trending highlights eligible → only top 5 returned
- 20 total highlights → only top 20 (5 per section) returned
- Two highlights same priority within section → stable order (alphabetical by player name)

## Integration Tests

| Query | Test scenario |
|---|---|
| `GetCurrentStreaks` | 5 consecutive wins → streak=5; add 1 loss → win streak resets to 0 |
| `GetLongestStreaks` | Historical 8-win streak, current 3-win streak → longest=8, current=3 |
| `GetPointsMovementToday` | Matches on today + yesterday → only today's sum returned |
| `GetPointsMovementLastHour` | Matches 30 min ago + 2 hours ago → only 30-min match counted |
| `GetRankSnapshot` | User ranked 5th today, was 8th yesterday → rankChange=+3 |
| `GetRecentForm` | 10 decisive matches (8W 2L) → rate=0.80, draws excluded |
| `GetWeeklyActivity` | 3 consecutive days active → consecutiveActiveDays=3; gap breaks streak |
| `GetTotals` | 99 wins + 1 new win inserted → totalWins=100 (milestone boundary) |
| `GetLast10WinRate` | Draws do not count toward last-10 window |

## Test Data & Environments

- Follows existing `_test.go` pattern (see `user_repository_test.go`)
- Each test inserts its own match/participant rows and cleans up after
- Standard `DB_*` env vars — no additional setup

## Execution

```bash
# Run all highlight tests
go test ./internal/service/... -run Highlight -v
go test ./internal/repository/... -run Highlight -v

# Coverage report
go test ./internal/service/... -coverprofile=cover.out
go tool cover -func=cover.out
```

## Coverage & Quality Gates

- `HighlightService`: ≥ 90% line coverage — every category rule path exercised
- `HighlightRepository`: ≥ 80% line coverage — all 9 queries covered
- All threshold boundaries (e.g. streak 4 vs 5, winRate 0.79 vs 0.80) must have explicit test cases
- Social variant pool: assert all defined variants are reachable (iterate pool, not just rand output)

## Risks & Gaps

- Milestone "exactly at threshold" requires inserting matches to cross the boundary (not seeding the total directly), since milestone detection compares against live query totals
- Fast-climb query uses `NOW()` — integration tests should insert matches with explicit timestamps to avoid clock-dependent flakiness
- `HighlightFeedPanel.vue` has no automated tests (consistent with other panel components in this project)
