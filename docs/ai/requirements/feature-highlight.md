---
phase: requirements
title: Player Highlight Feed
description: Auto-generated news-feed highlighting player streaks, rank/point movement, recent form, activity, hot/cold status, fast climb/collapse, milestones, and social commentary on the Dashboard
---

# Requirements & Problem Understanding

## Problem Statement

The dashboard shows a leaderboard (score ranking) but gives no narrative context about *what has been happening*. Viewers cannot tell at a glance who is on fire, who is struggling, who is grinding, or who just hit a milestone. All the data to answer these questions already exists in `matches` and `match_participants`, but is never surfaced.

**Affected users:** All players and viewers who open the Dashboard to get a quick read on the current state of play.

**Current workaround:** Users must manually scan the match list or player table and mentally piece together trends.

## Goals & Objectives

**Primary goals:**
- Compute and display auto-generated highlight cards on the Dashboard alongside the leaderboard
- Cover all eight highlight categories defined in the spec (see User Stories below)
- Generate highlights on-demand (fresh API call per page load) — no background jobs needed
- Group highlights into four dashboard sections: Trending, Daily Recap, Competitive, Social Feed

**Secondary goals:**
- Display in Vietnamese with emoji indicators matching each highlight type
- Surface the most impactful highlights first (sorted by priority within each section)

**Non-goals (this feature):**
- Push notifications or real-time updates
- Storing highlight history — always recomputed from live data

**Truly not possible (data not in schema):**
- Goals scored / conceded
- Favourite FIFA team / most-picked team
- Clean sheets
- Match duration (only start time stored, no end time)
- Comeback wins (FIFA definition — no halftime score data)

**Possible with current data, deferred to Phase 2:**
- Head-to-head rivalry highlights — `match_participants` has both players per match; can compute 1v1 record between any two players
- Time-of-day performance — `match_date` has full timestamp; can derive morning/evening win rate patterns
- Tournament-specific highlights — `tournament_matches` has round, effective_winner, and status

## Highlight Categories

### 1. Winning / Losing Streak Metrics
**Metrics:** current win streak, current lose streak, longest win streak, longest lose streak, unbeaten streak

**Example highlights:**
- 🔥 Player X đang thắng liên tục 5 trận
- 😵 Player X đã thua 7 trận liên tiếp
- 🛡️ Player X bất bại 10 trận gần nhất
- 💥 Player X vừa kết thúc chuỗi thắng của Player Y
- 🎯 Player X vừa phá chuỗi thua kéo dài

### 2. Rank / Point Movement Metrics
**Metrics:** points gained today, points lost today, biggest point gain (single match), biggest point loss, rank increase, rank decrease, fastest climb, biggest collapse

**Example highlights:**
- 🚀 Player X vừa tăng +120 điểm hôm nay
- 📉 Player X mất -95 điểm chỉ trong 1 giờ
- 🏆 Player X vừa leo lên Top 1
- ⚡ Player X tăng rank nhanh nhất hôm nay
- 😬 Player X tụt 4 bậc trên BXH

### 3. Recent Form Metrics
**Metrics:** last 5 match form, last 10 match form, winrate last 10 matches, daily winrate, weekly winrate

**Example highlights:**
- 🔥 Player X thắng 8/10 trận gần nhất
- 📈 Player X đang có phong độ tốt nhất hôm nay
- 🥶 Player X đang tụt phong độ
- 🎮 Player X là người chơi ổn định nhất tuần

### 4. Activity Metrics
**Metrics:** total matches today, total matches this week, most active player, consecutive active days, matches per hour

**Example highlights:**
- 🎮 Player X chơi nhiều trận nhất hôm nay
- 🕹️ Player X vừa marathon 20 trận
- 🌙 Player X online cực muộn nhưng vẫn thắng
- 📅 Player X active liên tục 5 ngày

### 5. Hot / Cold Player
**Hot conditions:** last10WinRate ≥ 80% AND totalMatchesToday ≥ 5
**Cold conditions:** loseStreak ≥ 5

**Example highlights:**
- 🔥 Không ai cản nổi Player X hôm nay
- 🔥 Player X đang rất nóng với winrate 90%
- 🥶 Player X đang gặp khó khăn với 6 trận thua
- 📉 Player X mất nhiều điểm nhất hôm nay

### 6. Fast Climb / Collapse Metrics
**Metrics:** fastest point gain in 1 hour, fastest point loss in 1 hour, rank climb speed, rank drop speed

**Example highlights:**
- 🚀 Player X tăng điểm nhanh nhất hôm nay
- ⚡ Player X vừa gain 50 điểm trong 30 phút
- 📉 Player X tụt rank mạnh nhất tuần

### 7. Milestone Metrics
**Metrics:** total matches played (100, 500, 1000), total wins (100, 500), total losses, total points (1000, 2000, 5000), first Top-10 entry, first Top-3 entry

**Example highlights:**
- 🏅 Player X đạt 1000 điểm
- 🎉 Player X cán mốc 100 trận thắng
- 👑 Player X lần đầu vào Top 10
- 📈 Player X đạt tỷ lệ thắng 70%

