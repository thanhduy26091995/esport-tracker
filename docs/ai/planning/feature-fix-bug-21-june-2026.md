---
phase: planning
title: Bug Fix Batch — 21 June 2026 — Planning
description: Task breakdown, estimates, dependencies, and implementation order for all 8 bug fixes
---

# Project Planning & Task Breakdown

## Milestones

- [ ] **M1 — Easy wins (front-end only, no DB):** Bugs 1, 2, 3, 5 fixed and verified
- [ ] **M2 — Backend correctness:** Bugs 4, 8 fixed
- [ ] **M3 — Smart cron:** Bug 7 implemented and tested
- [ ] **M4 — Multi-pick champion:** Bug 6 fully implemented (DB + backend + frontend)

---

## Task Breakdown

### Phase 1 — Frontend-only fixes (Bugs 1, 2, 3, 5)

- [ ] **1.1** `WcScheduleView.vue` — Auto-redirect: in `onMounted`, after config is fetched, call `router.push({ name: 'wc-predict' })` if `featureEnabled && wcAuthStore.isLoggedIn`; else push `wc-login` if `featureEnabled && !wcAuthStore.isLoggedIn`
- [ ] **1.2** `WcScheduleView.vue` — Scroll to next match: import `nextTick`; after `selectedFilter.value = computeDefaultFilter(...)`, call `await nextTick()`, then query `.wc-date-group` and call `.scrollIntoView({ behavior: 'smooth', block: 'start' })` on the first group
- [ ] **1.3** `WcPredictView.vue` — Replace `expandedMatchId: ref<string|null>` with `expandedMatchIds: ref<Set<string>>`; add local `matchPredictionsMap: ref<Record<string, WcPrediction[]>>`; update `toggleMatchPredictions` and template `v-if`
- [ ] **1.4** `WcChampionPanel.vue` — Responsive CSS: lower `champion-main` media query breakpoint to `700px`; wrap `<el-table>` in scrollable div; add `420px` breakpoint for `teams-pick-grid`

### Phase 2 — Backend model/service fixes (Bugs 4, 8)

- [ ] **2.1** `backend/internal/model/wc_match.go` — Change `WcSettlementDetailWithUser.Name` JSON tag from `"name"` to `"user_name"`; change `WcSettlementPreviewRow.Name` JSON tag from `"name"` to `"user_name"` (Bug 8)
- [ ] **2.2** `backend/internal/service/wc_service.go` — In `buildPreviewResult`, guard the `HouseSummary` accumulation with `if m.SettledAt == nil` (Bug 4)

### Phase 3 — Smart cron (Bug 7)

- [ ] **3.1** `backend/internal/service/wc_service.go` — Add `GetMatchScheduleSummary()` method: runs two queries (`COUNT(*) WHERE status='live'`; `MIN(match_date) WHERE status='scheduled' AND match_date > NOW()`)
- [ ] **3.2** `backend/internal/cron/wc_sync.go` — Replace fixed `time.NewTicker` with adaptive loop; read `WC_LIVE_SYNC_INTERVAL_MINUTES` (default 5) and `WC_PRE_MATCH_INTERVAL_MINUTES` (default 10) env vars; call `computeNextInterval(svc)` after each sync to determine sleep duration

### Phase 4 — Multi-pick champion (Bug 6)

