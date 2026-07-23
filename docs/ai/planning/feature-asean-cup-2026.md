---
phase: planning
title: ASEAN Cup 2026 — Project Planning & Task Breakdown
description: Phased implementation plan for multi-tournament architecture and ASEAN Cup frontend
---

# Project Planning & Task Breakdown

## Milestones

- [ ] **M1: Schema migration** — `tournament_type` column added and backfilled; all WC queries unchanged
- [ ] **M2: Backend — multi-tournament API** — `/api/v1/ac/*` routes live; same handlers serve both tournaments
- [ ] **M3: Frontend — ASEAN Cup routes** — `/asean-cup/*` pages functional with full feature parity to WC
- [ ] **M4: Admin tooling** — Admin can manage ASEAN Cup matches, odds, settle bets, run tất toán
- [ ] **M5: Polish & launch** — Feature flag enabled, translations complete, WC read-only confirmed

---

## Task Breakdown

### Phase 1: Database Migration

- [ ] **1.1** Add `tournament_type VARCHAR(20) NOT NULL DEFAULT 'world_cup'` to: `wc_config`, `wc_matches`, `wc_bets`, `wc_predictions`, `wc_custom_bets`, `wc_settlements`, `wc_champion_config`, `wc_champion_teams`, `wc_champion_predictions`, `wc_sync_logs`
- [ ] **1.2** Create DB indexes: `idx_wc_matches_tournament_type`, `idx_wc_bets_tournament_type`, `idx_wc_predictions_tournament_type`, `idx_wc_custom_bets_tournament_type`, `idx_wc_settlements_tournament_type`, `idx_wc_champion_teams_tournament_type`
- [ ] **1.3** Retire `wc_config` id=1 singleton: add `UNIQUE(tournament_type)` constraint; seed `asean_cup` config row (`is_enabled = false`)
- [ ] **1.4** Retire `wc_champion_config` id=1 singleton: add `UNIQUE(tournament_type)` constraint; seed `asean_cup` champion config row (`is_open = false`)
- [ ] **1.5** Update `wc_champion_teams` unique constraint: drop `UNIQUE(name)` → add `UNIQUE(name, tournament_type)`
- [ ] **1.6** Update GORM model structs to include `TournamentType string` field on all affected models
- [ ] **1.7** Verify AutoMigrate does not drop existing WC data (test on local DB first)

### Phase 2: Backend — Middleware & Router

- [ ] **2.1** Create `backend/internal/middleware/tournament.go` — `TournamentMiddleware(tournamentType string) gin.HandlerFunc`
- [ ] **2.2** Update `WcFeatureMiddleware` to read `tournament_type` from Gin context (set by `TournamentMiddleware`) and query `WHERE tournament_type = ?` instead of `WHERE id = 1`
- [ ] **2.3** In `router.go`: add `TournamentMiddleware("world_cup")` to all existing `/wc` route groups — must run before `WcFeatureMiddleware`
- [ ] **2.4** In `router.go`: add new `/ac` route groups (public, feature-gated, auth, admin) with `TournamentMiddleware("asean_cup")` → `WcFeatureMiddleware` → auth middleware chain
- [ ] **2.5** In `router.go`: register `CreateMatch` handler on `acAdmin.POST("/matches", ...)` — new endpoint
- [ ] **2.6** In `router.go`: register ASEAN Cup background cron for match sync (only if odds-api.io confirms coverage)

### Phase 3: Backend — Repository Layer

- [ ] **3.1** `wc_config` repository: change `GetConfig()` + `UpdateConfig()` from `WHERE id = 1` to `WHERE tournament_type = ?`
- [ ] **3.2** `wc_champion_repository.go`: change champion config queries (`GetChampionConfig`, `UpdateChampionConfig`) from `WHERE id = 1` to `WHERE tournament_type = ?`; add `tournamentType` param to champion team queries
- [ ] **3.3** `wc_match` repository: add `tournamentType` param to `List`, `GetByID`, `Create`, `Upsert`; add new `CreateMatch` method
- [ ] **3.4** `wc_bet` repository: add `tournamentType` param to all query methods; set on `CreateBet`
- [ ] **3.5** `wc_prediction` repository: add `tournamentType` param; set on `CreatePrediction`
- [ ] **3.6** `wc_custom_bet` repository: add `tournamentType` param; set on `CreateCustomBet`
- [ ] **3.7** `wc_settlement` repository: add `tournamentType` param
- [ ] **3.8** Leaderboard query: add `tournament_type` filter
- [ ] **3.9** House P&L query: add `tournament_type` filter

### Phase 4: Backend — Service Layer

- [ ] **4.1** `wc_service.go`: add `tournamentType string` param to all public methods that touch tournament-scoped tables
- [ ] **4.2** `wc_custom_bet_service.go`: add `tournamentType` param
- [ ] **4.3** `wc_champion_service.go`: add `tournamentType` param to all methods; add `CreateChampionTeam` with `tournamentType`; add `CreateMatch` service method
- [ ] **4.4** `FinalizeMatch`, `FinalizeAllMatches`: propagate `tournamentType` to bet settlement and custom bet checks
- [ ] **4.5** Background cron sync: pass `"world_cup"` explicitly when upserting WC matches

### Phase 5: Backend — Handler Layer

- [ ] **5.1** All `wc_handler.go` methods: read `c.MustGet("tournament_type").(string)` and pass to service
- [ ] **5.2** `wc_custom_bet_handler.go`: same
- [ ] **5.3** `wc_champion_handler.go`: same
- [ ] **5.4** Admin handlers (config toggle, match management, settle, tất toán): same
- [ ] **5.5** Verify existing `/wc` routes return identical responses to pre-migration (regression check)

