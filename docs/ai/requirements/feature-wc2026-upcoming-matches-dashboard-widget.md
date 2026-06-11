---
phase: requirements
title: WC2026 Upcoming Matches Dashboard Widget — Requirements
description: Show upcoming WC2026 matches (next 48h) in the dashboard above the champion-banner
---

# Requirements & Problem Understanding

## Problem Statement

Người dùng muốn theo dõi lịch thi đấu World Cup 2026 sắp tới ngay trên trang Dashboard mà không cần phải điều hướng sang trang WC Schedule riêng. Hiện tại Dashboard không có thông tin gì về WC2026 — người dùng bỏ lỡ các trận sắp diễn ra.

- **Ai bị ảnh hưởng?** Tất cả người dùng truy cập Dashboard (không cần đăng nhập WC).
- **Tình trạng hiện tại:** Dashboard chỉ có stat cards, champion-banner và bảng lịch sử trận FC25. Không có widget WC.

## Goals & Objectives

**Primary goals:**
- Hiển thị danh sách các trận WC2026 `scheduled` hoặc `live` có `match_date` trong vòng **48 giờ tới** (từ `now` đến `now + 48h`) ngay trên Dashboard.
- Đặt widget **phía trên champion-banner** trong `DashboardView.vue`.
- Widget luôn hiển thị bất kể trạng thái WC feature flag (`wc_config.is_enabled`). Lịch thi đấu là thông tin public.

**Secondary goals:**
- Nếu không có trận nào → ẩn widget hoàn toàn (no empty state). Widget cũng ẩn trong khi đang fetch.
- Click vào trận → navigate tới `/world-cup/schedule` (không mở betting modal).
- Hiển thị: hai đội, giờ thi đấu (VN timezone — Asia/Ho_Chi_Minh), tình trạng trận (`scheduled` → giờ thi đấu; `live` → 🔴 LIVE badge).
- **Giới hạn 5 trận** đầu tiên (sorted by `match_date ASC`). Nếu có hơn 5 trận → hiển thị link "Xem thêm →" dẫn tới `/world-cup/schedule`.
- **Auto-refresh mỗi 5 phút** để cập nhật khi trận chuyển trạng thái sang `live`.

**Non-goals:**
- Không show odds/kèo trên widget dashboard.
- Không hỗ trợ đặt cược trực tiếp từ widget.
- Không thay thế trang WC Schedule đầy đủ.

## User Stories & Use Cases

- **As a** user trên Dashboard, **I want to** thấy các trận WC2026 sắp tới ngay hôm nay và ngày mai, **so that** tôi không bỏ lỡ trận nào và biết cần vào đặt cược khi nào.
- **As a** user, **I want to** click vào một trận trong widget **to** mở trang WC Schedule để xem đầy đủ chi tiết.
- **As a** user lúc không có trận nào sắp diễn ra (ví dụ giữa giai đoạn giải), **I want** widget không xuất hiện, **so that** Dashboard không bị "trống" hay "no matches".

## Success Criteria

- [ ] Widget hiển thị đúng các trận `status = scheduled` hoặc `status = live` có `match_date` trong `[now, now+48h]`.
- [ ] Không hiển thị trận `completed` hoặc `cancelled`.
- [ ] Trận `live` hiển thị badge `🔴 LIVE` thay cho giờ thi đấu.
- [ ] Giờ thi đấu (trận `scheduled`) hiển thị theo múi giờ `Asia/Ho_Chi_Minh`.
- [ ] Tối đa 5 trận được hiển thị; nếu có hơn 5 → có link "Xem thêm".
- [ ] Widget ẩn hoàn toàn (không render DOM) khi: đang fetch hoặc kết quả rỗng.
- [ ] Widget tự động refresh data mỗi 5 phút.
- [ ] Vị trí: ngay trước `.champion-banner` trong dashboard grid (full-width).
- [ ] Backend API `GET /api/v1/wc/matches?date_from=&date_to=` trả đúng kết quả với date range filter.
- [ ] Không gây regression cho trang WC Schedule hiện tại.

## Constraints & Assumptions

- **Technical:** Backend hiện chỉ có `Date` filter (1 ngày cụ thể). Cần thêm `DateFrom` / `DateTo` vào `MatchFilter` và handler.
- **Public API:** `GET /api/v1/wc/matches` không yêu cầu auth — widget gọi trực tiếp mà không cần WC token.
- **Frontend:** Widget không dùng `wcStore` (tránh load toàn bộ match list). Gọi API riêng từ `DashboardView` hoặc tạo `WcUpcomingWidget.vue` component.
- **Assumption:** WC feature flag không ảnh hưởng đến widget — API public `/wc/matches` không bị chặn bởi `WcFeatureMiddleware` (cần confirm).
- **Assumption:** `wc_matches` đã được sync (admin đã chạy `/admin/sync`) trước khi widget có data.

## Questions & Open Items

- [x] **RESOLVED:** `GET /api/v1/wc/matches` là **exempt** khỏi `WcFeatureMiddleware` — route được đặt ngoài `wcFeature` group trong router. Widget luôn có data bất kể feature flag.
- [x] **RESOLVED:** Loading state → ẩn hoàn toàn (invisible) khi đang fetch. Không dùng skeleton.
- [x] **RESOLVED:** Status filter → `scheduled` + `live`. Không show `completed` hay `cancelled`.
- [x] **RESOLVED:** Auto-refresh mỗi 5 phút (`setInterval` + `clearInterval` on unmount).
- [x] **RESOLVED:** Match cap → tối đa 5, còn lại dẫn link "Xem thêm" tới `/world-cup/schedule`.
