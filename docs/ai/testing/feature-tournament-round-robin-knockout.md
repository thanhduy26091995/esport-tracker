---
phase: testing
title: Round Robin + Top N Knockout Tournament — Testing Strategy
description: Test scope, unit and integration test plan for the round-robin configurable-knockout tournament format
feature: tournament-round-robin-knockout
created: 2026-06-18
updated: 2026-06-18
---

# Testing Strategy

## Scope

- **Unit tests**: schedule generation algorithm, standings computation (including knockout size variants)
- **Integration tests**: create tournament API, generate-knockouts API, record result API with auto-final-creation
- **No E2E tests** (project pattern: no E2E framework in use)

## Test Files

| File | Package/Layer | Status | Coverage Target |
|------|---------------|--------|----------------|
| `backend/internal/service/round_robin_teams_test.go` | service | ✅ Written | `GenerateTeamSchedule` correctness |
| `backend/internal/service/tournament_standings_test.go` | service | ✅ Written | `ComputeStandings` correctness incl. knockoutSize param |
| `backend/internal/api/tournament_handler_test.go` | handler/integration | ⬜ Deferred | Create, GenerateKnockouts, RecordResult flows |

> **Note:** Pre-existing WC test failures in `poisson_service_test.go` and `wc_integration_test.go`
> (stale field references) prevent the package from building in test mode. Those must be fixed
> before `go test ./internal/service/...` will run. The new test files themselves compile cleanly
> (verified via `go build ./internal/service/` and `go vet`).

## Unit Tests

### `GenerateTeamSchedule` (`round_robin_teams_test.go`)

| Test | Description | Status |
|------|-------------|--------|
| `TestGenerateTeamSchedule_5Teams_10Matches` | 5 teams → C(5,2) = 10 matches | ✅ |
| `TestGenerateTeamSchedule_4Teams_6Matches` | 4 teams → 6 matches | ✅ |
| `TestGenerateTeamSchedule_6Teams_15Matches` | 6 teams → 15 matches | ✅ |
| `TestGenerateTeamSchedule_NoPairRepeated` | No (A,B) pair appears twice, for n=4..8 | ✅ |
| `TestGenerateTeamSchedule_EachTeamPlaysNMinus1Times` | Each team plays exactly n-1 times | ✅ |
| `TestGenerateTeamSchedule_RoundAndOrderSet` | Round ≥ 1, Order ≥ 1 on every slot | ✅ |
| `TestGenerateTeamSchedule_5Teams_5Rounds2MatchesEach` | 5 rounds × 2 matches for 5 teams | ✅ |
| `TestGenerateTeamSchedule_NoZeroUUIDTeamIDs` | No slot has zero-UUID (covers the ID fix) | ✅ |
| `TestGenerateTeamSchedule_TeamDataPreserved` | All input team IDs appear in output | ✅ |

### `ComputeStandings` (`tournament_standings_test.go`)

| Test | Description | Status |
|------|-------------|--------|
| `TestComputeStandings_EmptyTeamsAndMatches` | nil input returns empty slice | ✅ |
| `TestComputeStandings_OneTeam_NoMatches` | 1 team, no matches → 0 pts | ✅ |
| `TestComputeStandings_Team1WinsAllMatches_12Points` | 4 wins → 12 pts, seed=1 | ✅ |
| `TestComputeStandings_DrawGivesOnePointEach` | Draw → 1 pt each, Drawn++ | ✅ |
| `TestComputeStandings_GoalDifference` | GF/GA/GD calculated correctly | ✅ |
| `TestComputeStandings_SortByPointsPrimary` | Higher points ranks first | ✅ |
| `TestComputeStandings_SortByGDSecondary` | Same points → higher GD ranks first | ✅ |
| `TestComputeStandings_SortByGFTertiary` | Same pts+GD → higher GF ranks first | ✅ |
| `TestComputeStandings_Top4Seeds` | knockoutSize=4 → 4 teams get seed, 5th = 0 | ✅ |
| `TestComputeStandings_Top2Seeds_KnockoutSize2` | knockoutSize=2 → only top 2 seeded | ✅ |
| `TestComputeStandings_SeedOrderMatchesSortOrder` | seed 1-4 matches sort order | ✅ |
| `TestComputeStandings_PlayedCountAccumulates` | Played/Won/Drawn across multiple matches | ✅ |
| `TestComputeStandings_PendingMatchExcluded` | Pending match → not counted | ✅ |
| `TestComputeStandings_NonGroupStageMatchExcluded` | stage="semi" → ignored | ✅ |
| `TestComputeStandings_ZeroUUIDTeamIDSkipped` | Zero-UUID match → no panic, skipped | ✅ |
| `TestComputeStandings_NilTeamIDSkipped` | nil TeamID pointer → no panic, skipped | ✅ |

