---
phase: testing
title: Testing – Dynamic 2v2 Scheduler
feature: dynamic-2v2-scheduler
---

# Testing Strategy

## Test Files

| File | Status |
|------|--------|
| `backend/internal/service/dynamic_scheduler_test.go` | TODO |

## Unit Tests

### Coverage completeness
- `TestAllOpponentPairsCovered_4Players` — 4 players → C(4,2)=6 pairs, verify all covered
- `TestAllOpponentPairsCovered_6Players` — 6 players → 15 pairs, verify all covered, ≤15 rounds
- `TestAllOpponentPairsCovered_8Players` — 8 players → 28 pairs, verify all covered

### Fairness / balance
- `TestSitOutBalance_6Players` — no player sits out more than `ceil(rounds/3)` times
- `TestNoPlayerAppearsInBothTeams` — each round, all 4 players are distinct
- `TestTeamsAlwaysHave2Players` — team1 and team2 each have exactly 2 players

### Tier balance
- `TestTierBalance_ProPairedWithNonPro` — when NonPro available, Pro is never alone with another Pro
- `TestTierBalance_AllNormal` — all Normal players → still generates valid schedule
- `TestTierBalance_AllPro_FallbackAllowed` — all Pro → Pro-Pro pairing allowed as fallback

### Edge cases
- `TestMinPlayers_4Players` — minimum 4 players works correctly
- `TestRoundsTerminate` — no infinite loop for any N in [4..16]
- `TestHandicapAppliedPerRound` — each TournamentMatch has correct handicap (min of team members)

## Integration with existing tests

Run: `go test ./internal/service/... -timeout 120s`

All 46 existing tests must remain green after wiring new scheduler into `tournament_service.go`.
