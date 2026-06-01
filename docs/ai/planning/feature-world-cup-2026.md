---
phase: planning
title: World Cup 2026 Tracker & Betting — Task Breakdown
description: Phased implementation plan for the WC2026 tracker + betting feature
---

# Project Planning & Task Breakdown

## Milestones

- [ ] Milestone 1: Backend live — DB migrated, API sync working, match CRUD + settlement logic complete
- [ ] Milestone 2: Schedule page — users can browse all WC matches, scores, and status
- [ ] Milestone 3: Betting live — wallet, bet placement, settlement, leaderboard all working end-to-end

---

## Task Breakdown

### Phase 1: Database & Models

- [x] 1.1 Write migration — all 9 WC tables added to GORM AutoMigrate in `database.go`; seed `wc_config` row `is_enabled=false` in `seedWcConfig()`
- [x] 1.2 Create `internal/model/wc_match.go` — `WcMatch`, `WcScoreOdds`, `WcWallet`, `WcWalletLog`, `WcBet`, `WcSettlement`, `WcSettlementDetail` + all constants + `WcMatchWithOdds`, `WcLeaderboardEntry` response types
- [x] 1.3 Create `internal/model/wc_user.go` — `WcUser`, `WcConfig` structs (standalone, no FK to existing users)
- [x] 1.4 First admin seeded via `WC_ADMIN_NAME` + `WC_ADMIN_PASSWORD` env vars in `seedWcConfig()`
- [x] 1.5 `go build ./...` passes — schema compiles cleanly

### Phase 2: Backend — Repository

- [x] 2.1 Create `internal/repository/wc_repository.go`:
  - `UpsertMatches(matches []WcMatch)` — bulk upsert on `external_id`
  - `ListMatches(filter MatchFilter)` — filter by status/stage/group/date
  - `GetMatch(id)` — single match with `score_odds` slice preloaded
  - `UpdateMatch(id, fields)` — score, status, handicap kèo, locked_at
  - `CreateScoreOdds(matchID, homeScore, awayScore int, odds float64)` — add a scoreline option
  - `UpdateScoreOdds(id, odds float64)` — change odds for an existing scoreline
  - `DeleteScoreOdds(id)` — remove a scoreline option
  - `GetScoreOdds(matchID, homeScore, awayScore int)` — lookup for bet validation
  - `CreateWallet(tx, userID)` — called inside Register transaction; creates wallet with balance=0
  - `GetWallet(userID)` — fetch wallet (guaranteed to exist for any registered user)
  - `UpdateWalletBalance(tx, userID, delta)` — within transaction; no lower bound check
  - `LogWalletChange(tx, userID, adminID, delta, balanceBefore, balanceAfter, note)` — insert into `wc_wallet_logs`; always called alongside `UpdateWalletBalance` for admin top-ups
  - `GetWalletLogs(userID)` — full top-up/deduction history for a user
  - `ListAllWallets()` — all wallets with user name joined (for settlement preview)
  - `ResetAllWallets(tx)` — set all `wc_wallets.balance = 0` within settlement transaction
  - `CreateSettlement(tx, settlement, details)` — insert `wc_settlements` + batch insert `wc_settlement_details` within same transaction as `ResetAllWallets`
  - `ListSettlements()` — all settlement events ordered by `created_at DESC`
  - `GetSettlement(id)` — single settlement with full `wc_settlement_details` slice
  - `UpdateSettlementDetailStatus(id, userID, status, doneNote)` — mark one user done
  - `ListBetsForMatchPublic(matchID)` — all bets on a match joined with user name (no sensitive data); always accessible
  - `CreateBet(tx, bet)` — within transaction; unique check per type: handicap on (user_id, match_id, bet_type, bet_choice), exact_score on (user_id, match_id, predicted_home_score, predicted_away_score)
  - `ListBets(userID)` — bet history with match join
  - `ListBetsForMatch(matchID)` — all bets on one match (for settlement)
  - `UpdateBetResult(tx, betID, result, payout)` — within transaction
  - `GetLeaderboard()` — all registered users with `SUM(payout - stake) WHERE result IS NOT NULL` from settled `wc_bets` only as `net_profit`, sorted DESC; pending bets excluded; covers entire tournament regardless of tất toán giải resets