## Integration Tests (Deferred)

### Create Tournament — `round_robin_top4`

- **Happy path top-4**: POST with 5 teams + `knockout_size=4` → 201, `knockout_size=4` in response, 10 group matches
- **Happy path top-2**: POST with 4+ teams + `knockout_size=2` → 201, `knockout_size=2` in response
- **KnockoutSize defaults to 4**: POST without `knockout_size` → `knockout_size=4`
- **Invalid knockoutSize 3**: POST with `knockout_size=3` → normalized to 4 (backend coerces unknown values)
- **Duplicate player**: same player UUID in two teams → 400
- **Classic format unchanged**: no `knockout_size` in classic response

### Generate Knockouts — Top 4

- **Happy path**: all group matches completed → 2 `stage="semi"` matches; seed1+seed4, seed2+seed3
- **Group matches incomplete**: one pending → 400
- **Already generated**: call twice → 409 / error
- **Tie at boundary**: teams at 4th/5th position tied → 400 with tie message

### Generate Knockouts — Top 2

- **Happy path**: all group matches completed, `knockout_size=2` → 1 `stage="final"` match; seed1 vs seed2
- **No semis created**: response has no `stage="semi"` matches
- **Tie at boundary**: 2nd and 3rd tied → 400 with tie message

### Record Result — Auto-create Final (Top 4 only)

- **First semi recorded**: 1 semi done → no final/3rd-place yet
- **Both semis recorded**: 2nd semi → final + third_place matches auto-created
- **Idempotency**: recording a completed semi twice → no duplicate final

### Record Result — Champion (Top 2)

- **Final recorded**: winner of `stage="final"` → `champion_team_id` set on tournament

## Test Data & Environments

- Unit tests: no database, pure in-memory struct construction
- Integration tests (deferred): use PostgreSQL test DB or testcontainers following existing `wc_integration_test.go` pattern
- Fixture: `makeTeam()`, `makeTournamentTeam()`, `completedGroupMatch()` helpers defined in test files

## Execution

```bash
# After fixing pre-existing WC test failures:
cd backend && go test ./internal/service/... -run "TestGenerateTeamSchedule" -v
cd backend && go test ./internal/service/... -run "TestComputeStandings" -v
cd backend && go test ./internal/service/... -v
```

## Coverage & Quality Gates

- `GenerateTeamSchedule`: 100% branch coverage (pure function, no external deps)
- `ComputeStandings`: 100% branch coverage — all nil checks, stage filters, sort comparisons, seed paths
- `knockoutSize=2` path must be explicitly covered (regression against hardcoded top-4 assumption)
- Zero-UUID regression test must remain green (covers the `ID: uuid.New()` fix)

## Risks & Gaps

- **Pre-existing WC test failures** block running the package test suite — `poisson_service_test.go` and `wc_integration_test.go` reference stale model fields; fix before CI
- **Integration tests deferred**: create/generate/record flows not yet covered by automated tests; manual QA required until written
- **Frontend not covered**: Vue component rendering, bracket display, and knockout size selector are not automatically tested
- **Head-to-head tiebreaker gap**: standings sort by GF as final tiebreaker — two equal teams will have undefined relative order; documented as known v1 limitation
- **Concurrent generate-knockouts**: no concurrency test; low risk for a friend-group app
