---
phase: requirements
title: House P&L Dashboard — Requirements
description: Admin cần xem tổng lời/lỗ của nhà cái từ hệ thống cá cược WC2026
---

# Requirements & Problem Understanding

## Problem Statement

Admin hiện tại **không biết nhà cái đang lời hay lỗ** sau khi settle các trận WC2026. Dữ liệu đã có đầy đủ trong bảng `wc_bets` (cột `stake`, `payout`) nhưng chưa được aggregate ở bất kỳ đâu.

Sau mỗi trận settle, admin chỉ thấy `bets_processed` và `total_payout` — không có `total_stake` hay `house_profit`. Không có cái nhìn tổng thể nào về kết quả tài chính của toàn bộ giải.

**Ai bị ảnh hưởng:** Admin quản lý hệ thống cá cược WC2026.  
**Workaround hiện tại:** Không có — admin không thể biết P&L trừ khi tính tay trong DB.

---

## Goals & Objectives

### Primary Goals
- Hiển thị tổng `stake` đã thu, tổng `payout` đã trả, và `house P&L = stake − payout` trong trang admin.
- Breakdown theo từng trận: trận nào nhà cái lời, trận nào lỗ.
- Phân biệt rõ **đã settled** (chắc chắn) và **đang chờ** (pending bets, chưa settled).

### Secondary Goals
- Hiển thị số lượng bets theo loại (handicap vs exact score).
- Filter P&L theo stage (group, r16, qf, sf, final).

### Non-Goals
- Không tính P&L cho hệ thống prediction (điểm) — chỉ tính hệ thống **cá cược tiền (wc_bets)**.
- Không tính wallet top-up admin → user vào P&L (đây là vốn bơm vào, không phải kết quả cược).
- Không cần chart/graph — text + số là đủ trong giai đoạn này.

---

## User Stories & Use Cases

### US-1: Xem tổng P&L toàn giải
> Là admin, tôi muốn thấy ngay trên trang admin một card tóm tắt: tổng stake thu được, tổng payout đã trả, và house profit/loss tính bằng điểm.

**Acceptance criteria:**
- Card "House P&L" hiển thị ở đầu trang admin (hoặc trong tab tài chính).
- Hiển thị: `Tổng stake`, `Tổng payout`, `Lời/Lỗ` (= stake − payout), số bets đã settled.
- `Lời/Lỗ` màu xanh nếu dương, đỏ nếu âm.
- Có phần "Pending" riêng: tổng stake của các bets chưa settled (result IS NULL).

### US-2: Xem P&L theo từng trận
> Là admin, tôi muốn xem breakdown P&L theo từng trận để biết trận nào nhà cái lời/lỗ nhiều nhất.

**Acceptance criteria:**
- Bảng liệt kê các trận đã settled, mỗi dòng có: tên trận, tổng stake, tổng payout, profit/loss.
- Sort mặc định: lỗ nhiều nhất lên đầu (để admin chú ý).
- Các trận chưa settle hiển thị riêng với cột "Stake đang chờ".

### US-3: P&L tự cập nhật sau khi settle trận
> Là admin, khi tôi vừa settle xong một trận, tôi muốn P&L dashboard tự refresh để phản ánh kết quả mới.

**Acceptance criteria:**
- Sau khi gọi settle match API thành công, P&L card reload dữ liệu mới.
- Không cần refresh cả trang.

---

## Success Criteria

- Admin thấy P&L tổng (settled) và pending chỉ trong 1 lần load trang admin.
- Số liệu chính xác khớp với: `SELECT SUM(stake), SUM(payout) FROM wc_bets WHERE result IS NOT NULL`.
- Response time endpoint P&L < 500ms.

---

## Constraints & Assumptions

| Constraint | Chi tiết |
|---|---|
| Đơn vị | P&L tính bằng **điểm** (`NUMERIC(10,2)` sau khi áp dụng feature betting-refinements). VND = điểm × point_rate của settlement |
| Dữ liệu | Chỉ tính `wc_bets` — không tính prediction points |
| Pending bets | `result IS NULL` = chưa settled. Stake đã bị trừ khỏi ví user, nhưng payout chưa xác định |
| Void bets | `result = 'void'` thường có payout = stake (hoàn tiền). Cần include trong P&L |
| Auth | Chỉ admin mới gọi được endpoint này |

---

## Questions & Open Items

1. **[Low]** Có cần filter P&L theo date range (ví dụ: chỉ xem group stage) không, hay tổng toàn giải là đủ?
2. **[Low]** Có cần export P&L ra CSV không?