- [ ] **4.1** DB migration — Drop `UNIQUE` on `wc_champion_predictions(user_id)`; add composite UNIQUE on `(user_id, team_id)`
- [ ] **4.2** `backend/internal/model/wc_champion.go` — Remove `uniqueIndex` GORM tag on `UserID`; add `uniqueIndex:uq_champion_user_team` on `UserID` and `TeamID`; add `ID` field to `WcChampionPredictionResponse` type
- [ ] **4.3** `backend/internal/repository/wc_champion_repository.go` — Rename `GetMyPrediction → GetMyPredictions` (returns slice); rename `CreateOrUpdatePrediction → CreatePrediction` (insert only, let DB enforce composite unique); rename `DeletePrediction(userID) → DeletePredictionByID(predictionID, userID)`
- [ ] **4.4** `backend/internal/service/wc_champion_service.go` — Update `PlaceChampionPrediction`: remove upsert logic, call `CreatePrediction`; update `DeleteChampionPrediction(userID, predictionID)`; update `GetMyChampionPrediction → GetMyChampionPredictions`; update `SettleChampion` to pay out all matching predictions (loop all where `team_id = winnerTeamID`)
- [ ] **4.5** `backend/internal/api/wc_champion_handler.go` — Update `GetMyPrediction` handler to return array; update `PlaceChampionPrediction` to use new service signature; update `DeleteChampionPrediction` to read `:id` from path + pass to service
- [ ] **4.6** `backend/internal/api/router.go` — Change `DELETE /champion/predict` route to `DELETE /champion/predictions/:id`
- [ ] **4.7** `frontend/src/types/wc.ts` — Add `id: string` to `WcChampionPredictionMine`; change `getMyChampionPrediction` return type to `WcChampionPredictionMine[]`
- [ ] **4.8** `frontend/src/services/wcService.ts` — Update `getMyChampionPrediction()` → `getMyChampionPredictions()`; update `deleteChampionPrediction(predictionId: string)` to pass ID in path
- [ ] **4.9** `frontend/src/components/wc/WcChampionPanel.vue` — Replace `myPrediction: ref<...| null>` with `myPredictions: ref<...[]>`; render a list of "my prediction" cards; each card has its own delete button with `predictionId`; update `handleDelete(id)` and `handlePlace()` logic

---

## Dependencies

```
Phase 1 — no dependencies, can start immediately
Phase 2 — independent of Phase 1 and 4; can run in parallel with Phase 1
Phase 3 — depends on Phase 2 (uses WcService); can run after 2.1
Phase 4 tasks — sequential: 4.1 → 4.2 → 4.3 → 4.4 → 4.5 → 4.6 (backend chain) → 4.7 → 4.8 → 4.9 (frontend chain)
```

---

## Timeline & Estimates

| Task | Effort |
|------|--------|
| 1.1 Auto-redirect | 15 min |
| 1.2 Scroll to next match | 20 min |
| 1.3 Per-match collapse | 30 min |
| 1.4 Champion responsive | 20 min |
| 2.1 Settlement JSON tag fix | 5 min |
| 2.2 buildPreviewResult HouseSummary guard | 10 min |
| 3.1 GetMatchScheduleSummary | 20 min |
| 3.2 Adaptive cron | 30 min |
| 4.1 DB migration | 10 min |
| 4.2 Model update | 10 min |
| 4.3 Repository update | 20 min |
| 4.4 Service update | 30 min |
| 4.5 Handler update | 20 min |
| 4.6 Router update | 5 min |
| 4.7 TS types | 10 min |
| 4.8 Service client | 15 min |
| 4.9 Champion panel UI | 45 min |
| **Total** | **~5.5 hours** |

---

## Risks & Mitigation

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Bug 6 DB migration on live data breaks existing single predictions | Low | Migration only drops a constraint and adds a scoped one; existing single rows are valid |
| Bug 7 adaptive cron overshoots StatsAPI rate limit during World Cup group stage (many simultaneous live matches) | Low | 5-min interval means max 12 calls/hour; StatsAPI free tier is typically 100 calls/hour |
| Bug 3 reactivity issue with `Set<string>` in Vue 3 | Medium | Must always assign a new Set (not mutate in place) to trigger `ref` reactivity |
| Bug 1 redirect loops if router guard also checks config | Low | Check guard logic in `router/index.ts` to ensure no ping-pong redirect between wc-schedule and wc-predict |

---

## Implementation Order (Recommended)

1. Start with Phase 2 (bugs 4, 8) — pure backend, low risk, easy wins.
2. Phase 1 (bugs 1–3, 5) — frontend, no server restart needed.
3. Phase 3 (bug 7) — requires backend restart.
4. Phase 4 (bug 6) — requires DB migration + full stack changes; should be done last and tested end-to-end.
