---
phase: planning
title: Refactored WC2026 — Task Breakdown
description: Ordered task list for the three WC2026 enhancements. Each feature is independent and can be done in parallel or sequentially.
---

# Project Planning & Task Breakdown

## Milestones

- [x] Milestone 1: Backend — player filter endpoint live
- [x] Milestone 2: Frontend — all three features working in dev
- [ ] Milestone 3: QA pass — timezone, live glow, admin route, player filter verified

---

## Task Breakdown

### Feature 1 — Player filter on /matches

- [x] **1.1** `backend/internal/repository/match_repository.go` — added `GetAllFiltered`; `GetAll` now delegates to it.
- [x] **1.2** `backend/internal/service/match_service.go` — `GetAllMatches` updated to accept `playerID *uuid.UUID`.
- [x] **1.3** `backend/internal/api/match_handler.go` — parses `player_id` query param, validates UUID, passes to `GetAllMatches`; bonuses and `buildFeed` unchanged.
- [x] **1.4** `frontend/src/types/api.ts` — added `MatchFilterParams extends PaginationParams`.
- [x] **1.5** `frontend/src/services/matchService.ts` — `getAll` accepts `MatchFilterParams`.
- [x] **1.6** `frontend/src/stores/matchStore.ts` — `fetchMatches` params updated to `MatchFilterParams`.
- [x] **1.7** `frontend/src/views/MatchesView.vue` — player dropdown added; stat cards hidden when filtering; `handlePlayerFilterChange` re-fetches.

### Feature 2 — VN timezone + live glow

- [x] **2.1** `frontend/src/components/wc/WcMatchCard.vue` — `timeZone: 'Asia/Ho_Chi_Minh'` added to both locale computed.
- [x] **2.2** `frontend/src/components/wc/WcMatchCard.vue` — `glow-live` keyframe animation on `.wc-match-card--live`.

### Feature 3 — Dedicated /world-cup/admin page

- [x] **3.1** `frontend/src/views/WcAdminView.vue` — created; wraps `WcAdminPanel` with header + logout.
- [x] **3.2** `frontend/src/router/index.ts` — `/world-cup/admin` route added with `requiresWcAdmin` meta.
- [x] **3.3** `frontend/src/router/index.ts` — `beforeEach` guard reads `wc_user` from localStorage to check `isAdmin`.
- [x] **3.4** `frontend/src/views/WcPredictView.vue` — admin tab pane and `WcAdminPanel` import removed.
- [x] **3.5** `frontend/src/views/WcScheduleView.vue` — Admin link added in header for logged-in admins; Login link for guests.

---

## Dependencies

- Task 1.2 depends on 1.1 (service calls repo).
- Task 1.3 depends on 1.2 (handler calls service).
- Tasks 1.4–1.7 can be done in parallel with 1.1–1.3.
- Feature 2 is fully independent (frontend only, no backend touch).
- Feature 3 tasks 3.1–3.3 can be done together; 3.4 should come after 3.1 is confirmed working to avoid losing admin functionality before the new page is in place.

---

## Timeline & Estimates

| Task group | Effort |
|---|---|
| Feature 1 backend (1.1–1.3) | ~1 h |
| Feature 1 frontend (1.4–1.6) | ~1 h |
| Feature 2 (2.1–2.2) | ~30 min |
| Feature 3 (3.1–3.5) | ~1.5 h |
| **Total** | **~4 h** |

---

## Risks & Mitigation

| Risk | Mitigation |
|---|---|
| `GetAllFiltered` returns duplicate rows if a match has multiple `match_participants` for the same user | Add `.Distinct()` or `GROUP BY matches.id` to the query |
| Removing admin tab from `WcPredictView` breaks admin workflows before new page is live | Implement 3.1–3.3 and verify the new route works before doing 3.4 |
| JWT parse in `beforeEach` fails if token is expired or malformed | Wrap in try/catch; on any error treat as non-admin and redirect |

---

## Resources Needed

- Existing codebase: Go + GORM backend, Vue 3 + Pinia + Element Plus frontend.
- No new env vars, packages, or infrastructure.
