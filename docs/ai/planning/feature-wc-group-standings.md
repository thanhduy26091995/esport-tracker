---
phase: planning
title: WC Group Standings — Task Breakdown
description: Implementation task list with effort estimates and dependency ordering
---

# Project Planning & Task Breakdown

## Milestones

- [x] **M1: Backend endpoint live** — `GET /api/v1/wc/standings` returns correct JSON for all groups
- [x] **M2: Frontend component done** — `WcGroupStandings.vue` renders the table correctly
- [x] **M3: Integrated & polished** — standings appear contextually in `WcScheduleView`, responsive, i18n complete

## Task Breakdown

### Phase 1: Backend

- [x] **T1.1** Add models to `backend/internal/model/wc_match.go`
  - Added `WcTeamStanding`, `WcGroupStanding`, `WcStandingsResponse` structs
  - _Effort: 15 min_

- [x] **T1.2** Add `GetGroupStandings()` to `backend/internal/repository/wc_repository.go`
  - Single query: all group-stage matches (any status), ordered by match_date ASC
  - Team roster built from all matches; stats accumulated only for completed matches with non-nil scores
  - Form = last 5 results (chronological, tail of results slice)
  - Also added helper `sortStrings` (insertion sort for small N)
  - _Effort: 45 min_

- [x] **T1.3** Add `GetGroupStandings()` to `backend/internal/service/wc_service.go`
  - Calls repo, sorts each group's teams: points DESC → GD DESC → GF DESC → name ASC
  - Added `sortTeamStandings` and `standingLess` helpers
  - _Effort: 20 min_

- [x] **T1.4** Add handler method to `backend/internal/api/wc_handler.go`
  - `GetGroupStandings` handler added
  - _Effort: 10 min_

- [x] **T1.5** Wire route in `backend/internal/api/router.go`
  - Added `wc.GET("/standings", wcHandler.GetGroupStandings)` alongside `/matches` (bypasses WcFeatureMiddleware)
  - _Effort: 5 min_

### Phase 2: Frontend

- [x] **T2.1** Add TypeScript types to `frontend/src/types/wc.ts`
  - Added `WcTeamStanding`, `WcGroupStanding`, `WcStandingsResponse`
  - _Effort: 10 min_

- [x] **T2.2** Add `getStandings()` to `frontend/src/services/wcPublicApi.ts`
  - Added `getStandings(): Promise<WcStandingsResponse>`
  - _Effort: 10 min_

- [x] **T2.3** Create `frontend/src/components/wc/WcGroupStandings.vue`
  - Props: `standing: WcGroupStanding`
  - Columns: # | Team (flag + code + name) | T | W | D | L | GD | Pts | Form
  - Row highlights: rows 1–2 green, row 3 yellow (WC2026 third-place rule), row 4 none
  - Form badges: W=green, D=grey, L=red
  - `teamCodeToFlag` utility with TLA→alpha2 lookup map (exported for reuse)
  - Mobile: form column hidden below 480px
  - _Effort: 60 min_

- [x] **T2.4** Integrate into `frontend/src/views/WcScheduleView.vue`
  - `standingsData` ref, `showAllGroupsStandings` and `selectedGroupStanding` computed
  - `getStandings()` called in `onMounted` (fail silently)
  - All-groups 2-column grid when no group filter; single-group card when `group_X` filter active
  - Grid CSS with mobile breakpoint (1-column below 640px)
  - _Effort: 30 min_

### Phase 3: i18n & Polish

- [x] **T3.1** Add i18n keys to `frontend/src/locales/vi.json` and `en.json`
  - Added `wc.standings.*` keys: title, rank, team, played, won, drawn, lost, goalDiff, points, form
  - _Effort: 15 min_

- [ ] **T3.2** Manual test on desktop and mobile
  - Verify table layout and responsiveness
  - Verify form badge colours
  - Verify top-2 highlight
  - Verify "no completed matches" state (all zeros)
  - _Effort: 20 min_

## Dependencies

```
T1.1 → T1.2 → T1.3 → T1.4 → T1.5   (backend: sequential)
T2.1 → T2.2 → T2.3 → T2.4           (frontend: sequential)
T1.5 must be done before T2.4 can be tested end-to-end
T3.1 can run in parallel with T2.3
```

## Timeline & Estimates

| Phase | Tasks | Estimated Time |
|-------|-------|---------------|
| Backend | T1.1–T1.5 | ~1.5 hours |
| Frontend | T2.1–T2.4 | ~1.5 hours |
| i18n & polish | T3.1–T3.2 | ~35 min |
| **Total** | | **~3.5 hours** |

## Risks & Mitigation

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| WC2026 group names in DB use a different format than expected | Low | Check `wc_matches.group_name` values in DB before implementing — e.g. `"Group A"` vs `"A"` |
| Team codes (3-letter) don't map to standard ISO 3166-1 alpha-3 for flag emoji | Medium | Build a manual lookup map for the 48 WC2026 teams; fallback to show code text if unknown |
| `wc_matches` has no completed matches yet (pre-tournament) | Low | Handle gracefully: return teams with all-zero stats, still shows team list in alphabetical order |
| Mobile table overflow | Medium | Add `overflow-x: auto` wrapper; reduce columns on small screens |

## Resources Needed

- No new infrastructure (no new DB tables, no new services)
- Flag emoji mapping for 48 WC2026 team codes (frontend lookup table, ~50 lines)