### Phase 6: Frontend — Tournament Config & Composable

- [ ] **6.1** Create `frontend/src/config/tournaments.ts` with `TournamentConfig` interface and `TOURNAMENTS` map
- [ ] **6.2** Create `frontend/src/composables/useTournament.ts` — derives tournament from route path
- [ ] **6.3** Refactor `frontend/src/services/wcApi.ts` to `createTournamentApi(apiPrefix, loginRoute, featureName)` factory; export `wcApi` (WC) and `acApi` (AC) instances
- [ ] **6.4** Update `wcService.ts` to use injected api instance from composable; add `createMatch` call

### Phase 7: Frontend — ASEAN Cup Routes & Views

- [ ] **7.1** In `router/index.ts`: add `/asean-cup` route group mirroring `/world-cup` structure
  - `/asean-cup` → redirect to `/asean-cup/schedule`
  - `/asean-cup/schedule`
  - `/asean-cup/predict`
  - `/asean-cup/bet`
  - `/asean-cup/leaderboard`
  - `/asean-cup/champion` (if champion prediction included)
  - `/asean-cup/admin` (admin pages)
- [ ] **7.2** Parameterize router guard: `isTournamentEnabled(apiBase)` — called with WC or AC base URL depending on route; add `requiresAcFeature` route meta
- [ ] **7.3** Pass `tournamentConfig` to all WC views via `provide/inject` or props — views read `apiPrefix` from it
- [ ] **7.4** Confirm `wcAuthStore` token is shared (it is — `wc_token` in localStorage); 401 redirect now goes to correct login route via `acApi` interceptor

### Phase 8: Frontend — Navigation & UI

- [ ] **8.1** Add ASEAN Cup section to `NavSidebar.vue` — collapse/expand group, show entries (Schedule, Predict, Bet, Leaderboard)
- [ ] **8.2** WC nav section: mark as "archived" or add "(Đã kết thúc)" label; no new betting CTAs
- [ ] **8.3** Top-3 honor banner: scope to active tournament (show ASEAN Cup top-3 when AC is active)
- [ ] **8.4** Dashboard widget: update upcoming matches widget to fetch from active tournament (AC when WC flag is off)

### Phase 9: Translations

- [ ] **9.1** Add ASEAN Cup strings to `vi.json` and `en.json`:
  - Tournament name, short name
  - "ASEAN Cup 2026" page titles
  - Any tournament-specific stage names (`Bán kết`, `Chung kết`, `Vòng bảng`)
- [ ] **9.2** Audit all WC-specific hardcoded strings in components — generalize via `tournamentConfig.displayName`

### Phase 10: Admin — Match Management

- [ ] **10.1** New `WcAdminMatchCreateDialog.vue` — create match form (home/away teams + codes, date, group, stage, venue); calls `POST /admin/matches`; shown via "Tạo trận đấu" button in admin panel
- [ ] **10.2** New `WcAdminScoreDialog.vue` — "Enter Score" dialog (home score, away score inputs); calls `PUT /admin/matches/:id` with `{ home_score, away_score, status: "completed" }`; shown on each match card for scheduled/live matches
- [ ] **10.3** Seed ASEAN Cup champion teams via admin panel or seed script (Thailand, Vietnam, Indonesia, Malaysia, Philippines, Singapore + others)
- [ ] **10.4** If odds-api.io confirms ASEAN Cup coverage: add AC-specific cron + `acOddsApiLeague` constant — score/odds sync automatic, manual entry no longer needed
- [ ] **10.5** Verify admin settle, void, tất toán all work correctly for ASEAN Cup matches

---

## Dependencies

- Phase 1 (migration) must complete before all backend phases
- Phase 2 (middleware) must complete before Phase 3–5 (tests depend on tournament context in handler)
- Phase 3–5 (backend) must complete before Phase 6–8 (frontend depends on `/api/v1/ac/*`)
- Phase 6 (composable) before Phase 7 (routes use it)
- Phase 9 (translations) can run in parallel with Phase 7–8

---

## Timeline & Estimates

| Phase | Effort |
|-------|--------|
| Phase 1: DB migration | 2–3h |
| Phase 2: Middleware + router | 1–2h |
| Phase 3: Repository layer | 3–4h |
| Phase 4: Service layer | 2–3h |
| Phase 5: Handler layer | 2–3h |
| Phase 6: Frontend config + composable | 1–2h |
| Phase 7: Routes + views | 3–4h |
| Phase 8: Nav + UI polish | 2–3h |
| Phase 9: Translations | 1h |
| Phase 10: Admin match mgmt | 2–3h |
| **Total** | **~20–28h** |

---

## Risks & Mitigation

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Migration breaks existing WC queries (forgets `tournament_type` filter) | Medium | Regression test all `/wc/*` endpoints after migration; add integration test for WC leaderboard |
| `wc_config` id=1 singleton referenced in more places than expected | Medium | `grep -r "id = 1" backend/` before starting; audit all usages |
| StatsAPI does not support ASEAN Cup — no match sync | High | Plan manual match entry as primary path; StatsAPI sync is a bonus |
| Component prop/provide chain breaks when adding `tournamentConfig` | Low | Use composable pattern to avoid deep prop drilling |
| Admin accidentally settles ASEAN Cup match with WC handler | Low | Handler reads tournament_type from URL middleware, not from request body |

---

## Resources Needed

- PostgreSQL migration script (can use GORM AutoMigrate + manual seed for config row)
- Confirm ASEAN Cup 2026 team list and group stage format before Phase 10
- Verify StatsAPI coverage for ASEAN Cup (admin task, not blocking development)
