---
phase: planning
title: Player Highlight Feed — Task Breakdown
description: Phased implementation plan covering all 8 highlight categories
---

# Project Planning & Task Breakdown

## Milestones

- [x] Milestone 1: Backend API live — all 9 repository queries + rule engine returning data
- [x] Milestone 2: Frontend panel rendering all four sections on Dashboard
- [ ] Milestone 3: All 8 highlight categories covered, tested, and displaying correctly

## Task Breakdown

### Phase 1: Backend — Data Layer

- [x] 1.1 Create `internal/model/highlight.go` — `Highlight`, `HighlightsResponse` structs + all type constants
- [x] 1.2 Create `internal/repository/highlight_repository.go` with 9 aggregation queries:
  - `GetCurrentStreaks()` — current win/lose/unbeaten streak per active user
  - `GetLongestStreaks()` — all-time longest win/lose streak per user
  - `GetPointsMovementToday()` — sum of `point_change` (gained + lost) + match count today per user
  - `GetPointsMovementLastHour()` — sum of `point_change` in rolling `NOW() - INTERVAL '1 hour'` window per user
  - `GetRankSnapshot()` — current rank vs yesterday's rank using `yesterday_score = current_score - SUM(today's deltas)` for all users
  - `GetRecentForm()` — last 10 decisive results as `[]bool` per user (win rate derived in service — no separate win-rate query needed)
  - `GetWeeklyActivity()` — match count this week + consecutive active days per user
  - `GetTotals()` — total wins, losses, matches, `current_score` per user (milestones + weekly collapse)
  - `GetStreakBreakers()` — for each match today where the loser had a streak ≥ 3, returns breaker + victim + streak length

### Phase 2: Backend — Service & Handler

- [x] 2.1 Create `internal/service/highlight_service.go` — rule engine covering all 8 categories, running all 9 queries concurrently via `errgroup`:
  - Streak rules (cat. 1): win/lose/unbeaten thresholds → `streak_*` types
  - Rank/point movement (cat. 2): points today, rank change, fastest climber → `points_*`, `rank_*` types
  - Recent form (cat. 3): last-10 win rate → `form_*` types
  - Activity (cat. 4): matches today, marathon, active-day streak → `most_active_today`, `marathon`, `active_streak_days`
  - Hot/cold (cat. 5): winRate + matchesToday threshold → `hot_player`, `cold_player`
  - Fast climb/collapse (cat. 6): last-hour delta → `fast_climb_hour`, `fast_collapse_hour`
  - Milestones (cat. 7): total wins/points/matches thresholds → `milestone_*` types
  - Social/fun (cat. 8): derived from underlying conditions → `social_*` types with random variant selection
  - Assign section (`trending` / `daily_recap` / `competitive` / `social`) and priority per type
  - Cap at 5 highlights per section (20 total)
- [x] 2.2 Create `internal/api/highlight_handler.go` with `GetHighlights` Gin handler
- [x] 2.3 Register `GET /api/v1/highlights` in `router.go`
- [x] 2.4 Smoke test all 8 categories via `curl /api/v1/highlights` with real match data

### Phase 3: Frontend — Types, Service, Store

- [x] 3.1 Create `src/types/highlight.ts` — `Highlight` + `HighlightsResponse` interfaces
- [x] 3.2 Create `src/services/highlightService.ts` — `getHighlights()` API call
- [x] 3.3 Create `src/stores/highlightStore.ts` — Pinia store with sectioned highlights + `fetchHighlights()`

### Phase 4: Frontend — UI Component

- [x] 4.1 Create `src/components/HighlightFeedPanel.vue` + `HighlightList.vue`
- [x] 4.2 Mount `HighlightFeedPanel` in `DashboardView.vue` below the existing content-grid (not inside it)
- [ ] 4.3 Verify responsive layout on 375 px mobile viewport

### Phase 5: Testing & Polish

- [x] 5.1 Unit tests for `HighlightService` — all 8 category rule engines, threshold boundaries, priority scoring, section assignment, random social variant selection
- [ ] 5.2 Repository integration tests for all 9 query methods
- [ ] 5.3 End-to-end smoke: record matches to trigger each highlight type, reload dashboard, verify all appear in correct section
- [ ] 5.4 Verify Vietnamese strings and emoji render on iOS Safari / Android Chrome

## Dependencies

- Phase 2 depends on Phase 1 (service needs repository)
- Phases 3–4 depend on Phase 2 (frontend needs the API)
- Phase 5 backend tests can run in parallel with Phase 4

**External dependencies:** None — all data is in the existing schema.

## Timeline & Estimates

| Phase | Effort estimate |
|---|---|
| Phase 1 — Data layer (9 queries) | 3–4 h |
| Phase 2 — Service (8 categories) + handler | 4–5 h |
| Phase 3 — Frontend types/store | 1 h |
| Phase 4 — UI component (4 sections) | 3–4 h |
| Phase 5 — Tests + polish | 3 h |
| **Total** | **~14–17 h** |

## Risks & Mitigation

| Risk | Likelihood | Mitigation |
|---|---|---|
| Streak / rank queries slow on large history | Low at current volume | Batch all users in one query; add index on `match_participants(user_id)` if needed |
| "1-hour window" fast-climb query returns stale data near hour boundary | Low | Use `NOW() - INTERVAL '1 hour'` in SQL — PostgreSQL handles this correctly |
| Social message variants feel repetitive | Medium | Pool of 2–3 variants per condition with random selection; easy to extend |
| Dashboard layout breaks on mobile with 4-section panel | Medium | Test 375 px viewport before closing Phase 4; use tab layout if space is tight |
| Rule thresholds need tuning after real use | High | All thresholds are named constants in `HighlightService` — change without schema migration |

## Resources Needed

- No new infrastructure, dependencies, or external services
- Existing Go + Vue + PostgreSQL stack is sufficient