### Phase 3: Backend — Auth (Register / Login / Middleware)

- [x] 3.1 Create `internal/repository/wc_user_repository.go`:
  - `CreateUser(name, passwordHash)` — insert into `wc_users`; enforce unique name at app level
  - `GetUserByName(name)` — fetch for login check
  - `GetUserByID(id)` — fetch for JWT validation
  - `SetAdminRole(wcUserID, isAdmin bool)` — update `is_admin` flag
- [x] 3.2 Create `internal/service/wc_auth_service.go`
- [x] 3.3 Create `internal/middleware/wc_auth.go` — `WcJWTMiddleware`
- [x] 3.4 Create `internal/middleware/wc_admin.go` — `WcAdminMiddleware`
- [x] 3.4b Create `internal/middleware/wc_feature.go` — `WcFeatureMiddleware`
- [x] 3.5 Create `internal/api/wc_auth_handler.go` — `Register`, `Login`, `ResetPassword` handlers
- [x] 3.6 Add `WC_JWT_SECRET`, `WC_ADMIN_NAME`, `WC_ADMIN_PASSWORD`, `FOOTBALL_DATA_API_KEY` to `.env`

### Phase 4: Backend — Football API Client

- [x] 4.1 Create `internal/service/wc_football_client.go`:
  - Config: `FOOTBALL_DATA_API_KEY` env var, base URL `https://api.football-data.org/v4`
  - `FetchWCMatches()` — `GET /competitions/WC/matches` → parse into `[]WcMatch`
  - Map API response fields: `id` → `external_id`, `homeTeam.tla` → `home_team_code`, `utcDate` → `match_date`, `score.fullTime` → `home_score`/`away_score`, `status` → `status`, `stage` → `stage`, `group` → `group_name`

### Phase 5: Backend — Service & Settlement

- [x] 5.1 Create `internal/service/wc_service.go`:
  - `SyncMatches()` — call client, upsert into DB, set `bets_locked_at = match_date` for each
  - `PlaceBet(userID, req)` — validate: match not locked, no duplicate; deduct wallet + insert bet in one transaction; **no balance check** (negative balance allowed)
  - `SettleMatch(matchID)` — load match score, load all unsettled bets, evaluate each → update bet result + wallet in one transaction; mark `settled_at`; idempotent (re-settle reverses then re-applies)
  - `GetLeaderboard()` — delegate to repo
  - `PreviewSettlement(pointRate)` — read all wallets, compute direction + amount for each user; returns slice without writing to DB
  - `CreateSettlement(adminID, name, pointRate, note)` — snapshot wallets → insert settlement + details → reset all balances to 0; single transaction

- [x] 5.2 Settlement logic helpers (in `wc_service.go`):
  - `evaluateHandicapBet(bet, homeScore, awayScore) (result, payout)` — adjusted score, push detection
  - `evaluateExactScoreBet(bet, homeScore, awayScore) (result, payout)` — exact match → win; any difference → lose

### Phase 6: Backend — Handler & Routes

- [x] 6.1 Create `internal/api/wc_handler.go` — all WC handlers
- [x] 6.2 Register `/api/v1/wc/*` routes in `router.go`:
  - Config-exempt group (no feature check): `GET /admin/config`, `PUT /admin/config` (always accessible)
  - All remaining routes wrapped in `WcFeatureMiddleware` (returns 503 when disabled)
  - Public group (feature check + no auth): `/auth/register`, `/auth/login`, `/auth/reset-password`, `GET /matches*`, `GET /leaderboard`
  - JWT group (`WcFeatureMiddleware` + `WcJWTMiddleware`): `GET /wallet`, `POST /bets`, `GET /bets`
  - Admin group (`WcFeatureMiddleware` + `WcJWTMiddleware` + `WcAdminMiddleware`): all `/admin/*` routes including `/admin/matches/:id/lock`, `/admin/settlements*`
