---
phase: planning
title: WC Bot User Flag — Planning
description: Task breakdown for is_bot flag on wc_users
---

# Project Planning & Task Breakdown

## Milestones

- [ ] M1: Backend — model + migration + API
- [ ] M2: Frontend — leaderboard badge + admin toggle
- [ ] M3: Honor banner fix + smoke test

---

## Task Breakdown

### Phase 1: Backend

- [ ] **1.1** Add `IsBot bool` to `WcUser` model (`backend/internal/model/wc_user.go`)
- [ ] **1.2** Add `IsBot bool` to `WcLeaderboardEntry` model (`backend/internal/model/wc_match.go`)
- [ ] **1.3** Add migration SQL: `ALTER TABLE wc_users ADD COLUMN IF NOT EXISTS is_bot BOOLEAN NOT NULL DEFAULT FALSE` (`backend/internal/database/database.go`)
- [ ] **1.4** Add `u.is_bot` to `GetLeaderboard` SELECT in `wc_repository.go`
- [ ] **1.5** Add `SetUserBot(tx *gorm.DB, userID uuid.UUID, isBot bool) error` to `wc_repository.go`
- [ ] **1.6** Add `SetUserBot(userID uuid.UUID, isBot bool) error` to `wc_service.go`
- [ ] **1.7** Add `SetUserBot` handler to `wc_handler.go`
- [ ] **1.8** Wire `wcAdmin.PUT("/users/:wc_user_id/bot", wcHandler.SetUserBot)` in `router.go`

### Phase 2: Frontend

- [ ] **2.1** Add `is_bot?: boolean` to `WcLeaderboardEntry` and `WcUser` in `types/wc.ts`
- [ ] **2.2** Add `setUserBot(userId: string, isBot: boolean)` to `wcService.ts`
- [ ] **2.3** Add `setUserBot` action to `wcStore.ts`
- [ ] **2.4** Show "Bot" badge in `WcLeaderboard.vue` when `entry.is_bot`
- [ ] **2.5** Add bot toggle button in `WcAdminPanel.vue` user list (next to existing block/unblock)
- [ ] **2.6** Fix `WcTop3Banner.vue`: change `top3` computed to `wcStore.leaderboard.filter(e => !e.is_bot).slice(0, 3)`

### Phase 3: Polish & verify

- [ ] **3.1** Verify banner shows correct real users when bots are in positions 1–3
- [ ] **3.2** Verify leaderboard tab shows "Bot" badge
- [ ] **3.3** Verify admin can toggle bot flag on/off
- [ ] **3.4** Verify existing users unaffected (`is_bot = false` default)

---

## Dependencies

- Phase 2 depends on Phase 1 (needs `is_bot` on leaderboard entry response).
- Tasks within each phase are independent and can be done in any order.

---

## Timeline & Estimates

| Phase | Effort |
|---|---|
| Phase 1 (backend) | ~1 hour |
| Phase 2 (frontend) | ~1 hour |
| Phase 3 (verify) | ~15 min |

Total: ~2.5 hours

---

## Risks & Mitigation

| Risk | Mitigation |
|---|---|
| Leaderboard query change breaks existing callers | `is_bot` is additive — existing consumers ignore unknown fields |
| Admin accidentally marks all users as bots | Toggle is reversible; no cascading effects |
| Banner shows 0 entries if all users are bots | `filter().slice(0,3)` returns empty array — banner hides gracefully (`v-if="top3.length >= 1"` already handles this) |
