---
phase: planning
title: Planning – Dynamic 2v2 Scheduler
feature: dynamic-2v2-scheduler
---

# Project Planning & Task Breakdown

## Milestones

- [ ] **M1**: Core scheduler algorithm implemented and unit tested
- [ ] **M2**: Tournament service wired to new scheduler, existing tests green
- [ ] **M3**: Frontend standings updated to per-player view

## Task Breakdown

### Phase 1: Backend — New Scheduler Algorithm

- [ ] **p1-dynamic-scheduler**: Create `backend/internal/service/dynamic_scheduler.go`
  - Define `RoundSlot`, `pairKey`, `schedulerState` types
  - Implement `GenerateSchedule2v2(players []*model.User) []RoundSlot`
  - Implement `pickBestRound(players, state)` greedy selector
  - Implement scoring: newOpponentPairs, repeatTeammate, repeatOpponent, sitOutImbalance
  - Tier-balance override logic (Pro must have NonPro partner when available)
  - Termination: stop when all C(N,2) opponent pairs covered

- [ ] **p1-scheduler-tests**: Create `backend/internal/service/dynamic_scheduler_test.go`
  - TestAllOpponentPairsCovered_6Players
  - TestAllOpponentPairsCovered_4Players
  - TestAllOpponentPairsCovered_8Players
  - TestSitOutBalance_6Players (no player sits out >ceil(rounds/3) times)
  - TestTierBalance_ProNeverAloneWithPro (when nonPro available)
  - TestTeamsAlwaysHave2Players
  - TestNoPlayerAppearsInBothTeams (same round)
  - TestRoundsTerminate (no infinite loop)
  - TestHandicapMin_AppliedPerRound

### Phase 2: Wire into Tournament Service

- [ ] **p2-tournament-service**: Update `tournament_service.go`
  - Replace `generate2v2Schedule(users)` to call `GenerateSchedule2v2`
  - Map `[]RoundSlot` → `[]model.TournamentMatch` (preserving Round, MatchOrder, Handicap fields)
  - Remove or deprecate `AssignTeams2v2` if no longer needed

- [ ] **p2-regression-tests**: Run full service test suite; ensure all 46 existing tests pass

### Phase 3: Frontend — Per-Player Standings

- [ ] **p3-standings**: Update `TournamentDetailView.vue`
  - Rewrite `computeStandings` to aggregate per-player (not per-fixed-team)
  - Each player gets W/D/L credit for every round their team wins/draws/loses
  - Points: Win=3, Draw=1, Loss=0
  - Sort by Pts desc, then GD (goal difference) desc

## Dependencies

```
p1-dynamic-scheduler → p1-scheduler-tests (write tests after implementation)
p1-dynamic-scheduler → p2-tournament-service
p2-tournament-service → p2-regression-tests
p2-tournament-service → p3-standings (can be parallel)
```

## Risks & Mitigation

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Greedy doesn't terminate for edge cases | Low | Add hard cap: maxRounds = N*(N-1)/2 + 5 |
| Performance on N=16 | Low | Pre-benchmark; greedy is O(N^4) per round |
| Frontend standings logic regression | Medium | Add snapshot test with known fixture |
