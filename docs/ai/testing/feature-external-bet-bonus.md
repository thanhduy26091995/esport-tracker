---
phase: testing
title: "External Bet Bonus Points — Testing Strategy"
description: Test cases for bonus point creation, fund deposit, and revert on delete
---

# Testing Strategy

## Scope

- Unit: `CreateBonus` validation, `fund_amount` calculation
- Integration: create bonus → score update → fund deposit → delete → both reverted

## Test Files

| File | Package | Coverage Target |
|---|---|---|
| `backend/internal/service/score_bonus_service_test.go` (new) | service | All cases below |

## Unit / Integration Tests

| Case | Expected |
|---|---|
| Valid: user exists, points=3 | ScoreBonus created; user score +3; no fund deposit created |
| points=0 | HTTP 400, error "points must be positive" |
| points=-1 | HTTP 400, error "points must be positive" |
| user_id not found | HTTP 404 USER_NOT_FOUND |
| Delete bonus (score was +3) | User score -3; bonus record removed |
| Delete non-existent bonus ID | HTTP 404 |
| GET /api/v1/score-bonuses | Returns paginated list of all bonuses |
| GET /api/v1/users/:id/score-bonuses | Returns bonuses for specific user |

## Test Data & Environments

- One registered user with `current_score = 0`
- Fund balance is unaffected by this feature

## Acceptance Criteria (manual)

- [ ] Create bonus via UI → leaderboard score updates immediately
- [ ] `MatchesView` shows `[Cược]` entry with player name, `+Npts`, and description, sorted by date with real matches
- [ ] Fund transactions list is NOT affected (no auto-deposit)
- [ ] Delete bonus from MatchesView → score reverts correctly, entry removed from list
- [ ] Form shows validation error when player is not selected or points ≤ 0