- [ ] 6.3 Smoke test all endpoints via curl (register, login, sync, view matches, place bet, settle, leaderboard)

### Phase 7: Frontend — Types, Service, Store

- [x] 7.1 Create `src/types/wc.ts` — all WC types: `WcMatch`, `WcBet`, `WcWallet`, `WcLeaderboardEntry`, `WcAuthUser`, filter/stage enums
- [x] 7.2 Create `src/stores/wcAuthStore.ts` — Pinia store: `token` (localStorage), `user` (id, name, isAdmin); `login`, `register`, `logout`, `resetPassword` actions
- [x] 7.2b Create `src/services/wcApi.ts` — separate axios instance for WC; attaches `Authorization: Bearer <token>` from localStorage; handles 401/503
- [x] 7.3 Create `src/services/wcAuthService.ts` — `register`, `login`, `resetPassword` calls via main `api` (no auth needed)
- [x] 7.4 Create `src/services/wcService.ts` — all non-auth WC API calls via `wcApi`
- [x] 7.5 Create `src/stores/wcStore.ts` — Pinia store: `matches`, `wallet`, `bets`, `leaderboard`, `allUsers`, `allWallets`, `settlements` + all actions

### Phase 8: Frontend — Auth Pages

- [x] 8.1 Create `src/views/WcLoginView.vue` — login form: name + password; on success store JWT in `wcAuthStore`, redirect to `/world-cup/bet`; includes password reset panel
- [x] 8.2 Create `src/views/WcRegisterView.vue` — register form: free-text name + password + confirm password; on success auto-login + redirect to `/world-cup/bet`
- [x] 8.3 Add route guard to `/world-cup/bet` via `router.beforeEach`: redirect to `/world-cup/login` if no token in localStorage

### Phase 9: Frontend — Schedule Page

- [x] 9.1 Create `src/components/wc/WcGroupFilter.vue` — pill filter bar: All / Group A–L / R32 / R16 / QF / SF / Final
- [x] 9.2 Create `src/components/wc/WcMatchCard.vue` — match row: emoji flags + team name, score (or date/time if upcoming), status badge; handicap info; live pulse animation; locked/settled badges
- [x] 9.3 Create `src/views/WcScheduleView.vue` — full schedule with filter, grouped by date heading (public, no auth needed)
- [x] 9.4 Add `/world-cup` route in `router.ts`; add "World Cup 2026" nav link in sidebar using `Promotion` icon

### Phase 10: Frontend — Betting Page

- [x] 10.1 Create `src/components/wc/WcBetForm.vue` — modal with two tabs:
  - **Kèo chấp tab**: two side buttons (Home / Away), each togglable; shows `team -X.X @ odds` per side, stake input + live payout preview; "Đặt cược" button submits
  - **Tỉ số tab**: grid of score cards (each shows `A:B × odds`), click to select **multiple** cards; each selected card expands with its own stake input + payout preview; "Đặt cược" button submits all selected bets in sequence
- [x] 10.2 Create `src/components/wc/WcBetHistoryList.vue` — list with match name, type, bet choice, stake × odds, result badge, payout
- [x] 10.2b Create `src/components/wc/WcMatchBetList.vue` — match detail panel showing all bets from all users; shows name, bet type, choice/scoreline, stake, result badge
- [x] 10.3 Create `src/components/wc/WcLeaderboard.vue` — rank list: medal/rank, name, wins/total, net_profit (+/- points); gold/silver/bronze medal display
- [x] 10.4 Create `src/views/WcBettingView.vue` — wallet balance header + current user name + logout; tabs: Đặt cược | Cược đang chờ | Lịch sử | Bảng XH | Quản trị (admin only)
- [x] 10.5 Admin panel (`WcAdminPanel.vue`) — user management: list with is_admin badge, promote/demote toggle, wallet balance + top-up dialog (delta + note)
- [x] 10.5b Admin panel — feature flag toggle: on/off switch at top of admin panel calling `PUT /admin/config`
- [x] 10.5c Admin panel — match management: sync button, lock/settle/score-odds buttons per match; score odds dialog with add/update/delete
- [x] 10.7 Create `src/components/wc/WcSettlementPreview.vue` — `point_rate` input; preview table: Tên | Balance | Hướng badge | Số tiền; "Tạo tất toán" → confirm dialog with name + note
- [x] 10.8 Create `src/components/wc/WcSettlementHistory.vue` — list of settlements; click → expand detail table with per-user status + "Đánh dấu đã xong" button
- [x] 10.6 Add `/world-cup/bet` route (with `meta: { requiresWcAuth: true }`)

