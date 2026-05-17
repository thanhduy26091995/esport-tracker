---
phase: planning
title: Win Rate & Tier Evaluation — Task Plan
description: Ordered task breakdown for implementing win rate display and auto tier evaluation
---

# Project Planning & Task Breakdown

## Milestones

- [x] Milestone 1: Backend — win rate query + tier recalculation
- [x] Milestone 2: Frontend — type updates + UserTable win rate column
- [x] Milestone 3: UI polish — WinRateBadge component + i18n strings

## Task Breakdown

### Phase 1: Backend — Data Layer

- [x] **1.1** Add `UserWithStats` struct to `backend/internal/model/user.go`
  - Fields: `WinRate float64`, `TotalMatches int`, `WonMatches int`
- [x] **1.2** Update `user_repository.go` — `GetAll` to use LEFT JOIN subquery returning `UserWithStats`
- [x] **1.3** Update `user_repository.go` — `GetByID` to return `UserWithStats`
- [x] **1.4** Update `user_repository.go` — `GetLeaderboard` (if separate from GetAll) to return `UserWithStats`
- [x] **1.5** Add `UpdateTier(id uuid.UUID, tier string) error` to `user_repository.go`
- [x] **1.6** Add index on `match_participants(user_id)` in a new migration if missing

### Phase 2: Backend — Tier Service

- [x] **2.1** Create `backend/internal/service/tier_service.go`
  - `TierService` struct with `userRepo` and `matchRepo` dependencies
  - `RecalculateForUsers(ids []uuid.UUID) error`: fetches win rate per user, evaluates tier with thresholds, calls `UpdateTier`
  - Pure function `EvaluateTier(winRate float64, totalMatches int) string` (testable in isolation)
- [x] **2.2** Wire `TierService` into `MatchService` (inject as dependency)
- [x] **2.3** Call `TierService.RecalculateForUsers(participantIDs)` in `MatchService.CreateMatch` after successful insert
- [x] **2.4** Call `TierService.RecalculateForUsers(participantIDs)` in `MatchService.DeleteMatch` before/after delete

### Phase 3: Backend — Service & Handler Updates

- [x] **3.1** Update `user_service.go` methods (`GetAll`, `GetByID`, `GetLeaderboard`) to return `UserWithStats`
- [x] **3.2** Update `user_handler.go` — no structural change needed (passes through service response)
- [x] **3.3** Verify JSON serialization: `win_rate`, `total_matches`, `won_matches` appear in API responses

### Phase 4: Frontend — Types & Services

- [x] **4.1** Update `frontend/src/types/user.ts` — add `UserWithStats` interface extending `User` with `win_rate`, `total_matches`, `won_matches`
- [x] **4.2** Update `frontend/src/stores/userStore.ts` — use `UserWithStats` where user lists are stored
- [x] **4.3** Update `frontend/src/services/userService.ts` — update return types to `UserWithStats`

### Phase 5: Frontend — UI Components

- [x] **5.1** Create `frontend/src/components/WinRateBadge.vue`
  - Props: `tier: string`, `winRate: number`, `totalMatches: number`
  - Shows tier chip with color + win rate % (or "Unranked" if `totalMatches < 10`)
- [x] **5.2** Add "Win Rate" column to `frontend/src/components/UserTable.vue`
  - Column shows `WinRateBadge` component
  - Column is visible in both UsersView and DashboardView (same component)

### Phase 6: i18n

- [x] **6.1** Add tier label translations to `frontend/src/locales/en.json`
  - Keys: `tier.pro`, `tier.normal`, `tier.noob`, `tier.unranked`
- [x] **6.2** Add same keys to `frontend/src/locales/vi.json`

## Dependencies

```
Phase 1 → Phase 2 → Phase 3  (backend must be complete before frontend types)
Phase 4 → Phase 5 → Phase 6  (types before components, components before i18n)
Phase 3 ∥ Phase 4 can run in parallel once Phase 1 is done
```

## Timeline & Estimates

| Phase | Effort |
|-------|--------|
| Phase 1: Data layer | ~2h |
| Phase 2: Tier service | ~1.5h |
| Phase 3: Service/handler updates | ~1h |
| Phase 4: Frontend types | ~0.5h |
| Phase 5: UI components | ~2h |
| Phase 6: i18n | ~0.5h |
| **Total** | **~7.5h** |

## Risks & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Existing matches have no re-evaluation on deploy | Tiers stale on first run | Provide a one-time `RecalculateAllTiers()` call on startup or a migration script |
| `match_participants` table missing `user_id` index | Slow win rate queries | Task 1.6 adds the index |
| `UserTable` column overflow on small screens | UX regression | Make win rate column hidden on mobile (Element Plus responsive columns) |

## Resources Needed

- Go backend (existing)
- Vue 3 + Element Plus (existing)
- PostgreSQL migration tooling (existing migrations/ directory)
