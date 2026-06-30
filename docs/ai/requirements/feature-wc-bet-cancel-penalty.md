---
phase: requirements
title: WC Bet Cancel Penalty — Requirements
description: Admin-controlled penalty deduction when a user cancels a pending bet
---

# Requirements & Problem Understanding

## Problem Statement

Hiện tại người dùng có thể hủy cược (`DELETE /wc/bets/:id`) hoàn toàn miễn phí và không có hậu quả.
Điều này cho phép các hành vi lạm dụng — đặt cược nhiều match, chờ xem tình hình, rồi hủy ngay trước khi kết quả rõ ràng.
Admin cần công cụ để ngăn chặn hành vi này bằng cách thu phí phạt khi hủy cược.

**Ai bị ảnh hưởng:** Admin (muốn kiểm soát), WC users (sẽ bị phạt nếu hủy).

**Tình trạng hiện tại:**
- Hủy cược → bet bị hard-delete khỏi DB → không để lại dấu vết
- Ví (wallet) không thay đổi (vì hệ thống dùng deferred deduction — điểm chỉ bị trừ khi tất toán)
- Không có record nào trong "Lịch sử" cho bet đã hủy

---

## Goals & Objectives

**Primary goals:**

*Cancel penalty:*
1. Admin có thể bật/tắt tính năng phạt hủy cược qua cài đặt (`cancel_penalty_enabled`)
2. Admin có thể cấu hình % phạt hủy (`cancel_penalty_percent`, default 20%)
3. Khi user hủy cược và penalty đang bật: trừ `floor(stake × percent / 100)` điểm trực tiếp từ ví
4. Tạo `wc_wallet_logs` entry cho khoản khấu trừ (audit trail)
5. Bet bị hủy vẫn hiển thị trong tab "Lịch sử" (thay vì bị xóa hoàn toàn)

*Reduce stake penalty:*
6. Admin cấu hình % giảm tối đa được phép (`bet_reduce_max_percent`, ví dụ 50%) — nếu = 0 thì không giới hạn
7. Admin cấu hình % phạt phần giảm vượt quá (`bet_reduce_penalty_percent`, config riêng)
8. Khi user giảm stake vượt quá giới hạn: phạt trên **phần dư thừa** = `floor(excess × bet_reduce_penalty_percent / 100)`, trừ vào ví
9. `original_stake` được lưu khi đặt cược lần đầu — giới hạn tính trên stake gốc, không đổi dù user đã sửa trước đó
10. Tăng stake hoàn toàn tự do, không phạt
11. Popup xác nhận hiển thị số điểm bị phạt trước khi áp dụng
12. Áp dụng cho tất cả loại cược (cược thường + kèo phụ)

**Secondary goals:**
- Cancel: Dialog xác nhận **luôn xuất hiện**, hiển thị số điểm bị phạt (kể cả khi = 0)
- Reduce: Popup chỉ xuất hiện khi reduction vượt quá giới hạn (không hiện khi giảm hợp lệ)

**Non-goals:**
- Không áp dụng penalty cho admin void bets
- Không áp dụng penalty cho bets bị hủy do block user
- Không retroactive với bets đã được hủy trước khi tính năng được bật
- Không thay đổi settlement logic cho các bet đã settle

> **Scope note:** Penalty áp dụng cho **tất cả** loại cược — bao gồm cả kèo phụ (`wc_custom_bet_entries`), không chỉ cược thường (`wc_bets`).

---

## User Stories & Use Cases

**Admin:**
- AS an admin, I want to enable a cancel penalty so that users think twice before cancelling bets
- AS an admin, I want to set the penalty percentage (e.g., 20%) so that it feels fair but discouraging
- AS an admin, I want to see an audit trail in wallet logs when penalties are applied

**WC User:**
- AS a user with a pending bet, I want to know the cancel penalty before I confirm cancellation so I can make an informed decision
- AS a user reducing my stake beyond the limit, I want to see how much I'll be penalized before confirming
- AS a user, I want to see my cancelled bets in history with: tên trận, loại cược, stake, số điểm bị phạt, ngày hủy
- AS a user, I want to see the penalty deduction in my wallet transaction history so I know why my balance changed

**Edge cases:**

*Cancel:*
- User has insufficient wallet balance → still allow cancel, deduct whatever is available (balance → 0, no negative)
- Penalty amount rounds to 0 → cancel still proceeds; dialog still appears and shows "0 điểm bị phạt"; no wallet log created
- Bet is already locked → cannot cancel
- Bet is already settled → cannot cancel

