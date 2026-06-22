---
phase: testing
title: WC Group Standings — Testing Strategy
description: Test scope, test cases, and validation criteria for the standings feature
---

# Testing Strategy

## Scope

- **Unit tests**: `GetGroupStandings` service logic (sorting, form assembly, edge cases)
- **Integration tests**: Repository query against real test DB (verify SQL produces correct team stats)
- **Manual / visual tests**: Frontend table rendering, responsiveness, colours

## Test Files

| File | Package/Layer | Coverage Target | Status |
|------|---------------|----------------|--------|
| `backend/internal/service/wc_standings_test.go` | service | `standingLess`, `sortTeamStandings` | ✅ 15 tests, all pass |
| `backend/internal/repository/wc_repository_test.go` | repository | `GetGroupStandings` | ✅ 5 integration tests (require `TEST_DATABASE_URL`) |

## Unit Tests

### Service layer (`wc_standings_test.go`)

**Sorting**:
- Two teams with same points → team with higher GD ranks first
- Two teams same points and same GD → team with more GF ranks first
- Two teams fully equal → sort alphabetically (stable)
- Team with 9 points always ranks above team with 6 points regardless of GD

**Form calculation**:
- Team with 6 completed matches → form shows only last 5 (oldest dropped)
- Team with 2 completed matches → form shows 2 entries
- Team with 0 completed matches → form is empty slice `[]`
- Form order: chronological, oldest first (most recent is rightmost/last in slice)

**Stats accumulation**:
- Home team wins 3-1: home gets W/+3pts/GF3/GA1; away gets L/0pts/GF1/GA3
- Draw 1-1: both get D/+1pt/GF1/GA1
- No score (nil scores) → match is skipped

**Edge cases**:
- Group with no completed matches → returns 4 teams, all with zeroed stats, sorted alphabetically
- Group with only 1 team entry (data anomaly) → returns that team's stats safely

### Points formula

```
Points = Won × 3 + Drawn × 1
GoalDifference = GoalsFor − GoalsAgainst
```

## Integration Tests (Repository)

Extend `backend/internal/repository/wc_repository_test.go`:

- Seed a test group ("Group Z") with 2 completed matches and 4 teams
- Call `GetGroupStandings()`
- Assert:
  - Returned `groups` contains "Group Z"
  - Teams within Group Z have correct played/won/drawn/lost/gf/ga/gd/points
  - Form slice is correct and in the right order
  - `scheduled` and `live` matches are excluded from standings

## Test Data & Environments

**Seeded test matches** for unit tests (in-memory):
```
Match 1: ARG 3-0 BRA  (Group Z, completed, date: 2026-06-11)
Match 2: FRA 1-1 ESP  (Group Z, completed, date: 2026-06-11)
Match 3: ARG 1-1 FRA  (Group Z, completed, date: 2026-06-15)
Match 4: BRA 2-0 ESP  (Group Z, completed, date: 2026-06-15)
Match 5: ARG 1-0 ESP  (Group Z, completed, date: 2026-06-19, scheduled — should NOT count)
```

Expected standings after matches 1–4:
| Rank | Team | P | W | D | L | GF | GA | GD | Pts | Form |
|------|------|---|---|---|---|----|----|-----|-----|------|
| 1 | ARG | 2 | 1 | 1 | 0 | 4 | 1 | +3 | 4 | W D |
| 2 | BRA | 2 | 0 | 0 | 2 | 2 | 5 | -3 | 0 | L L |
| 3 | FRA | 2 | 0 | 2 | 0 | 2 | 2 | 0  | 2 | D D |
| 4 | ESP | 2 | 0 | 1 | 1 | 1 | 2 | -1 | 1 | D L |

Wait — re-check: ARG beats BRA (3pts), draws FRA (1pt) = 4pts. FRA draws ARG (1pt), draws ESP... but wait Match 2 is FRA 1-1 ESP = 1pt for FRA, 1pt for ESP. So:
- ARG: M1 W (vs BRA), M3 D (vs FRA) → 4pts, GF=4, GA=1, GD=+3
- BRA: M1 L (vs ARG), M4 W (vs ESP) → 3pts, GF=2, GA=3, GD=-1 — actually BRA wins M4!
- FRA: M2 D (vs ESP), M3 D (vs ARG) → 2pts, GF=2, GA=2, GD=0
- ESP: M2 D (vs FRA), M4 L (vs BRA) → 1pt, GF=1, GA=2, GD=-1

Correct standings:
| Rank | Team | P | W | D | L | GF | GA | GD | Pts |
|------|------|---|---|---|---|----|----|-----|-----|
| 1 | ARG | 2 | 1 | 1 | 0 | 4 | 1 | +3 | 4 |
| 2 | BRA | 2 | 1 | 0 | 1 | 2 | 3 | -1 | 3 |
| 3 | FRA | 2 | 0 | 2 | 0 | 2 | 2 | 0  | 2 |
| 4 | ESP | 2 | 0 | 1 | 1 | 1 | 2 | -1 | 1 |

BRA and ESP both have GD=-1; BRA ranks higher (3pts vs 1pt).

## Execution

```bash
# Service unit tests (no DB required)
cd backend && go test ./internal/service/... -run "TestStanding|TestSort" -v

# Repository integration tests (requires TEST_DATABASE_URL)
cd backend && go test ./internal/repository/... -run TestGetGroupStandings -v

# All WC-related tests
cd backend && go test ./internal/service/... ./internal/repository/... -v

# Frontend type check (no type-check script — use vue-tsc directly)
cd frontend && npx vue-tsc --noEmit
```

## Coverage & Quality Gates

- Service sorting logic: 100% branch coverage
- Form slice assembly: cover 0, <5, ≥5 completed matches cases
- Repository: verified against real test DB with seeded data
- Frontend: `npm run type-check` passes with no errors on new types and component props

## Risks & Gaps

- **No frontend unit tests**: Vue component testing is not set up in this project. Manual visual verification covers the frontend.
- **Flag emoji mapping**: No automated test for the `teamCodeToFlag` utility — verify manually for a sample of 5–10 team codes.
- **Form ordering**: Ensure `ORDER BY match_date ASC` in the repository query makes form slices chronological before slicing the tail.
