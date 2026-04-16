---
phase: testing
title: Random Tournament - Test Plan
description: Test coverage for player tier system and random tournament feature
feature: random-tournament
created: 2026-04-16
---

# Test Plan: Random Tournament

## Scope
Unit tests for all pure business logic in the random tournament feature. Integration tests deferred (require test DB setup).

## Test Files

| File | Package | Coverage Target |
|------|---------|----------------|
| `backend/internal/service/handicap_test.go` | `service` | `EffectiveWinner` — 100% |
| `backend/internal/service/round_robin_test.go` | `service` | `GenerateRoundRobin` — 100% |
| `backend/internal/service/team_assigner_test.go` | `service` | `AssignTeams1v1`, `AssignTeams2v2` — 100% |
| `backend/internal/service/tournament_service_test.go` | `service` | `CreateTournament` validation, schedule generation logic |

## Unit Test Cases

### EffectiveWinner
- Team1 wins by score
- Team2 wins by score
- Draw (equal scores)
- Team1 wins after handicap applied (tie score but team1 has handicap)
- Team2 wins after handicap applied
- Draw after handicap applied (exact tie after subtraction)
- Zero handicap both sides → pure score comparison

### GenerateRoundRobin
- N=3 (odd): 3 rounds, 1 match each = 3 total matches
- N=4 (even): 3 rounds, 2 matches each = 6 total matches
- N=5 (odd): 5 rounds, 2 matches each = 10 total matches (with bye)
- N=2: 1 round, 1 match
- Each player appears exactly once per round
- Total matches = N*(N-1)/2

### AssignTeams2v2
- No Pros: pairs randomly, all teams balanced
- 1 Pro, rest Normal: Pro paired with Normal
- 1 Pro, 1 Noop, rest Normal: Pro paired with Noop
- 2 Pros, 2 Noops: each Pro gets a Noop
- All Pros: fallback pairing allowed (no error)
- Odd count: returns error
- Correct number of teams returned

### AssignTeams1v1
- Returns one team per player
- Each team has exactly 1 player
- Players shuffled (randomized)

## Run Command
```bash
cd backend && go test ./internal/service/... -v -run "TestEffectiveWinner|TestGenerateRoundRobin|TestAssignTeams|TestCreateTournamentRequest|TestTeamHandicap|TestGenerate"
cd backend && go test ./internal/service/... -cover -run "TestEffectiveWinner|TestGenerateRoundRobin|TestAssignTeams|TestCreateTournamentRequest|TestTeamHandicap|TestGenerate"
```

## Coverage Results (as of 2026-04-16)

| Function | File | Coverage |
|----------|------|----------|
| `EffectiveWinner` | handicap.go | **100%** |
| `GenerateRoundRobin` | round_robin.go | **100%** |
| `AssignTeams1v1` | team_assigner.go | **100%** |
| `AssignTeams2v2` | team_assigner.go | **100%** |
| `affectsScore` | tournament_service.go | **100%** |
| `generate1v1Schedule` | tournament_service.go | **100%** |
| `generate2v2Schedule` | tournament_service.go | **100%** |
| `teamHandicap` | tournament_service.go | **100%** |
| `CreateTournament` | tournament_service.go | 0% (needs DB) |
| `RecordMatchResult` | tournament_service.go | 0% (needs DB) |
| `CompleteTournament` | tournament_service.go | 0% (needs DB) |
| `DeleteTournament` | tournament_service.go | 0% (needs DB) |

Total tests: **46** — all passing ✅

## Deferred Tests (require test DB)
- `TournamentService.CreateTournament` — integration test with DB
- `TournamentService.RecordMatchResult` — integration test
- `TournamentService.DeleteTournament` — score revert verification
- Full API handler tests (Gin test server)

## Known Edge Cases
- `GenerateRoundRobin(1)` — only 1 player: 0 rounds returned
- `AssignTeams2v2` with all Pros and insufficient non-Pros: fallback pairing
- Handicap exactly equal to score difference: should be a draw (0)
- Float precision: 2 - 0.5 = 1.5, not 1.4999...
