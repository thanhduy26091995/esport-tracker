---
phase: planning
title: WC Tournament Analytics — Planning
description: Task breakdown for tournament analytics tab (World Cup 2026) inside WcAnalyticsPanel
---

# Project Planning & Task Breakdown

## Milestones

- [ ] **M1:** Backend — endpoint hoàn chỉnh (DB stats + football-data.org scorers + openfootball + 3-tier cache)
- [ ] **M2:** Frontend — tab "World Cup 2026" hiển thị đầy đủ 11 sections trong WcAnalyticsPanel
- [ ] **M3:** Polish — responsive, i18n, smoke test

## Design decisions (finalized)

- Route: `GET /api/v1/wc/analytics/world-cup-2026`
- Handler/Service: `WcAnalyticsHandler` / `WcAnalyticsService` (extend existing, không dùng WcHandler/WcService)
- Caching: `github.com/patrickmn/go-cache` (new dependency — chưa có trong go.mod)
- Frontend: Tab đầu tiên (default) trong `WcAnalyticsPanel.vue`, content ở `WcTournamentPanel.vue`
- 2 external providers: football-data.org (scorers + assists) + openfootball (goal events, timing, HT stats, team stats, venues)

## Task Breakdown

### Phase 1: Backend

- [ ] **1.1** Add `go-cache` dependency
  - `cd backend && go get github.com/patrickmn/go-cache`
  - Verify in `go.mod` / `go.sum`
  - *Effort: 5min*

- [ ] **1.2** Create `WcFootballDataClient` (`wc_football_data_client.go`)
  - File: `backend/internal/service/wc_football_data_client.go` (new)
  - Struct: `WcFootballDataClient { apiKey string; httpClient *http.Client }`
  - Method: `GetWCScorers(limit int) ([]model.WcScorer, time.Time, error)`
  - HTTP GET `/v4/competitions/WC/scorers?limit=N`, header `X-Auth-Token`
  - Map response → `[]model.WcScorer` (rank, player_name, team_name, team_code, team_crest, goals, assists *int, played_matches)
  - *Effort: 1h*

- [ ] **1.3** Create `WcOpenFootballClient` (`wc_open_football_client.go`)
  - File: `backend/internal/service/wc_open_football_client.go` (new)
  - Struct: `WcOpenFootballClient { httpClient *http.Client }`
  - Method: `GetWCData() (*WcOpenFootballData, error)`
  - Single GET `https://raw.githubusercontent.com/openfootball/worldcup.json/master/2026/worldcup.json`
  - Parse: goal timing buckets, HT stats (comebacks, held_lead, 1st/2nd half goals), own goals, penalty goals, team stats (GF/GA), goals by group, top 5 scoring matches, venue stats
  - `fixMojibake()` helper for accented player names
  - `parseBucket(minuteStr)` for 8 timing buckets
  - *Effort: 2h*

- [ ] **1.4** Add `GetCompletedMatchStats()` to `WcRepository`
  - File: `backend/internal/repository/wc_repository.go`
  - 3 SQL queries: aggregate totals, goals-by-stage (GROUP BY stage ORDER BY MIN(match_date)), highest-scoring match
  - Filter: `status = 'completed' AND home_score IS NOT NULL`
  - *Effort: 45min*

- [ ] **1.5** Add analytics model types to `model/` package
  - File: `backend/internal/model/wc_analytics.go` (new, or append to existing wc model file)
  - Types: `WcAnalyticsResponse`, `WcMatchStats`, `WcStageGoals`, `WcMatchResult`, `WcScorer`, `WcGoalTimingBucket`, `WcHalfTimeStats`, `WcTeamStat`, `WcGroupGoals`, `WcMatchDetail`, `WcVenueStat`
  - *Effort: 30min*

- [ ] **1.6** Extend `WcAnalyticsService` with `GetWorldCup2026Analytics()`
  - File: `backend/internal/service/wc_analytics_service.go`
  - Add fields: `cache *cache.Cache`, `fdClient *WcFootballDataClient`, `ofClient *WcOpenFootballClient`
  - Update `NewWcAnalyticsService()` constructor to accept these new deps
  - Implement 3-tier cache: `wc:analytics:wc2026` (5min) → `wc:analytics:scorers` (30min) → `wc:analytics:openfootball` (15min)
  - Graceful degradation at each external tier
  - *Effort: 1.5h*