### Phase 11: Testing & Polish

- [ ] 11.1 Auth tests: register → login → JWT valid; wrong password → 401; duplicate register → 409; unregistered user login → 401
- [ ] 11.2 Admin gate: non-admin JWT on `/admin/settle` → 403; admin JWT → 200
- [ ] 11.3 Unit tests for settlement helpers: `evaluateHandicapBet` (push, half-ball, all outcomes) and `evaluateExactScoreBet` (correct score, off-by-one, 0-0 match)
- [ ] 11.4 Integration test: PlaceBet → wallet deducted; SettleMatch → winner wallet credited, loser unchanged
- [ ] 11.4b PlaceBet with balance = 0 → accepted; PlaceBet with negative balance → accepted
- [ ] 11.5 Verify bet locking: attempt bet on locked match → 422 response
- [ ] 11.6 Verify duplicate bet rejection: same scoreline twice → 409; same handicap side twice → 409; handicap + different exact scoreline on same match → both accepted
- [ ] 11.7 Verify re-settle idempotency: settle twice with same score → wallets unchanged on 2nd call
- [ ] 11.8 Settlement tests: CreateSettlement snapshots correct balances; all wallets reset to 0 after; history queryable; mark-done updates status without changing amounts

---

## Dependencies

- Phase 2 depends on Phase 1 (DB must exist)
- Phase 3 (auth backend) depends on Phase 1
- Phase 4 depends on Phase 1
- Phase 5 depends on Phase 2 + 4
- Phase 6 depends on Phase 5 + 3 (routes need middleware)
- Phase 7–10 depend on Phase 6 (API must be live)
- Phase 11 can start once Phase 5 + 3 complete

**External dependencies:**
- `FOOTBALL_DATA_API_KEY` — register free account at football-data.org
- `WC_JWT_SECRET` — any random 32+ char string; add to `.env` and production secrets

---

## Timeline & Estimates

| Phase | Effort |
|---|---|
| Phase 1 — DB + models | 1 h |
| Phase 2 — Repository | 2–3 h |
| Phase 3 — Auth backend | 1–2 h |
| Phase 4 — Football API client | 1–2 h |
| Phase 5 — Service + settlement | 2–3 h |
| Phase 6 — Handler + routes | 1–2 h |
| Phase 7 — Frontend types/store | 1 h |
| Phase 8 — Auth pages | 1–2 h |
| Phase 9 — Schedule page | 3–4 h |
| Phase 10 — Betting page | 4–5 h |
| Phase 11 — Tests + polish | 2–3 h |
| **Total** | **~19–28 h** |

---

## Risks & Mitigation

| Risk | Likelihood | Mitigation |
|---|---|---|
| football-data.org WC2026 data not ready until tournament starts | Medium | Seed a mock fixture file for dev; real sync when competition goes live |
| API rate limit exceeded during sync | Low | One sync call fetches all ~104 matches in 1–2 pages; no polling loop needed |
| wallet balance race condition on concurrent bets | Low | DB transaction with SELECT FOR UPDATE on wallet row |
| Settlement applied twice (double credit) | Medium | Idempotent settle: re-settle reverses previous payouts before re-applying |
| Handicap push confusion (whole-number handicap) | Low | Documented in settlement helper with unit tests for push case |

---

## Resources Needed

- Free API key from football-data.org (takes ~2 min to register)
- `FOOTBALL_DATA_API_KEY` env var added to `.env.local` and production secrets
- No new infrastructure — same Go + Vue + PostgreSQL stack
