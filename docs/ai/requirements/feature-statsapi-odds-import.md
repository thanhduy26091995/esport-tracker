---
phase: requirements
title: StatsAPI Odds Import — Requirements
description: Replace manual handicap/score-odds entry with automated import from thestatsapi.com
---

# Requirements & Problem Understanding

## Problem Statement

Admin hiện phải nhập tay toàn bộ dữ liệu kèo cho mỗi trận WC2026:
- **Kèo chấp**: handicap_team, handicap_value, odds_handicap_home, odds_handicap_away
- **Tài xỉu**: ou_line, odds_over, odds_under
- **Tỉ số cược (exact score odds)**: từng dòng `home_score – away_score @ odds`

Quá trình này tốn thời gian, dễ sai số, và phải lặp lại trước mỗi trận. Thay vào đó, admin muốn tự động hoá qua hai cơ chế:
1. **Import từ TheStatsAPI**: Kèo chấp và tài xỉu lấy từ thestatsapi.com (API key đã có).
2. **Generate từ mô hình Poisson**: Tỉ số cược (exact score) không có trên TheStatsAPI → tự generate odds dựa vào xác suất Poisson từ thống kê đội bóng.

**Ai bị ảnh hưởng:** Admin quản lý hệ thống cá cược WC2026.  
**Workaround hiện tại:** Nhập tay qua dialog "Cấu hình kèo" và "Tỉ số cược" trong WcAdminPanel.

---

## Goals & Objectives

### Primary Goals
- Import kèo chấp (Asian handicap) từ TheStatsAPI bằng 1 click.
- Import tài xỉu (Over/Under) từ TheStatsAPI bằng 1 click, với fallback sang football-data.org nếu cần — đảm bảo mapping giữa hai nguồn để tránh breaking changes.
- Generate danh sách tỉ số cược (exact score odds) tự động từ **mô hình Poisson** dựa vào thống kê đội bóng, do TheStatsAPI không cung cấp correct score market.
- Tự động sync kèo chấp + tài xỉu theo lịch định kỳ (background cron) cho các trận sắp diễn ra.

### Secondary Goals
- Admin vẫn có thể chỉnh sửa thủ công sau khi import/generate (không lock).
- Preview dữ liệu sẽ import trước khi ghi vào DB.
- Ghi log mỗi lần sync (thời gian, kết quả, số trận cập nhật).

### Non-Goals
- Không import kết quả trận đấu (home_score/away_score) — scope riêng.
- Không import lịch thi đấu (fixtures) — đã được handle bởi feature khác.
- Không tự động ghi đè kèo đã được admin set thủ công nếu cron chạy — cron chỉ fill trống.

---

## User Stories & Use Cases

### US-1: Import kèo chấp thủ công
> Là admin, tôi muốn bấm "Import từ StatsAPI" trên một trận cụ thể để tự động điền handicap_team, handicap_value, odds_handicap_home, odds_handicap_away, thay vì nhập tay.

**Acceptance criteria:**
- Button "Import kèo" xuất hiện trong match admin card (cạnh button "Cấu hình kèo" hiện tại).
- Sau khi bấm, hệ thống gọi thestatsapi.com và hiển thị preview dữ liệu sẽ được ghi.
- Admin xác nhận → dữ liệu được upsert vào WcMatch (các field handicap_*).
- Nếu API không có dữ liệu cho trận này → thông báo lỗi rõ ràng, không ghi gì.

### US-2: Import tài xỉu (Over/Under) thủ công
> Là admin, tôi muốn bấm "Import tài xỉu" để tự động điền ou_line, odds_over, odds_under cho một trận từ TheStatsAPI.

**Acceptance criteria:**
- Button "Import O/U" xuất hiện trong match admin card.
- Hệ thống ưu tiên lấy từ TheStatsAPI (qua `statsapi_fixture_id`). Nếu không có ID mapping → fallback sang football-data.org.
- Mapping `external_id` (football-data.org) ↔ `statsapi_fixture_id` (TheStatsAPI) được lưu trong `wc_matches` — không xóa `external_id` hiện tại để tránh breaking changes.
- Preview trước khi ghi. Admin xác nhận → upsert `ou_line`, `odds_over`, `odds_under` vào `wc_matches`.

### US-3: Generate tỉ số cược bằng mô hình Poisson
> Là admin, tôi muốn bấm "Generate tỉ số" để hệ thống tự tính toán odds cho từng scoreline (0-0, 1-0, 1-1...) dựa trên mô hình Poisson, thay vì nhập tay.

**Mô hình Poisson — nguyên lý:**
- Lấy trung bình bàn thắng của mỗi đội từ lịch sử (thống kê mùa giải / thestatsapi.com).
- Dùng phân phối Poisson để tính xác suất từng tỉ số: `P(X=k) = (λ^k × e^−λ) / k!`
- Nhân xác suất home × away để ra xác suất từng scoreline.
- Chuyển xác suất → odds: `odds = 1 / probability × (1 + house_margin%)`.
- Chỉ generate các scoreline có xác suất ≥ ngưỡng tối thiểu (vd ≥ 1%).

