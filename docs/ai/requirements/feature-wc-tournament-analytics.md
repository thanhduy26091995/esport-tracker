---
phase: requirements
title: WC Tournament Analytics — Requirements
description: Overview analytics page for WC2026 tournament stats, top scorers, and group standings pulled from football-data.org + internal DB
---

# Requirements & Problem Understanding

## Problem Statement

Người dùng trong hệ thống WC2026 hiện không có cách nào xem tổng quan thống kê mùa giải: ai đang là vua phá lưới, trận nào có nhiều bàn nhất, tỉ lệ thắng home/away/hoà toàn giải là bao nhiêu. Tất cả thông tin đó đang nằm rải rác ở trang schedule (chỉ xem từng trận) hoặc không có trên hệ thống.

**Ai bị ảnh hưởng:** Toàn bộ WC users (và admin) muốn follow dõi diễn biến giải đấu ngoài việc đặt cược.

**Workaround hiện tại:** Phải tự lên FIFA / Google để tra.

## Goals & Objectives

### Primary goals
- Hiển thị tournament-level stats: tổng bàn thắng, trung bình bàn/trận, trận nhiều bàn nhất, tỉ lệ home/away/draw
- Hiển thị top scorers (vua phá lưới) kéo từ football-data.org
- Accessible cho tất cả WC users (cần login) — không public vì page là một phần của WC app

### Secondary goals
- Stats breakdown theo stage (group stage vs knockout)
- Hiển thị số clean sheets (trận không thủng lưới)
- Cache để không spam football-data.org free API (10 req/phút)

### Non-goals
- Odds analytics: odds-api key không hợp lệ; football-data.org odds yêu cầu paid plan → bỏ
- Betting analytics (ai dự đoán đúng nhất, house P&L): đã có trang riêng
- Player profile chi tiết, match events (cards, substitutions): ngoài scope
- Real-time live stats: không cần, refresh 30 phút là đủ

## User Stories & Use Cases

- **As a WC user**, I want to see how many total goals have been scored so far, so I can follow the tournament's pace.
- **As a WC user**, I want to see the top scorers list (vua phá lưới) with their goals and assists, so I can track star players.
- **As a WC user**, I want to see which match had the most goals, so I can find the most exciting game.
- **As a WC user**, I want to see home vs away vs draw breakdown, so I can understand tournament trends.
- **As a WC user**, I want to see goals by stage (group vs knockout), so I can compare scoring rates.
- **As an admin**, I want all stats to auto-refresh without manual action.

## Success Criteria

- Page loads in < 2s (data served from cache on most requests)
- Top scorers refresh every 30 minutes from football-data.org (auto, via backend cache)
- Match stats (goals, H/A/D) are computed from `wc_matches` in DB — no stale data issues
- All WC auth users can access `/world-cup/analytics`
- Page works correctly during group stage AND after knockout rounds begin
- Displays correctly on mobile (responsive)

## Constraints & Assumptions

### Technical constraints
- football-data.org free tier: **10 requests/minute** — must cache aggressively
- football-data.org scorers endpoint: assists field is sometimes `null` (show `—` in UI)
- odds-api key is invalid → no odds data in this feature
- DB source: `wc_matches` WHERE `status = 'completed'` AND `home_score IS NOT NULL`

### Assumptions
- `external_id` trong `wc_matches` là football-data.org match ID (format `537327`) — confirmed từ sync code
- football-data.org competition code là `WC` → endpoint `/v4/competitions/WC/...`
- Free tier không có player photo — chỉ dùng team crest (team logo)
- Nếu giải chưa bắt đầu knockout, goals_by_stage chỉ có group_stage

## Questions & Open Items

- [x] Odds-api key có dùng được không? → **Có**, `odds-api.io` (domain đúng, đã implement trong `statsapi_sync_service.go`). Tuy nhiên chỉ giữ ~44 events gần nhất, không đủ lịch sử cho analytics → dùng football-data.org thay thế.
- [x] football-data.org có scorers không? → **Có**, với goals + assists (sometimes null)
- [x] Rate limit? → 10 req/min free tier → cache 30 phút
- [ ] Bạn muốn page này cần WC login hay public? (Assumption: cần login — WC auth)
- [ ] Muốn thêm "Most clean sheets" (team nào giữ sạch lưới nhiều nhất) không?
