---
phase: testing
title: WC Champion Prediction — Testing Strategy
description: Test scope, key cases, and validation criteria
---

# Testing Strategy

## Scope

- **Unit/Integration tests (Go):** Service layer logic — place, settle, idempotency
- **Manual testing:** Frontend flow end-to-end
- **No E2E automated tests** (consistent với approach hiện tại của project)

---

## Test Files

| File | Package | Coverage Target |
|------|---------|----------------|
| `wc_champion_service_test.go` | `service` | Service layer happy paths + edge cases |

---

## Unit / Integration Tests (Go)

### Happy paths
- [ ] User đặt prediction thành công khi cửa sổ mở
- [ ] User sửa prediction (gọi lại POST /predict) → upsert đúng
- [ ] Admin settle với winner đúng → user đặt đúng nhận `floor(points × odds_snapshot)` điểm
- [ ] Admin settle với winner → user đặt sai mất `points` điểm
- [ ] Wallet balance được cập nhật đúng sau settle

### Edge cases
- [ ] Đặt khi `is_open = false` → error `"champion predictions are closed"`
- [ ] Đặt points = 6 → error `"points must not exceed 5"`
- [ ] Đặt points = 0 → error `"points must be greater than 0"`
- [ ] Team ID không tồn tại → error `"team not found"`
- [ ] Settle lần 2 (idempotent) → return no-op, không double-count wallet
- [ ] Xóa prediction khi `is_open = false` → error
- [ ] Settle khi chưa có prediction nào → success với `settled_count = 0`

---

## Manual Test Checklist

### User flow
- [ ] Đăng nhập → vào Predict page → thấy section Champion
- [ ] Khi cửa sổ đóng: thấy badge "Đã đóng", không có form
- [ ] Admin mở cửa sổ → user reload → thấy form và bảng đội
- [ ] Chọn đội, nhập điểm → preview payout đúng công thức
- [ ] Submit → card xác nhận hiện ra với thông tin đã đặt
- [ ] Sửa prediction → đội/điểm cập nhật đúng
- [ ] Hủy prediction → form trống lại
- [ ] Admin đóng cửa sổ → user không thể sửa/hủy

### Admin flow
- [ ] Admin panel → section Champion → bảng đội hiển thị với odds
- [ ] Sửa odds 1 đội → save → user thấy odds mới
- [ ] Toggle mở/đóng → status badge thay đổi
- [ ] Nhập winner → confirm → summary hiện (X đúng / Y sai / Z điểm tổng)
- [ ] Sau settle: user thấy kết quả ✅ hoặc ❌, wallet đã cập nhật
- [ ] Settle lần 2 → hiện thông báo "đã settle rồi", không có gì thay đổi

---

## Execution

```bash
# Backend tests
cd backend && go test ./internal/service/... -run TestWcChampion -v

# Type check frontend
cd frontend && npx vue-tsc --noEmit
```

---

## Risks & Gaps

- Settlement là irreversible — không có test cho "undo settle" (không cần, by design)
- Wallet correctness phụ thuộc vào `wc_wallets` existing — assume wallet luôn tồn tại (created at register)
