---
phase: requirements
title: WC Prediction Cancel & Reduce Penalty
description: Penalty system khi user huỷ hoặc giảm điểm prediction, nhằm hạn chế hành vi thay đổi cược tuỳ ý
---

# Requirements & Problem Understanding

## Problem Statement

Hiện tại user có thể huỷ hoặc giảm điểm cược bất kỳ lúc nào mà không có hậu quả. Điều này dẫn đến:
- User đặt cược nhiều pick rồi huỷ bớt khi thấy bất lợi
- Không có "cost" thực sự khi thay đổi quyết định
- Làm mất ý nghĩa của game dự đoán

**Người bị ảnh hưởng:** Tất cả WC player, admin quản lý cuộc chơi.

**Hiện trạng:** User có thể DELETE prediction hoặc giảm `points` thoải mái, ví không bị trừ.

## Goals & Objectives

**Primary goals:**
- Trừ điểm ví khi user huỷ prediction (soft-delete)
- Trừ điểm ví khi user giảm điểm vượt quá ngưỡng % cho phép
- Admin config được % phạt riêng cho từng loại
- User thấy rõ họ bị trừ bao nhiêu trước khi xác nhận
- User thấy lịch sử bị phạt sau khi hành động

**Secondary goals:**
- Wallet log ghi lý do trừ điểm (audit trail)

**Non-goals:**
- Không phạt khi user đổi pick (Home → Away) nhưng giữ nguyên số điểm
- Không phạt khi user tăng điểm cược
- Không phạt khi `cancel_penalty_enabled = false` (admin tắt feature)

## User Stories & Use Cases

**US-1: Huỷ prediction có penalty**
> Là player, khi tôi huỷ một prediction đang pending, tôi muốn thấy popup cảnh báo số điểm sẽ bị trừ, để tôi quyết định có huỷ hay không.

**US-2: Giảm điểm cược có penalty**
> Là player, khi tôi giảm số điểm cược xuống dưới ngưỡng cho phép, tôi muốn thấy popup cảnh báo trước khi xác nhận.

**US-3: Xem lịch sử bị phạt**
> Là player, sau khi huỷ/giảm điểm bị phạt, tôi muốn thấy trong tab Lịch sử: prediction đó với badge "Đã huỷ" và số điểm bị phạt, group theo ngày action. Wallet log chỉ là audit trail nội bộ, không hiển thị UI riêng.

**US-4: Admin config penalty**
> Là admin, tôi muốn bật/tắt penalty và config % phạt riêng cho huỷ và giảm điểm, từ trang admin panel.

**Edge cases:**
- User huỷ prediction khi `cancel_penalty_enabled = false` → không phạt, không popup
- User giảm điểm nhưng không vượt ngưỡng → không phạt, không popup
- User đổi pick (Home → Away) với cùng số điểm → không phạt, không popup
- Penalty > balance ví → trừ tối đa bằng balance (không âm ví)
- User huỷ rồi muốn đặt lại cùng prediction → cho phép (partial unique index)

## Success Criteria

- [ ] Huỷ prediction → ví bị trừ đúng `points × cancel_penalty_percent / 100.0` (float64, không làm tròn — ví dụ cược 2đ huỷ 20% → trừ 0.4đ)
- [ ] Giảm điểm vượt ngưỡng → ví bị trừ đúng công thức reduce penalty
- [ ] Popup confirm hiện đúng số điểm, chỉ khi `cancel_penalty_enabled = true`
- [ ] Tab Lịch sử hiện prediction đã huỷ với badge + penalty amount
- [ ] Wallet history có log entry với note mô tả lý do
- [ ] Sau khi huỷ, user có thể đặt lại cùng prediction (không bị block bởi unique constraint)
- [ ] Settlement không bao gồm prediction đã huỷ
- [ ] Analytics/leaderboard không đếm prediction đã huỷ

## Constraints & Assumptions

**Technical:**
- Penalty là `float64`, không làm tròn (floor bị bỏ)
- Soft-delete: prediction huỷ set `cancelled_at`, không hard-delete
- DB unique index phải là partial index `WHERE cancelled_at IS NULL`
- Mọi ghi ví phải trong DB transaction (atomic)
- Dùng `shopspring/decimal` hoặc float64 nhất quán (hiện tại: float64)

**Business:**
- Admin phải bật `cancel_penalty_enabled = true` thì penalty mới hoạt động
- % config: `cancel_penalty_percent`, `bet_reduce_max_percent`, `bet_reduce_penalty_percent`
- Ví không được âm: penalty cap tại balance hiện tại

**Assumptions:**
- `wc_config` đã có các cột config penalty
- Wallet log đã hỗ trợ `admin_id = NULL` (system-initiated)

## Questions & Open Items

- [x] Tab "Lịch sử" lấy data từ `GET /wc/predictions` — sửa query bỏ filter `cancelled_at IS NULL` để trả cả active + cancelled. Frontend group theo ngày.
- [x] Wallet history không có UI riêng — chỉ là audit trail nội bộ. Penalty info hiển thị trực tiếp trong tab Lịch sử.
- [x] Không cần pagination — số record bounded bởi số trận WC (~64), group theo ngày là đủ.
