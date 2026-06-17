---
phase: planning
title: Betting Refinements — Task Breakdown
description: R1 xoá đơn vị tiền tệ (UI), R2 payout số thập phân (DB + backend + frontend)
---

# Project Planning & Task Breakdown

## Milestones

- [x] **M1:** R1 — Zero ₫/VND trong toàn app (trừ fund)
- [x] **M2:** R2 — DB migrations xong, payout float trong backend
- [x] **M3:** R2 — Frontend preview payout chính xác (toFixed 2)

---

## Task Breakdown

### Phase 1: R1 — Remove Currency Symbol (UI only)

- [x] **T1.1** — Thêm `formatNumber(n)` vào `utils/formatters.ts`
  - *Effort: 10min*
- [x] **T1.2** — Thay `formatVND` → `formatNumber` trong WC settlement components
  - Files: `WcSettlementHistory.vue` (dòng 14, 102), `WcSettlementPreview.vue` (dòng 14, 120)
  - *Effort: 15min*
- [x] **T1.3** — Thay `formatVND` → `formatNumber` trong FC25 settlement components
  - Files: `SettlementList.vue`, `SettlementDetails.vue`, `WinnerContributors.vue`, `SettlementTriggerDialog.vue`
  - *Effort: 20min*
- [x] **T1.4** — Thay `formatVND` → `formatNumber` trong shared + views
  - Files: `Leaderboard.vue`, `UserTable.vue`, `TournamentDetailView.vue`, `DashboardView.vue` (dòng 128, 161), `ConfigView.vue`
  - *Effort: 20min*
- [x] **T1.5** — Verify: grep `₫` và `VND` khắp `frontend/src` (trừ fund folder), đảm bảo zero kết quả
  - *Effort: 5min*

### Phase 2: R2 — DB Migrations

- [x] **T2.1** — Viết migration `ALTER TABLE wc_bets ALTER COLUMN payout TYPE NUMERIC(10,2)`
  - *Effort: 15min*
- [x] **T2.2** — Viết migration `ALTER TABLE wc_wallets ALTER COLUMN balance TYPE NUMERIC(10,2)`
  - *Effort: 15min*
- [x] **T2.3** — Kiểm tra `WcWalletLog` (delta, balance_before, balance_after) — đổi sang `NUMERIC(10,2)` nếu cần
  - *Effort: 20min*
- [x] **T2.4** — Chạy migrations trên local DB, verify schema
  - *Effort: 10min*

### Phase 3: R2 — Backend Go

- [x] **T3.1** — Đổi `WcBet.Payout` từ `*int` → `*float64` trong `internal/model/wc_match.go`
  - *Effort: 10min*
- [x] **T3.2** — Đổi `WcWallet.Balance` từ `int` → `float64` trong `internal/model/wc_match.go`
  - *Effort: 10min*
- [x] **T3.3** — Đổi `WcWalletLog.Delta`, `BalanceBefore`, `BalanceAfter` → `float64`
  - *Effort: 10min*
- [x] **T3.4** — Cập nhật `UpdateBetResult(payout int)` → `(payout float64)` trong repository
  - *Effort: 15min*
- [x] **T3.5** — Cập nhật `UpdateWalletBalance(delta int)` → `(delta float64)` trong repository
  - *Effort: 15min*
- [x] **T3.6** — Cập nhật `SettleMatch` trong `wc_service.go`: dùng `math.Round(x*100)/100` thay `math.Floor`
  - *Effort: 30min*
- [x] **T3.7** — Build và fix compile errors
  - *Effort: 20min*

### Phase 4: R2 — Frontend

- [x] **T4.1** — `WcBetForm.vue`: thay 5 chỗ `Math.floor(...)` → `+(x).toFixed(2)` cho handicap + quarter handicap payout
  - *Effort: 20min*
- [x] **T4.2** — `WcPredictionForm.vue`: thay 3 chỗ `Math.floor(...)` → `+(x).toFixed(2)`
  - *Effort: 15min*
- [x] **T4.3** — Verify: preview bet form hiển thị `+1.80` cho stake=2, odds=1.9
  - *Effort: 10min*

---

## Dependencies

```
T1.1 → T1.2 → T1.3 → T1.4 → T1.5   (R1, độc lập với R2)

T2.1 → T2.2 → T2.3 → T2.4           (migrations)
T2.4 → T3.1 → T3.2 → T3.3
T3.3 → T3.4 → T3.5 → T3.6 → T3.7   (backend)
T3.7 → T4.1 → T4.2 → T4.3           (frontend)
```

---

## Timeline & Estimates

| Phase | Effort |
|---|---|
| Phase 1 — R1 Remove VND | ~1h |
| Phase 2 — DB Migrations | ~1h |
| Phase 3 — Backend Go | ~2h |
| Phase 4 — Frontend | ~45min |
| **Total** | **~5h** |

---

## Risks & Mitigation

| Risk | Impact | Mitigation |
|---|---|---|
| Compile errors lan rộng khi đổi int→float64 | Medium | Fix từng file theo thứ tự model → repo → service → handler |
| `WcWalletLog` có nhiều chỗ dùng int casting | Low | Grep toàn bộ codebase tìm `WalletLog` references |
| Settlement VND bị lẻ sau float | Low | Làm tròn ở bước cuối: `math.Round(balance * rate)` |
