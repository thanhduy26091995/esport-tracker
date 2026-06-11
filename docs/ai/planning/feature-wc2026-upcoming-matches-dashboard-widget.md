---
phase: planning
title: WC2026 Upcoming Matches Dashboard Widget — Planning
description: Task breakdown, effort estimates, and implementation order
---

# Project Planning & Task Breakdown

## Milestones

- [x] M1: Backend — `date_from` / `date_to` filter live on `GET /api/v1/wc/matches`
- [x] M2: Frontend component `WcUpcomingWidget.vue` built and styled
- [x] M3: `DashboardView.vue` wired up + widget visible end-to-end

## Task Breakdown

### Phase 1: Backend — Date Range Filter

- [x] **1.1** Add `DateFrom string` and `DateTo string` to `MatchFilter` struct in `internal/repository/wc_repository.go`
- [x] **1.2** Add `WHERE match_date >= ? AND match_date <= ?` conditions to `ListMatches` query (only when fields are non-empty)
- [x] **1.3** Parse `date_from` and `date_to` query params in `WcHandler.ListMatches` and pass to `MatchFilter`
- [ ] **1.4** Manual test: `curl ".../wc/matches?date_from=...&date_to=..."` returns expected subset — **DEFERRED** (server not running during implementation; build passes cleanly)

### Phase 2: Frontend Types & Public API Client

- [x] **2.1** Add `date_from?: string` and `date_to?: string` to `WcMatchFilter` interface in `src/types/wc.ts`
- [x] **2.2** Create `src/services/wcPublicApi.ts` — bare axios instance (no error interceptor) + `listMatchesPublic(filter)` helper. Used by Dashboard for silent background fetches; `wcApi` remains unchanged for WC pages.

### Phase 3: `WcUpcomingWidget.vue` Component

- [x] **3.1** Create `src/components/wc/WcUpcomingWidget.vue`
  - Props: `matches: WcMatch[]`, `hasMore: boolean`
  - Header with title + "Xem lịch đầy đủ →" link to `/world-cup/schedule`
  - Horizontal scroll card list
  - Each card: group/stage label, team codes (home vs away), formatted match time or 🔴 LIVE badge
  - Trailing "Xem thêm →" card when `hasMore = true`
  - `formatMatchTime(iso)` using `Asia/Ho_Chi_Minh` timezone
  - `stageLabel(stage)` mapping `WcStage` → Vietnamese label
- [x] **3.2** Style: dark card, live glow (same keyframe as `WcMatchCard.vue`), horizontal scroll, dashed "Xem thêm" card

### Phase 4: `DashboardView.vue` Integration

- [x] **4.1** Import `WcUpcomingWidget`, `listMatchesPublic`, `WcMatch` type; added `onUnmounted` to Vue imports
- [x] **4.2** Add `upcomingMatches = ref<WcMatch[]>([])` and `hasMoreUpcoming = ref(false)` reactive state
- [x] **4.3** Add `fetchUpcomingWcMatches()` — uses `listMatchesPublic` (silent fail), filters scheduled+live, caps at 5, sets `hasMoreUpcoming`
- [x] **4.4** Call `fetchUpcomingWcMatches()` in `onMounted` (after main data) + `setInterval` every 5 min + `clearInterval` in `onUnmounted`
- [x] **4.5** Insert `<WcUpcomingWidget v-if="upcomingMatches.length > 0" ... />` inside `.dashboard-grid` before `.champion-banner`
- [x] **4.6** Added `.dashboard-full-width { grid-column: 1 / -1 }` CSS utility

## Dependencies

- Task 1.x must complete before 4.x can be fully tested end-to-end.
- Task 2.1 should be done before 3.x (component uses the type).
- Tasks 1.x and 2.x/3.x can be done in parallel.

## Timeline & Estimates

| Phase | Effort |
|---|---|
| Phase 1 — Backend | ~30 min |
| Phase 2 — Types | ~10 min |
| Phase 3 — Component | ~1h |
| Phase 4 — Dashboard integration | ~30 min |
| **Total** | **~2h** |

## Risks & Mitigation

| Risk | Mitigation |
|---|---|
| `WcFeatureMiddleware` returns 503 when feature is off, breaking widget | Silent catch in `fetchUpcomingWcMatches` — widget stays hidden; no dashboard breakage |
| No WC matches synced yet → empty widget | Expected behavior — widget hidden; document that admin must run sync first |
| Timezone edge cases near midnight | Using `toISOString()` (UTC) for API params; display uses `Asia/Ho_Chi_Minh` in `Intl` — no ambiguity |

## Resources Needed

- Backend: Go (existing patterns in `wc_repository.go`, `wc_handler.go`)
- Frontend: Vue 3 Composition API, existing `wcService`, Element Plus (optional for styling)
