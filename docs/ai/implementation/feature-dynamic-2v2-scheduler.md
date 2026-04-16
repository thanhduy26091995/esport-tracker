---
phase: implementation
title: Implementation – Dynamic 2v2 Scheduler
feature: dynamic-2v2-scheduler
---

# Implementation Guide

## Code Structure

```
backend/internal/service/
├── dynamic_scheduler.go       NEW — core greedy algorithm
├── dynamic_scheduler_test.go  NEW — unit tests
├── tournament_service.go      MODIFIED — generate2v2Schedule
├── team_assigner.go           UNCHANGED (used by 1v1 only now)
└── round_robin.go             UNCHANGED (1v1 only)

frontend/src/views/
└── TournamentDetailView.vue   MODIFIED — computeStandings per-player
```

## Implementation Notes

### `GenerateSchedule2v2`

```go
func GenerateSchedule2v2(players []*model.User) []RoundSlot {
    state := newSchedulerState(players)
    maxRounds := len(players)*(len(players)-1)/2 + 5  // safety cap

    for round := 0; round < maxRounds && !state.allCovered(); round++ {
        slot := pickBestRound(players, state)
        state.apply(slot)
        slots = append(slots, slot)
    }
    return slots
}
```

### Candidate enumeration

For N=6 players, iterate:
```go
// C(N, 4) subsets of players to play
for each 4-subset {p0, p1, p2, p3}:
    // 3 ways to split into 2v2
    splits := [3][2][2]int{
        {{0,1},{2,3}},  // AB vs CD
        {{0,2},{1,3}},  // AC vs BD
        {{0,3},{1,2}},  // AD vs BC
    }
    for each split:
        score = computeScore(split, sitOuts, state)
        track best
```

### Tier balance check
```go
func isTierBalanced(t1, t2 []*model.User) bool {
    // A split is "balanced" if each team has ≤1 Pro
    // AND if any Pro exists, they have a NonPro partner
    return !allSameTier(t1) || !hasProWithoutNonPro(t1) ...
}
```

### Mapping RoundSlot → TournamentMatch
```go
for i, slot := range slots {
    h1 := teamHandicap(slot.Team1[:], userMap)
    h2 := teamHandicap(slot.Team2[:], userMap)
    match := model.TournamentMatch{
        Round:          i + 1,
        MatchOrder:     1,
        Team1Player1ID: slot.Team1[0],
        Team1Player2ID: &slot.Team1[1],
        Team2Player1ID: slot.Team2[0],
        Team2Player2ID: &slot.Team2[1],
        HandicapTeam1:  h1,
        HandicapTeam2:  h2,
        Status:         "pending",
    }
}
```

## Integration Points

- `tournament_service.go::generate2v2Schedule` → calls `GenerateSchedule2v2`
- `RecordMatchResult` — unchanged; still creates regular Match record per result
- `computeStandings` in Vue — iterates all completed TournamentMatches, credits both team members

## Error Handling

- If `len(players) < 4`: return error (already validated in CreateTournament)
- If greedy hits safety cap without full coverage: log warning, return partial schedule (still valid)
