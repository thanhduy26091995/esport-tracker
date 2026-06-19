---
phase: requirements
title: WC Admin Block/Unblock User — Requirements
description: Admin can block a WC user, preventing them from placing bets or predictions
---

# Requirements & Problem Understanding

## Problem Statement

Admin hiện không có cơ chế kiểm soát khi một user cụ thể cần bị ngừng hoạt động (tạm thời hoặc vĩnh viễn). Ví dụ: user đặt cược không trung thực, tài khoản bị tổn hại, hoặc admin cần "đóng băng" một user trong lúc kiểm tra số dư. Không có block/unblock, admin phải xóa user (destructive) hoặc không có cách nào cả.

## Goals & Objectives

**Primary goals:**
- Admin có thể block một WC user từ admin panel
- User bị block không thể đặt cược (PlaceBet) **và** không thể dự đoán (PlacePrediction) — nhận lỗi 403
- Khi block user: tất cả pending bets của họ bị auto void (hoàn tiền)
- Admin có thể unblock user bất cứ lúc nào
- Trạng thái blocked hiển thị rõ ràng trong user management

**Secondary goals:**
- Blocked user vẫn có thể đăng nhập và xem các trang (chỉ không đặt cược/dự đoán được)

**Non-goals:**
- Không có thông báo email/push khi bị block
- Không có tự động block theo điều kiện (manual only)

## User Stories & Use Cases

- **Admin block user:** Tìm user trong danh sách → nhấn "Khóa" → user bị block ngay lập tức. Trong panel hiện badge "Bị khóa".
- **User bị block đặt cược:** Gọi POST /wc/matches/:id/bet → nhận 403 `user is blocked from placing bets`.
- **Admin unblock user:** Tìm user bị block → nhấn "Mở khóa" → user có thể đặt cược lại.
- **Admin xem danh sách:** Thấy ai đang bị block (badge màu đỏ) trong user table.

## Success Criteria

- [ ] `wc_users.is_blocked` field tồn tại và DB migration chạy thành công
- [ ] `PUT /admin/users/:id/block` và `PUT /admin/users/:id/unblock` hoạt động đúng
- [ ] Block user tự động void toàn bộ pending bets, hoàn tiền vào wallet
- [ ] PlaceBet từ chối với 403 khi user bị block
- [ ] PlacePrediction từ chối với 403 khi user bị block
- [ ] Admin panel hiển thị blocked status và có button toggle
- [ ] Unblock khôi phục hoàn toàn quyền đặt cược và dự đoán
- [ ] `is_blocked` field được include trong `GET /admin/users` response

## Constraints & Assumptions

- DB migration: `ALTER TABLE wc_users ADD COLUMN is_blocked BOOLEAN NOT NULL DEFAULT FALSE`
- Block chỉ apply cho WC system — không ảnh hưởng core esport system
- Admin không thể block chính mình (safety guard)
- Action là admin-only (WcAdminMiddleware)

## Questions & Open Items

- ~~**Block prediction cũng không?**~~ **Resolved:** Block cả Bet + Prediction.
- ~~**Pending bets khi block?**~~ **Resolved:** Auto void pending bets khi block (hoàn tiền vào wallet).
- **Log block action?** → Scoped out cho phase 1. Có thể thêm audit log sau.
