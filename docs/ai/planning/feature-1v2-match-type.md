---
phase: planning
title: "1v2 Match Type — Planning"
description: Task breakdown and implementation order for asymmetric 1v2 match support
---

# Project Planning & Task Breakdown

## Milestones

- [ ] Milestone 1: Backend validation & scoring correct
- [ ] Milestone 2: Frontend form supports 1v2 selection with dynamic slots
- [ ] Milestone 3: End-to-end test: create → delete reverting correct points

## Task Breakdown

### Phase 1: Backend

- [ ] **1.1** Extend `MatchType` validation in `match_service.go`
  - Change single-type check to map-based `teamSizes` lookup
  - Update error message to include `'1v2'`
- [ ] **1.2** Implement asymmetric scoring in `CreateMatch`
  - Extract `calcPointChange(matchType, teamNumber, winnerTeam, base)` helper
  - Handle `1v2` case: solo wins → `+2×base / -1×base`; duo wins → `-2×base / +1×base`
- [ ] **1.3** Update `match_handler.go` error switch to match new error string
- [ ] **1.4** Update `DeleteMatch` — no logic change needed (reverting stored `point_change` is already generic)

### Phase 2: Frontend

- [ ] **2.1** Add `'1v2'` to `MatchType` union in `frontend/src/types/match.ts`
- [ ] **2.2** Update the match type dropdown (radio buttons) to include a "1v2" option with helper text
- [ ] **2.3** Make team slot limits reactive per-team: `team1Limit = 2v2?2:1`, `team2Limit = 1v1?1:2`
- [ ] **2.4** Add client-side validation guard before submission
- [ ] **2.5** `MatchList` filter: add `1v2` option to the type dropdown
- [ ] **2.6** `MatchList` match-type CSS class: replace ternary with object binding that handles all three types

### Phase 3: Polish

- [ ] **3.1** i18n: add `"1v2"` label keys in `vi.json` and `en.json`
- [ ] **3.2** Match history display: verify `1v2` label renders correctly in `MatchesView`

## Dependencies

- Phase 2 can start in parallel with Phase 1 once `MatchType` type is extended
- Phase 3 depends on Phase 2 UI existing

## Timeline & Estimates

| Task | Estimate |
|---|---|
| 1.1 Validation map | 15 min |
| 1.2 Asymmetric scoring | 30 min |
| 1.3 Handler error string | 5 min |
| 2.1–2.4 Frontend form | 45 min |
| 3.1–3.2 i18n + display | 20 min |
| **Total** | **~2 hours** |

## Risks & Mitigation

| Risk | Mitigation |
|---|---|
| Existing tests hardcoded to `"1v1"/"2v2"` error message | Update test assertions when changing the error string |
| Tournament service calls `CreateMatch` and may pass `"1v2"` inadvertently | Tournament service still validates its own match type — no change needed there |

## Resources Needed

- Backend: Go, GORM
- Frontend: Vue 3, TypeScript, Element Plus
