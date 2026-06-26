---
phase: planning
title: WC Prediction Analytics — Project Planning
description: Task breakdown, implementation order, and effort estimates
feature: analysis-trending-bet
status: draft
---

# Project Planning: WC Prediction Analytics

## Milestones

- [ ] **M1 — Backend API complete:** All 3 analytics endpoints returning correct data
- [ ] **M2 — Frontend MVP (My Analytics):** Personal stats page rendering with charts
- [ ] **M3 — Community + Compare:** Community trending and comparison table live
- [ ] **M4 — Polish + i18n:** All strings localized, empty states, mobile layout verified

---

## Task Breakdown

### Phase 1: Backend

#### 1.1 Analytics Repository
- [ ] Create `internal/wc/wc_analytics_repository.go`
- [ ] Implement `GetMyAccuracyTimeline(userID, period)` — grouped by day query
- [ ] Implement `GetMyBetStats(userID)` — totals, type distribution, favorite teams, favorite scorelines, home bias
- [ ] Implement `GetMyStreak(userID)` — walk settled bets DESC to compute current + longest streaks
- [ ] Implement `GetCommunityStats()` — distribution, trending teams, trending scorelines, total bets, active users
- [ ] Implement `GetTopPredictors(minBets int)` — ranked by accuracy, min 3 settled bets
- [ ] Implement `GetCompareStats(userID)` — returns both my stats and community averages in one pass

#### 1.2 Analytics Service
- [ ] Create `internal/wc/wc_analytics_service.go`
- [ ] Implement `ClassifyPredictionProfile(stats)` — applies profile heuristics, returns label string
- [ ] Implement `BuildMyResponse(userID, period)` — calls repository, classifies profile, returns DTO
- [ ] Implement `BuildCommunityResponse()` — calls repository, returns DTO
- [ ] Implement `BuildCompareResponse(userID)` — calls repository, returns DTO

#### 1.3 Analytics Handler
- [ ] Create `internal/wc/wc_analytics_handler.go`
- [ ] `GET /wc/analytics/my` handler — parse `period` query param, call service, return JSON
- [ ] `GET /wc/analytics/community` handler — call service, return JSON
- [ ] `GET /wc/analytics/compare` handler — call service, return JSON

#### 1.4 Router Wiring
- [ ] Add 3 routes to `internal/api/router.go` under WC auth middleware group
- [ ] Verify endpoints respond correctly with Postman/curl

---

### Phase 2: Frontend Infrastructure

#### 2.1 Dependencies
- [ ] Install `chart.js@^4` and `vue-chartjs@^5` (`npm install chart.js vue-chartjs`)
- [ ] Verify TypeScript types are included (both packages ship their own types)

#### 2.2 API Service
- [ ] Create `src/services/wcAnalyticsApi.ts`
- [ ] `fetchMyAnalytics(period: string)` → `GET /wc/analytics/my?period=...`
- [ ] `fetchCommunityAnalytics()` → `GET /wc/analytics/community`
- [ ] `fetchCompareAnalytics()` → `GET /wc/analytics/compare`

#### 2.3 Pinia Store
- [ ] Create `src/stores/wcAnalyticsStore.ts`
- [ ] State: `myData`, `communityData`, `compareData`, `loading`, `error`
- [ ] Actions: `loadMyAnalytics(period)`, `loadCommunityAnalytics()`, `loadCompareAnalytics()`

#### 2.4 Tab wiring in WcPredictView
- [ ] Add `<!-- ANALYTICS TAB -->` `el-tab-pane` to `WcPredictView.vue` between Leaderboard and Vô địch tabs (name=`"analytics"`, label=`"📊 Analytics"`)
- [ ] Add `v-if="activeTab === 'analytics'"` lazy render wrapping `<WcAnalyticsPanel />`
- [ ] Add analytics data fetch to `watch(activeTab)` in WcPredictView: `if (tab === 'analytics') await analyticsStore.loadAll()`
- [ ] Import `WcAnalyticsPanel` and `useWcAnalyticsStore` in WcPredictView

