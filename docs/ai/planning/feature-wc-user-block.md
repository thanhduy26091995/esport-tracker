---
phase: planning
title: WC Admin Block/Unblock User — Planning
description: Task breakdown for block/unblock user feature
---

# Project Planning & Task Breakdown

## Milestones

- [ ] **M1: DB + Backend** — migration, model, repo, service, handlers, routes
- [ ] **M2: Frontend** — user types, service methods, admin panel UI
- [ ] **M3: Validation** — blocked user cannot place bet end-to-end

## Task Breakdown

### Phase 1: Backend

- [ ] **1.1** Thêm `is_blocked bool` vào `WcUser` struct trong `internal/model/wc_user.go`

- [ ] **1.2** DB migration: thêm column `is_blocked BOOLEAN NOT NULL DEFAULT FALSE` vào `wc_users`
  - Cách 1: Thêm vào `AutoMigrate` trong `database.go` (GORM tự ADD COLUMN)
  - Cách 2: Migration file nếu project dùng versioned migrations

- [ ] **1.3** Thêm `SetBlocked(userID uuid.UUID, blocked bool) error` vào `wc_user_repository.go`

- [ ] **1.3b** Thêm `ListPendingBetsForUser(tx, userID)` và `VoidBet(tx, betID)` vào `wc_repository.go`

- [ ] **1.4** Thêm `BlockUser(adminID, targetID uuid.UUID) (int, error)` và `UnblockUser(targetID uuid.UUID) error` vào `wc_service.go`
  - `BlockUser` = DB transaction: void pending bets + refund wallet + set is_blocked=true
  - `BlockUser` guard: adminID == targetID → return error

- [ ] **1.5** Sửa `WcService.PlaceBet()`: fetch user → check `IsBlocked` → return `fmt.Errorf("user is blocked from placing bets")`

- [ ] **1.5b** Sửa `WcService.PlacePrediction()`: fetch user → check `IsBlocked` → return `fmt.Errorf("user is blocked from placing predictions")`

- [ ] **1.6** Sửa `WcHandler.PlaceBet()` và `WcHandler.PlacePrediction()`: map blocked error → HTTP 403

- [ ] **1.7** Thêm `BlockUser` và `UnblockUser` handlers vào `wc_handler.go`

- [ ] **1.8** Đăng ký routes vào `router.go`:
  ```go
  adminGroup.PUT("/users/:id/block", wcHandler.BlockUser)
  adminGroup.PUT("/users/:id/unblock", wcHandler.UnblockUser)
  ```

### Phase 2: Frontend

- [ ] **2.1** Thêm `is_blocked: boolean` vào `WcUser` interface trong `types/wc.ts`

- [ ] **2.2** Thêm `blockUser(userId: string)` và `unblockUser(userId: string)` vào `wcService.ts`

- [ ] **2.3** Sửa `WcAdminPanel.vue` — user table:
  - Thêm badge "Bị khóa" (màu đỏ) khi `user.is_blocked`
  - Thêm button "Khóa" / "Mở khóa" (toggle dựa trên `user.is_blocked`)
  - Disable button nếu user.id === currentUser.id
  - Handler `handleToggleBlock(user)`: gọi block/unblock service → `store.fetchAllUsers()`

### Phase 3: Validation

- [ ] **3.1** Test block flow: block user → pending bets voided + wallet refunded → try new bet → expect 403
- [ ] **3.2** Test block flow: blocked user tries predict → expect 403
- [ ] **3.3** Test unblock flow: unblock → try bet → expect success
- [ ] **3.4** Test self-block guard: admin tries to block themselves → expect 400
- [ ] **3.5** Test atomicity: voided_bets count in response matches actual pending bets count

## Dependencies

- Task 1.3 depends on 1.1 (model)
- Task 1.4-1.6 depends on 1.3
- Task 1.7-1.8 depends on 1.4
- Task 2.2 depends on 2.1 (types)
- Task 2.3 depends on 2.2
- Task 3.x depends on all Phase 1+2

## Timeline & Estimates

| Phase | Estimate |
|-------|----------|
| Phase 1 (Backend) | 1.5–2 giờ |
| Phase 2 (Frontend) | 1–1.5 giờ |
| Phase 3 (Validation) | 0.5 giờ |
| **Total** | **3–4 giờ** |

## Risks & Mitigation

| Risk | Mitigation |
|------|-----------|
| GORM AutoMigrate không ADD COLUMN khi model thay đổi | Verify AutoMigrate includes `WcUser{}` — nếu không, chạy raw SQL migration |
| PlaceBet không fetch user hiện tại | Thêm fetch userRepo.GetByID trong service nếu chưa có |
| Frontend `currentUser` không available trong WcAdminPanel | Lấy từ `wcAuthStore.user` để disable self-block button |