**Acceptance criteria:**
- Button "Generate Poisson" xuất hiện trong match admin card (cạnh "Thêm tỉ số" thủ công).
- Admin có thể điều chỉnh `house_margin` (%) trước khi generate — default 10%.
- Preview bảng scoreline + odds tính ra, admin confirm → upsert vào `wc_score_odds`.
- Nếu không đủ dữ liệu thống kê đội → thông báo rõ, không generate.

### US-4: Auto background sync (cron)
> Là hệ thống, tôi muốn tự động sync kèo chấp từ thestatsapi.com mỗi 30 phút cho các trận WC2026 có status = "scheduled" và match_date trong vòng 24h tới — chỉ cho trận chưa có kèo.

**Acceptance criteria:**
- Cron job chạy mỗi 30 phút (configurable).
- **Fill blank only:** Chỉ sync các trận có `handicap_value IS NULL` — bỏ qua trận đã có kèo (dù từ import hay nhập tay).
- Ghi log: thời gian chạy, số trận được cập nhật, số lỗi.
- Admin có thể xem log sync gần nhất trong admin panel.

### US-6: Setup mapping statsapi_fixture_id (one-time)
> Là admin, tôi muốn hệ thống tự động map các trận WC2026 với TheStatsAPI fixture ID một lần duy nhất, để các lần import sau hoạt động đúng.

**Acceptance criteria:**
- Button "Setup StatsAPI mapping" trong admin panel.
- Hệ thống fetch toàn bộ WC2026 fixtures từ TheStatsAPI (`competition_id=comp_6107&season_id=sn_118868`).
- Tự động match với `wc_matches` theo `home_team + away_team + match_date` (±12h).
- Admin xem kết quả mapping, confirm → lưu `statsapi_fixture_id` vào `wc_matches`.
- Các trận không match được → highlight để admin xử lý thủ công.
- Chạy lại được nhiều lần, idempotent.

### US-5: Xem trạng thái sync
> Là admin, tôi muốn biết lần cuối dữ liệu kèo của một trận được sync từ API là khi nào.

**Acceptance criteria:**
- Mỗi WcMatch hiển thị `odds_synced_at` (timestamp lần sync cuối).
- Admin panel hiển thị chip "Đã sync lúc HH:MM" hoặc "Chưa sync".

---

## Success Criteria

- Admin có thể import kèo chấp + tài xỉu cho 1 trận trong < 10 giây.
- Background cron tự động fill kèo cho tất cả trận scheduled trong 24h tới **chưa có kèo** (fill blank only).
- Không có dữ liệu sai/corrupt sau import (preview + confirm trước khi ghi).
- Zero downtime — cron chạy không block API.

---

## Constraints & Assumptions

| Constraint | Chi tiết |
|---|---|
| API key | Admin đã có API key của thestatsapi.com |
| Rate limit | Cần kiểm tra rate limit của thestatsapi.com — cron interval điều chỉnh theo đó |
| TheStatsAPI markets | **[Đã xác nhận]** Có Asian handicap ✅, Over/Under ✅, **Correct score ❌ không có** |
| TheStatsAPI O/U format | `{ "over_under_25": { "over": 1.83, "under": 1.95 } }` — cần parse line từ key tên |
| TheStatsAPI fixture ID | WC2026: `competition_id=comp_6107&season_id=sn_118868`. Odds endpoint: `GET /matches/{id}/odds` |
| ExternalID mapping | `wc_matches.external_id` = football-data.org ID. Cần thêm `statsapi_fixture_id` mới (không xóa cột cũ) |
| Fallback O/U | Nếu `statsapi_fixture_id` chưa có → fallback football-data.org (nếu họ có O/U) hoặc skip |
| Cron infrastructure | Dùng `time.Ticker` goroutine trong Go backend |
| Overwrite policy | Cron chỉ fill blank — không overwrite thủ công. Manual import luôn overwrite |
| Poisson data source | Lấy λ (avg goals) từ TheStatsAPI team stats hoặc admin nhập tay nếu API không có |
| House margin Poisson | Default 10%, admin có thể điều chỉnh trước khi generate |

---

## Questions & Open Items

1. **[Medium]** football-data.org có cung cấp O/U odds không? (để xác định fallback có thực sự khả dụng không)
2. **[Medium]** TheStatsAPI có endpoint nào trả về average goals per team cho WC2026 không? (cần cho Poisson λ)
3. **[Medium]** Rate limit của TheStatsAPI là bao nhiêu? (ảnh hưởng đến cron interval)
4. **[Low]** Cron có cần toggle on/off từ admin panel không?
5. **[Low]** Ngưỡng xác suất tối thiểu để include một scoreline trong Poisson — mặc định 1% có hợp lý không?
