---
phase: requirements
title: Random Tournament - Requirements
description: Player tier system with handicap and round-robin random tournament generation
feature: random-tournament
created: 2026-04-15
---

# Requirements & Problem Understanding

## Problem Statement
**What problem are we solving?**

- **Problem:** Tổ chức các trận đấu vòng tròn (round-robin) theo nhóm thủ công mất thời gian, không đảm bảo tính công bằng khi có sự chênh lệch kỹ năng giữa các player.
- **Who is affected:** Admin/người tổ chức trận đấu và toàn bộ players trong nhóm.
- **Current situation:** Không có cơ chế ghép cặp ngẫu nhiên hay tính handicap; mọi thứ đều thủ công.

## Goals & Objectives

### Primary Goals
- Thêm `tier` (Pro/Normal/Noop) và `handicap_rate` cho từng player
- Handicap tự động ảnh hưởng đến kết quả trận đấu trong tournament
- Tạo random tournament với round-robin schedule (mỗi cặp đấu 1 lần)
- Hỗ trợ 1v1 và 2v2; ghép đội tự động công bằng theo tier
- Lưu tournament để xem lại tiến độ
- Kết quả tournament có thể tính hoặc không tính vào điểm thường (config per tournament)

### Secondary Goals
- Tournament có thể thu phí tham gia (entry_fee)
- Xem bảng xếp hạng trong tournament
- Quản lý nhiều tournament song song

### Non-Goals (Out of Scope)
- Thể thức loại trực tiếp (knockout/bracket)
- Playoff sau round-robin
- Notification / real-time push
- Quản lý prize pool tiền giải

## User Stories & Use Cases

### Tier Management
- **As an admin**, I want to set a player's tier (Pro/Normal/Noop) so that the system knows their skill level
- **As an admin**, I want to set a player's handicap rate (e.g., 0.5 or 1.0) so that it's applied during tournament matches
- **As a viewer**, I want to see each player's tier on the leaderboard / player list

### Tournament Creation
- **As an admin**, I want to create a tournament by selecting a list of players and a match type (1v1 or 2v2) so that a round-robin schedule is automatically generated
- **As an admin**, I want to configure whether tournament results count toward regular match scores
- **As an admin**, I want to optionally set an entry fee for the tournament

### Team Assignment (2v2)
- **As the system**, I want to auto-assign players into balanced teams: Pro players are always paired with Normal or Noop players (prefer Noop if available), ensuring no team is entirely Pro
- **As the system**, I want to randomize remaining pairings after tier constraints are satisfied

### Match Result Recording
- **As an admin**, I want to record the actual FC25 score (team1_goals vs team2_goals) for each tournament match
- **As the system**, I want to apply team handicap to determine the effective winner: `effective_score = actual_score - team_handicap`
- **As the system**, if `affects_score = true` **and effective winner is not a draw**, I want to create a regular match entry so points are updated normally
- **As the system**, if effective result is a draw (after handicap) and `affects_score = true`, I want to **skip** creating a regular match (no points awarded)
- **As an admin**, I want to **override** a previously recorded result; the system will delete the linked regular match and recreate it with the corrected winner
- **As a viewer**, I want to see all tournament matches in the main match history (tagged as "tournament-only" when `affects_score = false`), as well as in the tournament detail view

### Tournament Tracking
- **As a player**, I want to see the full round-robin schedule and current results
- **As a player**, I want to see the tournament leaderboard (wins/draws/losses within tournament)
- **As an admin**, I want to mark a tournament as completed when all matches are done

## Handicap Rules
- Each player has a `handicap_rate` (float, default: 0.0)
- **In 1v1:** each side's handicap = player's `handicap_rate`
- **In 2v2:** team handicap = `max(handicap_rate of players in that team)`
- **Effective score:** `effective_score_team = actual_goals - handicap`
- **Example:** Pro(0.5) vs Normal(0.0), score 1-1 (tie) → effective: 0.5 vs 1.0 → Normal team wins

## Success Criteria

### Functional
- ✅ Player tier and handicap_rate can be set/edited without breaking existing players (default Normal / 0.0)
- ✅ Tournament generates a valid round-robin schedule for 2–16 players
- ✅ 2v2 team assignment ensures no Pro player is paired with another Pro
- ✅ Handicap is correctly applied to determine winner
- ✅ When `affects_score = true`, a regular match entry is created for score/debt tracking
- ✅ Tournament status progresses: active → completed (no draft; schedule generated immediately at creation)

### Non-Functional
- Schedule generation < 200ms for up to 16 players
- Backward compatible: all existing users default to tier=Normal, handicap_rate=0.0

## Constraints & Assumptions
- Player count per tournament: min 3 (meaningful round-robin), max 16 (practical limit for 1 session)
- For 2v2: player count must be even (minimum 4)
- Tier enum: `pro`, `normal`, `noop` (lowercase in DB/API, display as "Pro"/"Normal"/"Noop")
- Handicap rate typical values: 0.0, 0.5, 1.0 — but stored as float for flexibility

## Questions & Open Items
- [x] Khi tournament `affects_score = false` thì vẫn hiển thị match trong history không? → **Có, hiện cả 2 nơi: main match history (tag "tournament-only") và tournament view**
- [ ] Entry fee tournament — tiền thu từ đâu, ghi vào fund không? (Deferred to future phase)
- [x] Có cần tính tie-breaker trong bảng xếp hạng tournament không? → **Có: tie-breaker = goal difference (bàn thắng - bàn thua)**
- [x] Có thể sửa kết quả trận đã nhập không? → **Có, override trực tiếp; nếu affects_score=true thì delete + recreate linked regular match**
- [x] Effective draw + affects_score=true → **Không tạo regular match, chỉ lưu kết quả trong tournament**

## Resolved Decisions (for reference)
| Question | Decision |
|----------|----------|
| Effective draw + affects_score | Skip regular match creation |
| Tournament match in main history | Always show, tagged "tournament-only" when affects_score=false |
| Tie-breaker | Goal difference (actual goals scored − conceded) |
| Result override | Delete old regular match + recreate with new winner |
| Tournament status model | active → completed only (no draft) |
