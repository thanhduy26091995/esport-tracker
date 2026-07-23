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

- [ ] Sets `tournament_type = "world_cup"` in Gin context correctly
- [ ] Sets `tournament_type = "asean_cup"` in Gin context correctly
- [ ] Calls `c.Next()` after setting

### WcConfigRepository

- [ ] `GetConfig("world_cup")` returns WC config row (not ASEAN Cup row)
- [ ] `GetConfig("asean_cup")` returns ASEAN Cup config row (not WC row)
- [ ] `GetConfig("unknown")` returns not-found error

### WcMatchRepository

- [ ] `List(tournamentType="world_cup")` returns only WC matches
- [ ] `List(tournamentType="asean_cup")` returns only ASEAN Cup matches
- [ ] Creating a match with `tournament_type="asean_cup"` persists correctly
- [ ] Cross-tournament contamination: inserting AC match does not appear in WC list query

### WcBetRepository

- [ ] Placed bet inherits `tournament_type` from match's tournament
- [ ] Leaderboard query filters by tournament_type correctly

### WcCustomBetRepository

- [ ] Custom bets created under ASEAN Cup match carry `tournament_type = "asean_cup"`
- [ ] `ListCustomBetsForMatch` returns only bets for the correct tournament match

---

## Integration Tests

- [ ] **Migration regression**: after adding `tournament_type` column, all existing WC rows have `tournament_type = 'world_cup'`
- [ ] **WC API unchanged**: `GET /api/v1/wc/matches` returns same results before and after migration
- [ ] **AC isolation**: `GET /api/v1/ac/matches` returns empty set initially (no ASEAN Cup matches seeded yet)
- [ ] **Config isolation**: `GET /api/v1/ac/config` returns `is_enabled: false`; `GET /api/v1/wc/config` returns WC flag value
- [ ] **Bet isolation**: placing a bet via `/api/v1/ac/matches/:id/bet` stores `tournament_type = "asean_cup"` and does not appear in `/api/v1/wc/` bet queries
- [ ] **Leaderboard isolation**: leaderboard for ASEAN Cup shows only AC bet P&L; WC leaderboard unchanged
- [ ] **House P&L isolation**: admin house P&L dashboard scoped by tournament returns correct totals

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
cd backend && go test ./internal/repository/... ./internal/service/... -v
cd frontend && npm run type-check
```

Manual sign-off checklist:
- [ ] All `/api/v1/wc/*` smoke tests pass post-migration
- [ ] All `/api/v1/ac/*` endpoints return correct data
- [ ] No cross-tournament data leaks observed in browser network tab
- [ ] Admin can settle an ASEAN Cup match without affecting WC data
