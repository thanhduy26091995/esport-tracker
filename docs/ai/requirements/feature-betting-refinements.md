---
phase: requirements
title: Betting Refinements — Requirements
description: Hai cải tiến độc lập: (1) xoá đơn vị tiền tệ khỏi UI, (2) hỗ trợ payout số thập phân
---

# Requirements & Problem Understanding

## Problem Statement

Hai vấn đề hiện tại trong hệ thống cá cược WC2026:

**Vấn đề 1 — Hiển thị đơn vị tiền tệ:**  
Các component WC đang hiển thị "₫" hoặc "VND" kèm theo số tiền (vd: `320.000 ₫`). Việc này tạo ra hiểu nhầm vì trong hệ thống WC, đơn vị thực chất là **điểm** (point), không phải VND. VND chỉ xuất hiện sau settlement với point_rate. Hiển thị "₫" ngay cạnh số điểm là không chính xác về ngữ nghĩa.

**Vấn đề 2 — Payout làm tròn xuống mất tiền người chơi:**  
Toàn bộ tính toán payout đang dùng `Math.floor()` ở frontend và integer ở backend. Ví dụ: đặt 2 điểm, thắng với odds 1.9 → `Math.floor(2 × 1.9) = Math.floor(3.8) = 3` → người chơi nhận 3 thay vì 3.8. Mất 0.2 điểm mỗi bet thắng. Admin muốn payout giữ nguyên số thập phân (3.8) thay vì làm tròn xuống.

**Ai bị ảnh hưởng:**  
- Vấn đề 1: Admin và người chơi xem UI WC.
- Vấn đề 2: Người chơi bị thiệt payout; admin muốn hệ thống cược chính xác hơn.

---

## Goals & Objectives

### Primary Goals
- **[R1]** Xoá ký hiệu tiền tệ (₫, VND) khỏi tất cả các component liên quan đến WC — chỉ hiển thị số thuần.
- **[R2]** Payout cược hỗ trợ số thập phân — không làm tròn xuống khi tính lợi nhuận.

### Non-Goals
- Không thay đổi cách admin nhập `point_rate` trong WC settlement (vẫn là số nguyên).
- **[R1 only]** Không đổi logic tính toán — R1 chỉ thay đổi display layer.
- Không xoá hàm `formatVND` khỏi codebase — fund components vẫn cần.

---

## User Stories & Use Cases

### US-R1: Xoá đơn vị tiền tệ khắp toàn app (trừ fund management)
> Là người dùng, tôi muốn thấy số thuần (`320.000`) thay vì `320.000 ₫` ở mọi nơi không phải fund management, kể cả bảng tính tiền settlement và leaderboard.

**Quy tắc:**
- **GIỮ VND:** Chỉ các component trong `src/components/fund/` và `src/views/FundView.vue`.
- **XOÁ VND:** Tất cả nơi còn lại — kể cả FC25 settlement, leaderboard, user table, config, dashboard, tournament, WC.

**Danh sách đầy đủ — cần XOÁ ₫/VND:**

| File | Dòng | Nội dung hiện tại |
|---|---|---|
| `components/wc/WcSettlementHistory.vue` | 14 | `point_rate?.toLocaleString() + ' ₫/điểm'` → `/điểm` |
| `components/wc/WcSettlementHistory.vue` | 102 | `format(n) + ' ₫'` → `format(n)` |
| `components/wc/WcSettlementPreview.vue` | 14 | `VND / điểm` → `/điểm` |
| `components/wc/WcSettlementPreview.vue` | 120 | `format(n) + ' ₫'` → `format(n)` |
| `components/settlement/SettlementList.vue` | 37, 47, 51, 55, 61 | `formatVND(...)` → `formatNumber(...)` |
| `components/settlement/SettlementDetails.vue` | 20, 32, 37, 42, 60, 74, 75, 76 | `formatVND(...)` → `formatNumber(...)` |
| ~~`components/settlement/FundContributors.vue`~~ | — | **Giữ nguyên** — tiền quỹ thực tế |
| `components/settlement/WinnerContributors.vue` | 25 | `formatVND(...)` → `formatNumber(...)` |
| `components/settlement/SettlementTriggerDialog.vue` | 20, 27, 31, 119 | `formatVND(...)` → `formatNumber(...)` |
| `components/shared/Leaderboard.vue` | 152, 154 | `formatVND(...)` → `formatNumber(...)` |
| `components/user/UserTable.vue` | 65, 70 | `formatVND(...)` → `formatNumber(...)` |
| `views/TournamentDetailView.vue` | 25 | `formatVND(entry_fee)` → `formatNumber(entry_fee)` |
| `views/DashboardView.vue` | 128, 161 | `formatVND(...)` → `formatNumber(...)` |
| `views/ConfigView.vue` | 116, 332 | `formatVND(...)` → `formatNumber(...)` |

**Danh sách GIỮ NGUYÊN (fund management — real money):**

