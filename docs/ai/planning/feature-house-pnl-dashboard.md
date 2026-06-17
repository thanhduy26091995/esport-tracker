---
phase: planning
title: House P&L Dashboard — Task Breakdown
description: Implementation order, effort estimates, dependencies
---

# Project Planning & Task Breakdown

## Milestones

- [x] **M1:** Backend endpoint trả về P&L data chính xác
- [x] **M2:** Frontend component hiển thị summary + match breakdown
- [x] **M3:** Auto-refresh sau khi settle trận

---

## Task Breakdown

### Phase 1: Backend

- [x] **T1.1** — Thêm `HousePnLResponse` và `HousePnLMatch` struct vào `internal/model/wc_match.go` (hoặc file model riêng)
  - *Effort: 15min*
- [x] **T1.2** — Implement `GetHousePnL()` trong `internal/repository/wc_repository.go` (2 SQL queries)
  - *Effort: 45min*
- [x] **T1.3** — Thêm `GetHousePnL()` method vào `internal/service/wc_service.go`
  - *Effort: 15min*
- [x] **T1.4** — Thêm `GetHousePnL` handler vào `internal/api/wc_handler.go`
  - *Effort: 20min*
- [x] **T1.5** — Wire route `GET /admin/house-pnl` vào `cmd/server/main.go`
  - *Effort: 10min*
- [x] **T1.6** — Test query bằng cách gọi endpoint trực tiếp, verify số liệu với DB
  - *Effort: 30min*

### Phase 2: Frontend

- [x] **T2.1** — Thêm `HousePnLResponse`, `HousePnLMatch` types vào `frontend/src/types/wc.ts`
  - *Effort: 15min*
- [x] **T2.2** — Thêm `getHousePnL()` method vào `frontend/src/services/wcService.ts`
  - *Effort: 15min*
- [x] **T2.3** — Tạo `WcHousePnL.vue`: summary chips + match breakdown table
  - *Effort: 1.5h*
- [x] **T2.4** — Tích hợp `<WcHousePnL />` vào `WcAdminPanel.vue` (đầu panel)
  - *Effort: 20min*

### Phase 3: Auto-refresh

- [x] **T3.1** — Sau khi `SettleMatch` thành công trong `WcAdminPanel.vue`, emit event để `WcHousePnL` gọi lại API
  - *Effort: 20min*

---

## Dependencies

```
T1.1 → T1.2 → T1.3 → T1.4 → T1.5
T1.5 done → T2.1 → T2.2 → T2.3 → T2.4
T2.4 done → T3.1
```

---

## Timeline & Estimates

| Phase | Effort |
|---|---|
| Phase 1 — Backend | ~2h |
| Phase 2 — Frontend | ~2.5h |
| Phase 3 — Auto-refresh | ~20min |
| **Total** | **~5h** |

---

## Risks & Mitigation

| Risk | Impact | Mitigation |
|---|---|---|
| Query trả về NULL nếu chưa có bet nào | Medium | Dùng `COALESCE(..., 0)` trong SQL |
| `payout` NULL cho pending bets | Low | Filter `WHERE result IS NOT NULL` loại bỏ pending khỏi tổng |
| Void bets có payout = stake — nếu count vào profit sẽ sai | Medium | Tách riêng void bets, không cộng vào house_profit |
