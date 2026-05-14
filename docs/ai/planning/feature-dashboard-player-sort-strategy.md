---
phase: planning
title: Dashboard Player Sort Strategy — Planning
description: Task breakdown and implementation order for the player sort strategy feature
---

# Project Planning & Task Breakdown

## Milestones

- [x] M1: Backend endpoint `GET /api/v1/users/payment-ranking` implemented and tested
- [x] M2: Frontend sort logic + Leaderboard refactored to pure display + Dashboard sort control wired
- [x] M3: i18n labels added (mobile polish + unit tests deferred)

## Task Breakdown

### Phase 1: Backend — Payment Ranking Endpoint

- [x] **T1.1** Add `UserWithPaymentTotal` struct to `backend/internal/model/user.go`
- [x] **T1.2** Add `GetPaymentRanking()` method to `backend/internal/repository/user_repository.go` (SQL with LEFT JOIN + GROUP BY + SUM)
- [x] **T1.3** Add `GetPaymentRanking()` method to `backend/internal/service/user_service.go`
- [x] **T1.4** Add `GetPaymentRanking` handler to `backend/internal/api/user_handler.go`
- [x] **T1.5** Register route `GET /users/payment-ranking` in `backend/internal/api/router.go` (must be before `/users/:id`)
- [ ] **T1.6** Manual test: `curl /api/v1/users/payment-ranking` returns users ordered by `total_paid DESC` ← **deferred: requires VPS deploy**

### Phase 2: Frontend — Sort Utility + Leaderboard

- [x] **T2.1** Add `UserWithPaymentTotal` interface to `frontend/src/types/user.ts`
- [x] **T2.2** Add `getPaymentRanking()` to `frontend/src/services/userService.ts`
- [x] **T2.3** Create `frontend/src/utils/sort.ts` with `PlayerSortStrategy` type and `sortByStrategy()` for strategies `default` and `winners-first`
- [x] **T2.4** Refactor `Leaderboard.vue` — remove internal sort, `displayUsers` becomes a pure slice of `props.users`
- [x] **T2.x** Remove `default-sort` from `UserTable.vue` so parent-controlled sort is respected

### Phase 2b: Frontend — Extend userStore

- [x] **T2.5** Add `paymentRankingUsers` ref and `fetchPaymentRanking()` action to `frontend/src/stores/userStore.ts`

### Phase 3: Frontend — Dashboard Integration

- [x] **T3.1** Call `userStore.fetchPaymentRanking()` inside `onMounted` in `DashboardView.vue`
- [x] **T3.2** Add sort state: `sortStrategy` ref (default `'debt-first'`), initialized from `localStorage` key `dashboard-player-sort`, `onSortChange()` handler
- [x] **T3.3** Add `sortedUsersForLeaderboard` computed (debt-first → `paymentRankingUsers`; else → `sortByStrategy()`)
- [x] **T3.4** Add sort control buttons (`el-button-group`, 3 buttons) in leaderboard card header
- [x] **T3.5** Pass `sortedUsersForLeaderboard` to `<Leaderboard>` (no `sortStrategy` prop needed)

### Phase 3b: Frontend — Users Page Integration

- [x] **T3.6** Call `userStore.fetchPaymentRanking()` inside `onMounted` in `UsersView.vue`
- [x] **T3.7** Add sort state to `UsersView.vue` (same pattern, key `users-player-sort`, default `'debt-first'`)
- [x] **T3.8** Add `sortedUsers` computed using same strategy switch
- [x] **T3.9** Pass `sortedUsers` to `<UserTable>` and add sort buttons to Users page header

### Phase 4: i18n + Polish + Tests

- [x] **T4.1** Add i18n keys to `en.json` and `vi.json`
- [x] **T4.2** Translated labels used in sort control buttons (via `sortOptions` computed)
- [ ] **T4.3** Verify mobile layout at ≤ 375 px width — no button overflow ← **deferred: manual browser test**
- [ ] **T4.4** Write unit tests for `sortByStrategy()` ← **deferred**

## Dependencies

```
T1.1 → T1.2 → T1.3 → T1.4 → T1.5   (backend, sequential)
T2.1 → T2.2                            (frontend types/service)
T2.3 → T2.4                            (sort utility before Leaderboard change)
T1.5 + T2.2 → T3.1                     (endpoint must exist before frontend fetch)
T2.3 + T2.4 → T3.3                     (Leaderboard + sort util before computed)
T3.1 + T3.2 + T3.3 → T3.4 → T3.5
T4.1 → T4.2
```

## Timeline & Estimates

| Phase | Estimate |
|---|---|
| Phase 1 (backend endpoint) | ~1.5 hours |
| Phase 2 (frontend types + userStore + Leaderboard) | ~1 hour |
| Phase 3 (Dashboard + Users page integration) | ~1.5 hours |
| Phase 4 (i18n + tests) | ~45 min |
| **Total** | **~5 hours** |

## Risks & Mitigation

| Risk | Likelihood | Mitigation |
|---|---|---|
| Route shadowing — `/users/payment-ranking` matched as `/users/:id` | Medium | Register `/payment-ranking` before `/:id` in router.go |
| `debt_settlements` table name or column name mismatch | Low | Verify against actual migration files before writing SQL |
| Mobile button overflow at 375 px | Medium | Test at target width; abbreviate labels if needed |
| Leaderboard regression — removing internal sort | Low | Only one call site (DashboardView); verify visually after T2.4 |
| localStorage unavailable (private browsing) | Low | try/catch in readSort(); fallback to in-memory |

## Resources Needed

- Backend: Go, GORM, Gin, PostgreSQL
- Frontend: Vue 3 + TypeScript, Element Plus, vue-i18n, Pinia
- No new packages required (frontend or backend)