| File | Lý do |
|---|---|
| `components/fund/FundTransactionList.vue` | Giao dịch quỹ thực tế |
| `components/fund/FundForm.vue` | Nhập số dư quỹ |
| `components/settlement/FundContributors.vue` | Tiền quỹ của từng người — tiền thật |
| `views/FundView.vue` | Trang quản lý quỹ |
| `views/DashboardView.vue:43` | Stat card "Số dư quỹ" — vẫn là tiền thật |

**Cách implement:** Thêm helper `formatNumber(n)` vào `utils/formatters.ts`:
```ts
export function formatNumber(n: number): string {
  return new Intl.NumberFormat('vi-VN').format(n)
}
```
Thay thế toàn bộ `formatVND(...)` bằng `formatNumber(...)` tại các dòng trên. **Không xoá** hàm `formatVND` vì fund components vẫn dùng.

**Acceptance criteria:**
- Zero ký tự `₫` và chuỗi `VND` bên ngoài fund components.
- Số vẫn format đúng dấu phân cách ngàn: `320.000` chứ không phải `320000`.
- Fund management (`FundView`, `FundForm`, `FundTransactionList`) không thay đổi gì.

### US-R2: Payout số thập phân
> Là người chơi, khi tôi đặt 2 điểm và thắng với odds 1.9, tôi muốn nhận đúng 1.8 điểm lợi nhuận (payout = 3.8), không bị làm tròn thành 1 điểm.

**Vị trí cần thay đổi — Frontend:**

| File | Dòng | Hiện tại | Sau khi fix |
|---|---|---|---|
| `WcBetForm.vue` | 227 | `Math.floor(stake * odds) - stake` | `+(stake * odds - stake).toFixed(2)` |
| `WcBetForm.vue` | 130 | `Math.floor(stake * odds) - stake` | `+(stake * odds - stake).toFixed(2)` |
| `WcBetForm.vue` | 234 | `Math.floor(s * odds)` | `+(s * odds).toFixed(2)` |
| `WcBetForm.vue` | 235 | `Math.floor(half * odds) + Math.floor(half)` | `+(half * odds + half).toFixed(2)` |
| `WcBetForm.vue` | 236 | `Math.floor(half)` | `+half.toFixed(2)` |
| `WcPredictionForm.vue` | 213, 230, 231 | `Math.floor(...)` | Tương tự |

**Vị trí cần thay đổi — Backend (Go):**
- `WcBet.Payout` đang là `*int` → đổi thành `*float64` (hoặc `numeric(10,2)` trong DB).
- Settlement payout logic trong `wc_service.go` — tính toán cần giữ nguyên float trước khi ghi.
- `WcWallet.Balance` đang là `int` → **cân nhắc** đổi thành `float64` hoặc `numeric(10,2)`.

**Acceptance criteria:**
- Bet 2 điểm, odds 1.9 → preview hiển thị `+1.80`, ví nhận `+3.80` sau khi settle.
- Bet 2 điểm, odds 2.0 → preview hiển thị `+2.00` (không đổi với trường hợp số chẵn).
- Quarter handicap split payout tính đúng với decimal.
- Backend lưu `payout = 3.8` (float) vào DB, không còn truncate.

---

## Success Criteria

- **R1:** Zero `₫`/`VND` trong tất cả WC Vue components sau khi deploy.
- **R2:** Số liệu payout trong DB khớp với `stake × odds` không làm tròn. Test case: stake=2, odds=1.9 → payout=3.8 trong `wc_bets.payout`.

---

## Constraints & Assumptions

| Constraint | Chi tiết |
|---|---|
| DB migration — wc_bets | `wc_bets.payout` đổi từ `INTEGER` sang `NUMERIC(10,2)`. Migration cần viết cẩn thận để không mất data cũ |
| DB migration — wallet | `wc_wallets.balance` đổi từ `INTEGER` sang `NUMERIC(10,2)`. Tất cả logic cộng/trừ balance phải dùng float |
| Settlement VND | Sau R2, điểm trong wallet là float → `điểm × point_rate` vẫn ra VND integer (làm tròn ở bước cuối) |
| Backward compatibility | Bets cũ đã có `payout` integer — cast tự động sang NUMERIC, không mất data |
| Dependency | R2 phải được implement **trước** hoặc **cùng lúc** với feature `house-pnl-dashboard` (dashboard dùng SUM(payout)) |

---

## Questions & Open Items

1. **[Low — R1]** `WcSettlementHistory.vue` và `WcSettlementPreview.vue` — sau khi xoá `₫`, số điểm lớn như `1.250.000` sẽ hiển thị có dấu chấm ngàn không? Cần verify `Intl.NumberFormat('vi-VN')` format đúng.
2. **[Low — R2]** Decimal places hiển thị: `toFixed(2)` luôn hiện 2 chữ số (vd `+1.80`). Nếu kết quả là số chẵn như `+2.00` có nên rút gọn thành `+2` không?