### 8. Social / Fun Highlights
**Templates (triggered by underlying conditions):**
- 🔥 Player X is cooking
- 😅 Player X cần reset mental
- 🎯 Một ngày khó quên với Player X
- 🕹️ Player X đang spam rank
- 💀 Player X vừa bật mode tryhard
- 👀 Player X đang âm thầm leo rank

### Phase 2 — Additional categories (data already available, deferred)

**Head-to-head rivalry**
- 🆚 Player X thắng Player Y 7/10 lần đối đầu
- ⚔️ Player X và Player Y đang ngang bằng nhau

**Time-of-day performance**
- 🌙 Player X có winrate tốt nhất vào buổi tối
- ☀️ Player X chưa thua trận nào buổi sáng

**Tournament highlights**
- 🏆 Player X vô địch giải đấu hôm nay
- 🥈 Player X lọt vào chung kết

## Dashboard Sections

Highlights are grouped into four sections on the panel:

| Section | Content |
|---|---|
| **Trending** | Hot players, biggest climbers today, most active player |
| **Daily Recap** | Best player today, worst losing streak, most matches played |
| **Competitive** | Top rank battle, rank gap changes, rank movement highlights |
| **Social Feed** | Auto-generated social/fun commentary, momentum tracking |

Phase 2 additions:

| Section | Content |
|---|---|
| **Rivalries** | Head-to-head records, nemesis matchups |
| **Tournament** | Tournament-specific highlights and results |

## User Stories & Use Cases

- As a **player**, I want to see my current win/lose streak so I know my momentum.
- As a **viewer**, I want to see "🔥 Player X đang thắng liên tục 5 trận" so I know who to watch.
- As a **viewer**, I want point movement highlights ("🚀 +120đ hôm nay") so I can follow the day's action.
- As a **player**, I want my milestones ("🎉 100 trận thắng") celebrated automatically.
- As a **viewer**, I want to see hot/cold tags to read momentum quickly.
- As a **player**, I want to see recent form (last 10 WWLWW) to track my consistency.
- As a **viewer**, I want to see who is the most active grinder today.
- As a **player**, I want fast-climb highlights ("⚡ gain 50 điểm trong 30 phút") to recognise explosive sessions.
- As a **viewer**, I want social/fun commentary ("😅 Player X cần reset mental") to make the feed entertaining.

**Phase 2 user stories:**
- As a **player**, I want to see my head-to-head record against a rival ("🆚 Player X thắng Player Y 7/10 lần") to track our rivalry.
- As a **viewer**, I want time-of-day stats ("🌙 Player X có winrate tốt nhất vào buổi tối") to see performance patterns.
- As a **player**, I want tournament highlights ("🏆 Player X vô địch giải đấu") to celebrate tournament wins specifically.

**Edge cases:**
- Player with 0 matches today → no activity or point-movement highlight
- Player with < 10 total matches → no hot/cold or recent-form highlight (insufficient sample)
- All players inactive today → Trending/Daily Recap sections show empty state
- Multiple milestones hit simultaneously → one card per milestone threshold
- Player has both a win streak AND is hot → both highlights shown (different types)

## Success Criteria

- `GET /api/v1/highlights` returns all highlights grouped by section within 500 ms for up to 30 active players
- Dashboard panel renders all four sections: Trending, Daily Recap, Competitive, Social Feed
- All 8 highlight categories produce correct highlights against real match data
- Messages are in Vietnamese with correct emoji
- No highlights generated for players with 0 matches
- Panel is responsive on mobile widths

## Constraints & Assumptions

**Technical:**
- Data source: `matches` + `match_participants` + `users` — no new tables required
- Win = `point_change > 0`; loss = `point_change < 0`; draws (`point_change = 0`) excluded from streak/win-rate
- `point_change` is not always ±1 — it equals the `points_per_win` config value (default 1, can be any positive integer). "+120 điểm hôm nay" is realistic when `points_per_win > 1` or many matches are played
- Yesterday's rank is computable: `yesterday_score = current_score − sum(point_change today)`, then sort all users
- "Today" / "this hour" = server local time UTC+7
- `is_active = false` users excluded
- "1 hour" window for fast-climb/collapse = rolling `NOW() - INTERVAL '1 hour'` window on `match_date`

**Business:**
- Hot: last10WinRate ≥ 80% AND matchesToday ≥ 5
- Cold: loseStreak ≥ 5
- Social/fun triggers map to underlying metric conditions (e.g. "spam rank" = matchesToday ≥ 15)
- Milestone thresholds hardcoded for initial release; configurable via `config` table in a follow-up
- At most one highlight per type per player to avoid flooding

**Assumptions:**
- No auth required (same as `/leaderboard`)
- Tournament matches count equally toward all metrics
- Rank position derived from leaderboard order by `current_score` DESC

## Questions & Open Items

- Should milestone thresholds be in the `config` table from day one, or hardcoded initially? *(Hardcode first, migrate to config in follow-up)*
- Cap on total highlight cards shown in panel? *(Cap at 20 cards total, top by priority per section)*
- Social/fun message variants: one template per condition or random pick from multiple? *(Random pick from 2–3 variants per condition for variety)*
