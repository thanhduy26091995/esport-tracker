---
phase: implementation
title: Player Highlight Feed — Implementation Guide
description: Code structure, query patterns, and integration notes for all 8 highlight categories
---

# Implementation Guide

## Development Setup

No new dependencies. Uses the existing Go + Gin + GORM + PostgreSQL (backend) and Vue 3 + Pinia + Element Plus (frontend) stacks.

## Code Structure

### Backend — new files

```
backend/internal/
  model/
    highlight.go                  # Highlight, HighlightsResponse structs + type constants
  repository/
    highlight_repository.go       # 9 SQL aggregation queries
  service/
    highlight_service.go          # Rule engine for all 8 categories
  api/
    highlight_handler.go          # Gin handler
    router.go                     # register route (existing file)
```

### Frontend — new files

```
frontend/src/
  types/
    highlight.ts                  # Highlight, HighlightsResponse interfaces
  services/
    highlightService.ts           # getHighlights() API call
  stores/
    highlightStore.ts             # Pinia store
  components/
    HighlightFeedPanel.vue        # Four-section feed panel
  views/
    DashboardView.vue             # mount panel (existing file)
```

## Implementation Notes

### Category 1 — Streak queries

Use a window function to find the current consecutive run without N+1 queries:
```sql
-- rank each participant row per user by recency, then find the longest
-- leading run of the same result
WITH ranked AS (
  SELECT mp.user_id, mp.point_change,
         ROW_NUMBER() OVER (PARTITION BY mp.user_id ORDER BY m.match_date DESC) AS rn
  FROM match_participants mp
  JOIN matches m ON m.id = mp.match_id
  WHERE mp.point_change != 0
)
SELECT user_id,
  SUM(CASE WHEN point_change > 0 THEN 1 ELSE 0 END) FILTER (WHERE rn <= streak_len) AS win_streak,
  ...
```
Implement as a raw SQL query in `GetCurrentStreaks()` returning a map of `userID → StreakData`.

### Category 2 — Points movement today

```sql
SELECT mp.user_id,
       SUM(CASE WHEN mp.point_change > 0 THEN mp.point_change ELSE 0 END) AS gained,
       SUM(CASE WHEN mp.point_change < 0 THEN mp.point_change ELSE 0 END) AS lost,
       COUNT(*) FILTER (WHERE mp.point_change != 0) AS matches_today
FROM match_participants mp
JOIN matches m ON m.id = mp.match_id
WHERE m.match_date >= CURRENT_DATE AT TIME ZONE 'Asia/Ho_Chi_Minh'
GROUP BY mp.user_id
```

### Category 3 — Recent form

`GetRecentForm()` returns the last 10 decisive results as a `[]bool` (true=win) per user. The service computes win rate and form string (`WWLWW`) from this slice.

### Category 4 — Activity

`GetWeeklyActivity()` returns `matchesThisWeek` and a consecutive-active-days count (count distinct match days from today backwards without a gap).

### Category 5 — Hot/Cold rule (service layer)

```go
if stats.Last10WinRate >= 0.80 && stats.MatchesToday >= 5 {
    // emit hot_player
}
if stats.CurrentLoseStreak >= 5 {
    // emit cold_player
}
```

### Category 6 — Fast climb/collapse (last hour)

```sql
SELECT mp.user_id, SUM(mp.point_change) AS delta_1h
FROM match_participants mp
JOIN matches m ON m.id = mp.match_id
WHERE m.match_date >= NOW() - INTERVAL '1 hour'
  AND mp.point_change != 0
GROUP BY mp.user_id
```

### Category 7 — Milestones (service layer)

Milestone fires only when `totalWins` or `currentScore` *equals* a threshold — not exceeds — to avoid re-firing every load:
```go
milestoneWins := []int{100, 500, 1000}
for _, threshold := range milestoneWins {
    if stats.TotalWins == threshold {
        // emit milestone_wins
    }
}
```

### Category 8 — Social/fun (service layer)

Each social condition maps to a pool of Vietnamese variants; pick one at random:
```go
var socialCooking = []string{
    "🔥 %s is cooking",
    "🔥 Không ai cản nổi %s hôm nay",
    "🔥 %s đang trong trạng thái thăng hoa",
}
// select: socialCooking[rand.Intn(len(socialCooking))]
```

Trigger conditions:
| Social type | Trigger |
|---|---|
| `social_cooking` | hot_player conditions met |
| `social_tilt` | loseStreak ≥ 3 AND matchesToday ≥ 5 |
| `social_tryhard` | matchesToday ≥ 10 |
| `social_grinder` | matchesToday ≥ 15 |
| `social_sneaky` | rankClimbedToday ≥ 2 AND matchesToday ≤ 5 |

### Patterns & Best Practices

- Service constructor: `NewHighlightService(repo *HighlightRepository) *HighlightService`
- Handler: inject service, return `c.JSON(200, HighlightsResponse{...})`
- Frontend store follows `matchStore.ts` pattern: `loading` ref, silent fail on error
- `HighlightFeedPanel.vue` uses `el-tabs` for the four sections + `el-skeleton` for loading

## Integration Points

**Router (`router.go`):**
```go
v1.GET("/highlights", highlightHandler.GetHighlights)
```

**DashboardView:** add `<HighlightFeedPanel />` in the right column, call `highlightStore.fetchHighlights()` in `onMounted`.

## Error Handling

- Repository error → service returns empty `HighlightsResponse` — handler responds HTTP 200 with empty sections
- Frontend: fetch failure leaves store empty, panel shows empty-state message per section — no error toast

## Performance Considerations

- All 9 queries batch across all users — no per-user loops
- Queries hit `match_participants.user_id` and `matches.match_date` (indexed)
- No caching needed at current scale; add 60 s handler cache if player count exceeds ~200
