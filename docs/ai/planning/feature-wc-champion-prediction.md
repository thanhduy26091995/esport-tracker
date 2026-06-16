---
phase: planning
title: WC Champion Prediction — Task Breakdown
description: Implementation tasks, order, and estimates
---

# Project Planning & Task Breakdown

## Milestones

- [ ] M1: Backend — DB migration + core service logic
- [ ] M2: Backend — API handlers + routes
- [ ] M3: Frontend — UI trong WcPredictView
- [ ] M4: Admin panel — quản lý odds, open/close, settle

---

## Task Breakdown

### Phase 1: Database & Migration
- [ ] 1.1 Viết migration SQL tạo 3 bảng: `wc_champion_teams`, `wc_champion_config`, `wc_champion_predictions`
- [ ] 1.2 Viết seed data 12 đội với odds mẫu (theo design doc)
- [ ] 1.3 Đăng ký models GORM trong `database.go` (AutoMigrate)

### Phase 2: Backend Repository
- [ ] 2.1 Tạo `wc_champion_repository.go`:
  - `GetConfig()` — lấy singleton config
  - `UpsertConfig(isOpen bool)` — mở/đóng
  - `GetTeams()` — danh sách đội
  - `UpdateTeamOdds(id, odds)` — sửa odds
  - `GetMyPrediction(wcUserID)` — prediction của 1 user
  - `GetAllPredictions()` — public list
  - `UpsertPrediction(wcUserID, teamID, points, oddsSnapshot)` — tạo hoặc update
  - `DeletePrediction(wcUserID)` — xóa
  - `SettleChampion(winnerTeamID)` — cập nhật result + points_earned cho tất cả, update wallet

### Phase 3: Backend Service
- [ ] 3.1 Tạo `wc_champion_service.go`:
  - `PlaceOrUpdatePrediction(wcUserID, teamID, points)` — validate (is_open, max 5 pts, team exists), lấy odds_snapshot, gọi repo
  - `DeletePrediction(wcUserID)` — validate is_open
  - `SettleChampion(adminID, winnerTeamID)` — idempotent check, gọi repo, trả summary
  - `UpdateConfig(isOpen bool)`
  - `UpdateTeamOdds(teamID, odds)`

### Phase 4: Backend Handlers & Routes
- [ ] 4.1 Tạo `wc_champion_handler.go` với các handlers:
  - `GetTeams`, `GetConfig`, `GetAllPredictions` (public)
  - `GetMyPrediction`, `PlacePredict`, `DeletePredict` (JWT)
  - `UpdateTeamOdds`, `UpdateConfig`, `Settle` (admin)
- [ ] 4.2 Đăng ký routes trong `router.go`

### Phase 5: Frontend — User UI
- [ ] 5.1 Tạo `WcChampionPanel.vue` — section trong `WcPredictView`:
  - Hiển thị trạng thái (Đang mở / Đã đóng / Đã có kết quả)
  - Bảng đội + odds + flag emoji
  - Form chọn đội + input điểm (1–5) + preview payout
  - Nếu đã đặt: hiển thị lựa chọn hiện tại, cho phép sửa/xóa khi còn mở
- [ ] 5.2 Tích hợp `WcChampionPanel.vue` vào `WcPredictView.vue` (tab mới hoặc section)
- [ ] 5.3 Thêm các API calls vào `wcService.ts`

### Phase 6: Frontend — Admin UI
- [ ] 6.1 Thêm section "Champion" vào `WcAdminPanel.vue`:
  - Toggle mở/đóng cửa sổ dự đoán
  - Bảng đội với inline odds editing
  - Button "Công bố Vô địch" → chọn đội → confirm → settle
  - Hiển thị summary sau settle

---

## Dependencies

- Phase 2 depends on Phase 1 (models trước)
- Phase 3 depends on Phase 2 (service dùng repo)
- Phase 4 depends on Phase 3 (handler dùng service)
- Phase 5 & 6 depend on Phase 4 (cần API)
- Phase 6 (admin settle) phải sau Phase 5 (user predict) trong flow thực tế

---

## Timeline & Estimates

| Phase | Effort ước tính |
|-------|----------------|
| 1 — DB migration | ~30 phút |
| 2 — Repository | ~1.5 giờ |
| 3 — Service | ~1 giờ |
| 4 — Handlers + Routes | ~45 phút |
| 5 — Frontend user UI | ~2 giờ |
| 6 — Frontend admin UI | ~1 giờ |
| **Tổng** | **~7 giờ** |

---

## Risks & Mitigation

| Rủi ro | Mitigation |
|--------|-----------|
| Admin settle 2 lần → double points | Check `settled_at IS NOT NULL` → return no-op |
| User bypass max 5 pts qua API | Validate ở service layer |
| Odds thay đổi sau khi user đã đặt | `odds_snapshot` lưu ngay lúc đặt |
| Admin xóa team nhưng có prediction | FK constraint + không cho xóa khi có predictions |

---

## Resources Needed

- Không cần thêm package Go hay npm
- Không cần external API
- Cần: DB access để chạy migration
