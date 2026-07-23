---
phase: testing
title: ASEAN Cup 2026 — Testing Strategy
description: Test plan ensuring migration correctness, API isolation, and frontend parity
---

# Testing Strategy

## Test Coverage Goals

- Unit tests: all new repository and service methods with `tournament_type` param
- Integration: verify WC data isolation — ASEAN Cup queries never return WC rows and vice versa
- Regression: all existing `/api/v1/wc/*` endpoints return identical responses post-migration
- E2E (manual): full betting flow on `/asean-cup` pages in browser

---

## Unit Tests

### TournamentMiddleware

- [x] Sets `tournament_type = "world_cup"` in Gin context correctly — `TestTournamentMiddleware_SetsWorldCup`
- [x] Sets `tournament_type = "asean_cup"` in Gin context correctly — `TestTournamentMiddleware_SetsAseanCup`
- [x] Calls `c.Next()` after setting — `TestTournamentMiddleware_CallsNext`
- [x] `TournamentTypeKey` constant is `"tournament_type"` — `TestTournamentMiddleware_KeyConstant`

> File: `backend/internal/middleware/tournament_test.go`

### WcMatchRepository

- [x] `ListMatches("world_cup")` returns only WC matches — `TestListMatches_TournamentIsolation_WcDoesNotReturnAc`
- [x] `ListMatches("asean_cup")` returns only ASEAN Cup matches — `TestListMatches_TournamentIsolation_AcDoesNotReturnWc`

> File: `backend/internal/repository/wc_repository_test.go`

### WcLeaderboard isolation

- [x] WC leaderboard shows user with WC prediction; AC leaderboard excludes WC-only user — `TestGetLeaderboard_TournamentIsolation`

> File: `backend/internal/repository/wc_repository_test.go`

### WcConfigRepository

- [ ] `GetConfig("world_cup")` returns WC config row (not ASEAN Cup row)
- [ ] `GetConfig("asean_cup")` returns ASEAN Cup config row (not WC row)
- [ ] `GetConfig("unknown")` returns not-found error

> *(low priority — implicitly covered by integration tests that call `PlaceBet` via AC config)*

### isBetLocked / isLocked

- [x] match_date in past → locked, even if status=scheduled — `TestIsBetLocked/scheduled,_match_date_passed_—_locked`
- [x] match_date in future, no lock_time → not locked — `TestIsBetLocked/scheduled,_no_lock_time_—_not_locked`
- [x] live/completed/cancelled → always locked — `TestIsBetLocked`

> File: `backend/internal/service/wc_service_test.go`

---

## Integration Tests

- [x] **Bet isolation**: `ListBets(userID, "world_cup")` returns only WC bets; `ListBets(userID, "asean_cup")` returns only AC bets — `TestTournamentIsolation_ListBets`
- [x] **Prediction isolation**: `ListPredictions(userID, "world_cup")` / `"asean_cup"` returns only bets from the correct tournament — `TestTournamentIsolation_ListPredictions`
- [x] **Bet history isolation**: `GetBetHistory(userID, "world_cup")` / `"asean_cup"` does not cross-contaminate — `TestTournamentIsolation_GetBetHistory`
- [x] **Settlement tournament_type persisted**: `CreateSettlement("asean_cup")` stores `tournament_type = "asean_cup"`; appears in AC list, absent from WC list — `TestCreateSettlement_TournamentTypePersisted`

> File: `backend/internal/service/wc_integration_test.go`

- [ ] **Migration regression**: after adding `tournament_type` column, all existing WC rows have `tournament_type = 'world_cup'` *(manual verification via psql)*
- [ ] **House P&L isolation**: admin house P&L dashboard scoped by tournament returns correct totals *(not yet automated)*

---

## End-to-End Tests (Manual)

- [ ] **WC archive flow**: navigate to `/world-cup/schedule` → see WC matches → no bet button visible (or disabled) → can view leaderboard
- [ ] **ASEAN Cup schedule**: navigate to `/asean-cup/schedule` → see upcoming ASEAN Cup matches
- [ ] **Place handicap bet**: `/asean-cup/bet` → pick a match → place handicap bet → wallet balance shown
- [ ] **Place tài xỉu bet**: same flow with O/U bet type
- [ ] **Kèo phụ (custom bet)**: admin creates kèo phụ on AC match → player places entry → admin settles → wallet updated
- [ ] **Champion prediction**: `/asean-cup/champion` → pick a team → submit
- [ ] **Admin settle match**: admin logs in → settles an ASEAN Cup match → pending bets resolve → leaderboard updates
- [ ] **Admin tất toán**: admin runs tất toán for ASEAN Cup → settlement record created with `tournament_type = "asean_cup"` → WC settlement data unchanged
- [ ] **Nav isolation**: WC section in sidebar shows "(Kết thúc)" marker; ASEAN Cup section is the active one
- [ ] **Wallet**: wallet balance reflects ASEAN Cup bet wins/losses in addition to historical WC data

---

## Test Data

- Seed one `wc_config` row for `asean_cup` (`is_enabled = false` initially)
- Seed 2–3 test ASEAN Cup matches in `wc_matches` with `tournament_type = "asean_cup"`
- Existing WC seed data must remain unchanged post-migration

---

## Test Reporting & Coverage

```bash
cd backend && go test ./internal/middleware/... ./internal/repository/... ./internal/service/... -v
cd frontend && npx vue-tsc -b --noEmit
```

Manual sign-off checklist:
- [ ] All `/api/v1/wc/*` smoke tests pass post-migration
- [ ] All `/api/v1/ac/*` endpoints return correct data
- [ ] No cross-tournament data leaks observed in browser network tab
- [ ] Admin can settle an ASEAN Cup match without affecting WC data