- [ ] **1.7** Add handler + route
  - `wc_analytics_handler.go`: add `GetWorldCup2026Analytics(c *gin.Context)` method
  - `router.go`: add `wcAuth.GET("/analytics/world-cup-2026", wcAnalyticsHandler.GetWorldCup2026Analytics)` inside existing `wcAuth` group
  - Wire new deps in `router.go`: create `cache.New(...)`, `NewWcFootballDataClient(...)`, `NewWcOpenFootballClient(...)`, update `NewWcAnalyticsService(...)` call
  - *Effort: 45min*

### Phase 2: Frontend

- [ ] **2.1** Add TypeScript types to `wc.ts`
  - File: `frontend/src/types/wc.ts`
  - Add: `WcAnalyticsResponse`, `WcMatchStats`, `WcStageGoals`, `WcMatchResult`, `WcScorer`, `WcGoalTimingBucket`, `WcHalfTimeStats`, `WcTeamStat`, `WcGroupGoals`, `WcMatchDetail`, `WcVenueStat`
  - *Effort: 20min*

- [ ] **2.2** Add `getWC2026Analytics()` to `wcAnalyticsService.ts`
  - File: `frontend/src/services/wcAnalyticsService.ts`
  - `GET /analytics/world-cup-2026` → `WcAnalyticsResponse`
  - *Effort: 10min*

- [ ] **2.3** Extend `wcAnalyticsStore.ts`
  - Add `wc2026Data: Ref<WcAnalyticsResponse | null>`
  - Add `wc2026Loading: Ref<boolean>` (separate from existing `loading`)
  - Add `loadWC2026Analytics()` action
  - *Effort: 20min*

- [ ] **2.4** Create `WcTournamentPanel.vue`
  - File: `frontend/src/components/wc/analytics/WcTournamentPanel.vue` (new)
  - Props: `data: WcAnalyticsResponse | null`, `loading: boolean`
  - 11 sections (see design doc UI layout):
    1. Stat cards (6): total goals, avg/match, total matches, clean sheets, own goals, penalty goals
    2. Vua phá lưới — top scorers table (team crest, rank, goals, assists)
    3. Kết quả — horizontal bar (home wins / away wins / draws)
    4. Phân tích hiệp — 1st/2nd half goals + comebacks stat
    5. Thời điểm ghi bàn — 8-bucket bar chart (by minute)
    6. Bàn theo vòng — goals by stage bar chart
    7. Bàn theo bảng đấu — goals by group bar chart
    8. Top 5 trận nhiều bàn — match cards
    9. Đội ghi nhiều bàn — sortable team stats table
    10. Sân vận động — venue stats table
    11. Phản lưới & Penalty — 2 mini stat cards
  - Loading skeleton + empty states per section
  - *Effort: 4h*

- [ ] **2.5** Extend `WcAnalyticsPanel.vue`
  - Add "World Cup 2026" tab pane as first tab (before existing tabs)
  - Change default `analyticsTab` from `'my'` → `'wc2026'`
  - Update `onMounted`: load wc2026 (new default)
  - Update `watch`: add `'wc2026'` case, keep existing; old default `'my'` now lazy-loaded
  - Import `WcTournamentPanel`
  - Add i18n key call: `t('wc.analytics.wc2026Tab')`
  - *Effort: 30min*

### Phase 3: Polish

- [ ] **3.1** i18n strings
  - Add `wc.analytics.wc2026Tab` to vi + en locale files
  - All text inside `WcTournamentPanel.vue` through `t()`
  - *Effort: 30min*

- [ ] **3.2** Responsive check
  - Stat cards: 3×2 grid on mobile, 6×1 row on desktop
  - Top scorers table + team stats table: `overflow-x: auto` on mobile
  - Bar charts: full-width on all sizes
  - *Effort: 30min*

- [ ] **3.3** Manual smoke test
  - Verify total goals matches known WC2026 stats
  - Verify top scorer names + goals match football-data.org
  - Verify goal timing distribution looks realistic
  - Tab switching: other tabs still load correctly
  - *Effort: 20min*

## Dependencies

- `github.com/patrickmn/go-cache` — NEW, must `go get` first (task 1.1)
- `FOOTBALL_DATA_API_KEY` — already in `.env`
- No new DB tables or migrations
- Existing `wcAuth` group in router — analytics route slots in here

## Timeline & Estimates

| Phase | Effort |
|-------|--------|
| Phase 1: Backend | ~6.5h |
| Phase 2: Frontend | ~5h |
| Phase 3: Polish | ~1.5h |
| **Total** | **~13h** |