---

### Phase 3: Frontend Components

#### 3.1 Chart Wrappers
- [ ] `src/components/wc/analytics/AccuracyTimelineChart.vue` — Line chart (wins/losses over time)
- [ ] `src/components/wc/analytics/BetTypeChart.vue` — Doughnut (handicap / exact / O/U split)
- [ ] `src/components/wc/analytics/TrendingTeamsChart.vue` — Horizontal bar chart
- [ ] `src/components/wc/analytics/PredictionDistributionChart.vue` — Doughnut (home/away/other)

#### 3.2 Panel Components
- [ ] `src/components/wc/analytics/MyAnalyticsPanel.vue`
  - Profile label + description card
  - Accuracy % + streak badges
  - AccuracyTimelineChart with period filter (7d/30d/all)
  - BetTypeChart
  - Favorite teams list (top 5)
  - Favorite scorelines list (top 5)
  - Home/Away bias progress bar
- [ ] `src/components/wc/analytics/CommunityPanel.vue`
  - Total bets + active users stats cards
  - TrendingTeamsChart (top 10)
  - PredictionDistributionChart
  - Trending scorelines list
  - Top Predictors table (name, accuracy, bets count)
- [ ] `src/components/wc/analytics/ComparePanel.vue`
  - 5-row comparison table: Accuracy, Home Bias, Avg Goals, Exact Score Rate, Underdog Rate
  - Highlight rows where user is above/below community average

#### 3.3 Root Panel
- [ ] `src/components/wc/analytics/WcAnalyticsPanel.vue`
  - Element Plus sub-tab navigation (My Analytics / Community / Compare)
  - Loading skeleton on sub-tab switch
  - Error state with retry button
  - Lazy-load each panel's data when sub-tab is first activated

---

### Phase 4: i18n + Polish

#### 4.1 Localization
- [ ] Add all `analytics.*` keys to `src/locales/vi.json`
  - Profile labels: `analytics.profile.goalHunter`, etc.
  - Chart titles, axis labels, table headers, empty state messages
- [ ] Add all `analytics.*` keys to `src/locales/en.json`

#### 4.2 Empty States
- [ ] My Analytics: 0 bets → "Place your first bet to see your analytics" CTA
- [ ] My Analytics: bets but none settled → "Your bets are pending settlement"
- [ ] Community: no data → "No community data yet"
- [ ] Top Predictors: < 3 players qualify → show partial or "Not enough data"

#### 4.3 Responsive + Visual Polish
- [ ] Charts use `responsive: true, maintainAspectRatio: false` in Chart.js config
- [ ] Cards stack vertically on mobile using Tailwind `flex-col sm:flex-row`
- [ ] Test on 375px (iPhone SE) viewport

---

## Dependencies

- Phase 1 (backend) has no frontend dependency — can develop in parallel with Phase 2
- Phase 3 components depend on Phase 2 (store + service)
- Phase 4 depends on Phase 3

---

## Timeline & Estimates

| Phase | Effort estimate |
|-------|----------------|
| Phase 1: Backend API | 3–4 hours |
| Phase 2: Frontend infra (deps, service, store, route) | 1–2 hours |
| Phase 3: Components + charts | 4–6 hours |
| Phase 4: i18n + polish | 1–2 hours |
| **Total** | **~10–14 hours** |

---

## Risks & Mitigation

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Profile classification thresholds feel arbitrary | Medium | Start with simple 3-label system; can refine with real data |
| Chart.js bundle size increase | Low | Tree-shaking supported; only import needed chart types |
| Underdog detection requires handicap odds context | Medium | Fall back to "user picked away team" as proxy for underdog |
| `selection` field format for handicap differs from expected | Low | Read actual data in dev before writing queries |

---

## Resources Needed

- No new infrastructure (no new DB tables, no new services)
- `chart.js` + `vue-chartjs` npm packages (~70 KB gzip combined)
- Reading: `docs/ai/knowledge/wc-betting-system.md`, `docs/ai/knowledge/frontend-patterns.md`
