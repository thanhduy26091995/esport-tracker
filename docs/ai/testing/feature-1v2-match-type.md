---
phase: testing
title: "1v2 Match Type — Testing Strategy"
description: Test cases for asymmetric 1v2 match validation and scoring
---

# Testing Strategy

## Scope

- Unit: `calcPointChange` helper, `CreateMatch` validation
- Integration: full match create → score update → delete → score revert

## Test Files

| File | Package | Coverage Target |
|---|---|---|
| `backend/internal/service/match_service_test.go` (new or extend) | service | Validation + scoring |

## Unit Tests — `calcPointChange`

| Case | Input | Expected |
|---|---|---|
| 1v2, solo wins | matchType=1v2, team=1, winner=1, base=1 | +2 |
| 1v2, solo loses | matchType=1v2, team=1, winner=2, base=1 | -2 |
| 1v2, duo member wins | matchType=1v2, team=2, winner=2, base=1 | +1 |
| 1v2, duo member loses | matchType=1v2, team=2, winner=1, base=1 | -1 |
| 1v2, draw | matchType=1v2, team=1, winner=0, base=1 | 0 |
| 1v1, winner | matchType=1v1, team=1, winner=1, base=1 | +1 |
| 2v2, base=2 | matchType=2v2, team=1, winner=1, base=2 | +2 |

## Integration Tests — `CreateMatch` with 1v2

| Case | Expected |
|---|---|
| `match_type=1v2`, team1=[A], team2=[B,C], winner=1 | A: +2, B: -1, C: -1 |
| `match_type=1v2`, team1=[A], team2=[B,C], winner=2 | A: -2, B: +1, C: +1 |
| `match_type=1v2`, team1=[A], team2=[B,C], winner=0 | No score change |
| `match_type=1v2`, team1=[A,B], team2=[C,D] | HTTP 400 INVALID_TEAM_SIZE |
| `match_type=1v2`, team1=[A], team2=[B] | HTTP 400 INVALID_TEAM_SIZE |
| `match_type=3v3` | HTTP 400 INVALID_MATCH_TYPE |
| Delete 1v2 match (solo won) | Scores reverted: A: -2, B: +1, C: +1 |

## Test Data & Environments

- Standard test DB with 3 registered users (A, B, C)
- All with `current_score = 0` at test start
