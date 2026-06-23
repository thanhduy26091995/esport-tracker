---
phase: testing
title: WC Custom Bet — Testing Strategy
description: Test scope, key cases, and validation criteria for kèo phụ (custom proposition bets)
---

# Testing Strategy

## Scope

- **Integration tests (Go):** Service layer — create, place, cancel, settle, void. Wallet mutations verified against real DB.
- **Pure unit tests (Go):** Payout arithmetic (math.Round formula).
- **Manual frontend checklist:** Admin create/settle/void flow; player place/cancel/result display.
- **No E2E automated tests** — consistent with project approach.

---

## Test Files

| File | Package | Coverage Target |
|------|---------|----------------|
| `backend/internal/service/wc_custom_bet_service_test.go` | `service` | Service layer happy paths + edge cases + wallet correctness |

---

## Unit / Integration Tests (Go)

### CreateCustomBet
- [x] Happy path: creates bet with 2 options → returns WcCustomBetWithOptions with options populated
- [x] Min options: 1 option → error "cần từ 2 đến 10 lựa chọn"
- [x] Max options: 11 options → error
- [x] Empty label → error "label không được để trống"
- [x] Odds = 0 → error "odds phải > 0"

### PlaceEntry
- [x] Happy path: creates entry with odds_snapshot from option; wallet unchanged (deferred-deduction model)
- [x] Zero-balance user can place bet (no upfront balance check)
- [x] Bet closed → error "kèo đã đóng"
- [x] Option belongs to different bet → error "lựa chọn không hợp lệ"
- [x] Stake below min → error "điểm cược phải từ X đến Y"
- [x] Stake above max → error
- [x] Duplicate entry (same user same bet) → error (UNIQUE constraint); wallet unchanged

### CancelEntry
- [x] Happy path: deletes entry; wallet unchanged (stake was never taken, no refund needed)
- [x] Wrong user → error "không có quyền"
- [x] Entry not pending (already won/lost) → error
- [x] Bet is closed → error "kèo đã đóng, không thể huỷ"

### Settle
- [x] Happy path: winners credited net profit (payout − stake); losers deducted stake at settlement
- [x] Payout formula: math.Round(stake × odds × 100) / 100; net credit = payout − stake
- [x] Bet status becomes "settled", option is_winner = true
- [x] Bet already settled → error "kèo đã tất toán hoặc đã huỷ"
- [x] Winning option belongs to different bet → error
- [x] Floating-point rounding: 3 × 1.8 = 5.4 exact (no drift)

### VoidBet
- [x] Happy path: entry statuses set to "void", bet status = "void"; no wallet changes (stake was never deducted)
- [x] Already void → error "kèo đã huỷ rồi"
- [x] Already settled → error "kèo đã tất toán"

---

## Manual Test Checklist

### Admin flow
- [ ] Admin panel → match row → "Kèo phụ" button → dialog opens with match name
- [ ] Create form: enter title + 2 options with odds → Submit → bet appears in list with status "Đang mở"
- [ ] Add/remove option rows (min 2, max 10)
- [ ] Leave option label blank → show validation warning, no submit
- [ ] "Đóng cược" → status changes to "Đã đóng"
- [ ] "Mở lại" → status changes back to "Đang mở"
- [ ] "Tất toán" → settle dialog opens with option list → select winner → confirm → status "Đã tất toán", winner chip shown
- [ ] "Huỷ kèo" → confirm dialog → status "Đã huỷ"
- [ ] Entry count increments when player places bet

### Player flow
- [ ] Betting page → match row → "Kèo phụ" toggle → custom bet cards appear
- [ ] Open bet with no entry: options selectable, clicking selects/deselects, stake input appears
- [ ] Place bet → card shows "Đã cược X điểm"; wallet balance unchanged at placement
- [ ] "Huỷ" button visible only when bet is open and entry is pending → cancels; wallet unchanged
- [ ] After admin closes bet: "Huỷ" hidden, shows "Đang chờ kết quả"
- [ ] After admin settles: winner option highlighted in green, result badge shows win/loss amount
- [ ] After admin voids: shows "Kèo đã huỷ — đã hoàn tiền"

---

## Execution

```bash
# Run all WC integration tests (requires TEST_DATABASE_URL)
cd backend && TEST_DATABASE_URL="postgres://..." go test ./internal/service/... -run TestWcCustomBet -v

# Run all service tests
cd backend && go test ./internal/service/...
```

---

## Risks & Gaps

- **Frontend not covered by automated tests** — manual checklist above covers UI states
- **Negative balance at settlement** — if a loser's wallet goes below 0 at settle time, `UpdateWalletBalance` still applies; not explicitly guarded