*Reduce stake:*
- `bet_reduce_max_percent = 0` → no limit, all reductions free (no popup)
- User reduces to exactly the limit → free, no popup
- User reduces beyond limit → popup with penalty; penalty capped at available wallet balance (no negative)
- Penalty on excess reduction rounds to 0 → still show popup with "0 điểm bị phạt"; proceed without wallet change
- Increasing stake → always free, no popup
- Custom bet (kèo phụ) → same rules apply

---

## Success Criteria

*Cancel penalty:*
- [ ] Admin toggles `cancel_penalty_enabled` ON/OFF — change takes effect immediately
- [ ] Admin sets `cancel_penalty_percent` (1–100)
- [ ] Cancelling a 10-point bet at 20% penalty → deducts 2 points from wallet
- [ ] `wc_wallet_logs` row: `delta=-2`, `note="Hủy cược - phạt 20%"`, `admin_id=NULL`
- [ ] Cancel dialog appears **only** when `cancel_penalty_enabled = true` AND `penalty > 0`; otherwise cancel proceeds immediately without confirmation
- [ ] Penalty deduction visible in user's own wallet transaction history
- [ ] When penalty OFF: cancel is free, no wallet change, bet still soft-deleted (shows in history)
- [ ] Cancelled bet appears in Lịch sử tab (mixed chronologically): tên trận, loại cược, stake, điểm bị phạt, ngày hủy + `[Đã hủy]` badge
- [ ] Same logic for kèo phụ

*Reduce stake penalty:*
- [ ] Admin sets `bet_reduce_max_percent` (0 = no limit) and `bet_reduce_penalty_percent`
- [ ] `original_stake` stored at bet placement, never changes
- [ ] Reducing within limit → free, no popup
- [ ] Reducing from 10 (original) to 3 at 50% limit, 20% reduce-penalty → excess=2, penalty=`floor(2×20/100)=0`
- [ ] Reducing from 10 to 3 at 50% limit, 50% reduce-penalty → excess=2, penalty=`floor(2×50/100)=1`
- [ ] Popup shows penalty amount before user confirms over-limit reduction
- [ ] Increasing stake → always free, no popup
- [ ] Same logic for kèo phụ

*General:*
- [ ] All existing settlement and void logic unchanged

---

## Constraints & Assumptions

**Technical constraints:**
- Must use `shopspring/decimal` for penalty arithmetic (no `float64` for money ops)
- `wc_wallet_logs.admin_id` is currently `NOT NULL` — needs to become nullable to support user-initiated entries
- Soft-delete approach: add `cancelled_at TIMESTAMP NULL` to `wc_bets` (avoids changing existing query patterns)
- `wc_config` is a singleton row (id = 1) — new fields added via AutoMigrate
- `original_stake` must be stored at bet placement in both `wc_bets` and `wc_custom_bet_entries`; immutable thereafter
- 4 new config fields total: `cancel_penalty_enabled`, `cancel_penalty_percent`, `bet_reduce_max_percent`, `bet_reduce_penalty_percent`

**Business constraints:**
- Penalty cannot push wallet balance below 0 (floor at 0)
- Default penalty percent = 20 (already stated in requirements)
- Feature ships disabled by default (`cancel_penalty_enabled = false`)

**Assumptions:**
- "Lịch sử" tab = tab `history` in `WcBettingView.vue` (currently shows settled bets)
- Cancelled bets should appear alongside settled bets in that tab (with distinct visual treatment)
- The `wc_wallet_logs` note is the audit trail — no separate audit table needed

---

## Questions & Open Items

- [RESOLVED] Penalty when balance < penalty amount → floor at 0 (no negative balance)
- [RESOLVED] History entry format → soft-delete with `cancelled_at` + show in Lịch sử tab mixed chronologically
- [RESOLVED] Custom bets (kèo phụ) → penalty applies to all bet types (cancel + reduce)
- [RESOLVED] History fields → tên trận, loại cược, stake, điểm bị phạt, ngày hủy
- [RESOLVED] Cancel dialog → show only when `cancel_penalty_enabled = true` AND `penalty > 0`; if penalty = 0 or feature disabled, cancel immediately
- [RESOLVED] Wallet transaction visibility → user sees their own penalty deduction in wallet logs
- [RESOLVED] History tab ordering → mixed chronologically with settled bets
- [RESOLVED] Immediate removal from open bets on cancel → yes
- [RESOLVED] Reduce stake penalty basis → original_stake (immutable), not current stake
- [RESOLVED] Reduce stake penalty formula → floor(excess_reduction × bet_reduce_penalty_percent / 100)
- [RESOLVED] Increasing stake → always free
- [RESOLVED] Reduce stake toggle → no toggle; `bet_reduce_max_percent = 0` = disabled
- [OUT OF SCOPE] Admin view of all users' cancelled bets in house P&L
