---
phase: requirements
title: WC Settlement Preview — Requirements
description: Preview popup before executing Tính kết quả / Tính điểm toàn bộ / Tính lại toàn bộ
---

# Requirements & Problem Understanding

## Problem Statement

Hiện tại, khi admin nhấn **"Tính kết quả"** (FinalizeMatch), **"Tính điểm toàn bộ"** (FinalizeAll), hoặc **"Tính lại toàn bộ"** (RefinalizeAll), hệ thống thực thi ngay lập tức — không có bước xác nhận hay preview. Admin không biết:
- Mỗi user được/mất bao nhiêu điểm
- Ai thắng, ai thua trong từng trận
- House tổng cộng lời/lỗ bao nhiêu điểm

Nếu admin nhấn nhầm hoặc nhấn khi điểm số chưa chính xác, việc rollback rất phức tạp.

## Goals & Objectives

**Primary goals:**
- Hiển thị popup preview trước khi thực thi bất kỳ hành động settle/finalize nào
- Preview mirror chính xác những gì action thực thi — per match, per bet/prediction: loại kèo, odds/hệ số, kết quả, số tiền/điểm
- House summary ở cuối popup: tổng kết lời/lỗ cho admin
- Admin đọc xong mới nhấn confirm để thực thi

**Secondary goals:**
- Tái sử dụng component preview cho cả 3 button (single match, finalize-all, refinalize-all)
- Không thay đổi logic settle hiện tại — chỉ thêm bước preview phía trước

**Non-goals:**
- Không preview settlement tiền (WcSettlementPreview đã có)
- Không thay đổi thuật toán tính điểm
- Preview "Tính lại toàn bộ" không cần hiển thị diff (điểm cũ → mới) — chỉ hiện điểm mới

## User Stories & Use Cases

- **Admin muốn tính kết quả 1 trận:** Nhấn "Tính kết quả", popup hiện từng bet/prediction của trận đó: user, loại (handicap/tài xỉu/hệ số), kết quả (thắng/thua), điểm hoặc tiền thay đổi. Dưới cùng có tổng kết (house summary). Admin kiểm tra, nhấn "Xác nhận" → thực thi.
  - Ví dụ: "User B — Handicap +1: Thắng | Tài Xỉu 1.68: Thua | Hệ số x10: Đúng"
- **Admin muốn tính điểm toàn bộ:** Nhấn "Tính điểm toàn bộ", popup hiện tất cả các trận chưa settle (collapsible per match). Admin review tổng kết, nhấn "Xác nhận".
- **Admin muốn fix điểm:** Nhấn "Tính lại toàn bộ", popup hiện kết quả mới sẽ được áp dụng. Admin confirm.
- **Admin thấy kết quả lạ:** Nhìn preview thấy kết quả không đúng → đóng popup, kiểm tra lại tỉ số trước khi thực thi.

## Success Criteria

- [ ] Mọi 3 button đều mở popup preview trước khi thực thi
- [ ] Preview cho single match: hiển thị đúng user name, prediction type, result, net points
- [ ] Preview cho bulk: hiển thị từng trận, từng user
- [ ] House summary: tổng điểm payout (points_earned) vs tổng điểm stake (points)
- [ ] Nếu không có gì để preview (0 predictions), hiển thị thông báo và cho phép confirm
- [ ] Preview không ghi gì vào DB — read-only

## Constraints & Assumptions

- Preview phải tính toán bằng cùng logic evaluate hiện tại (không được diverge)
- Trận chưa có điểm số (home_score/away_score null) sẽ bị skip trong preview
- Chấp nhận latency preview < 500ms (số predictions nhỏ)
- Không cần cache — admin gọi ít
- Frontend đã có `el-dialog` pattern từ Element Plus

## Questions & Open Items

- ~~**Tính lại toàn bộ**: preview có cần hiển thị diff?~~ **Resolved:** Không cần diff — chỉ hiện điểm mới.
- **Bulk preview size**: nếu có 64 trận × 20 user = nhiều rows, UI có bị chậm không?
  → Resolved: group by match, collapsible rows per match (el-collapse).
- **Preview data scope:** Preview mirror exact cùng data action sẽ xử lý — cả WcBet (money) và WcPrediction (points) cho match đó.
