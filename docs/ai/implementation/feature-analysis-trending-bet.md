---
phase: implementation
title: WC Prediction Analytics — Implementation Guide
description: Technical implementation notes, patterns, and code guidelines for the analytics feature
feature: analysis-trending-bet
status: draft
---

# Implementation Guide: WC Prediction Analytics

## Development Setup

**Prerequisites:**
- Backend running: `cd backend && go run cmd/server/main.go`
- Frontend running: `cd frontend && npm run dev`
- Chart.js installed: `cd frontend && npm install chart.js vue-chartjs`

**No DB migration needed.** All analytics are derived from existing tables.

---

## Code Structure

### Backend additions

```
backend/internal/wc/
├── wc_analytics_repository.go   # SQL aggregate queries
├── wc_analytics_service.go      # Profile classification, DTO assembly
└── wc_analytics_handler.go      # HTTP handlers (3 endpoints)
```

Wire in `backend/internal/api/router.go` under the existing WC auth middleware group.

### Frontend additions

```
frontend/src/
├── services/
│   └── wcAnalyticsApi.ts
├── stores/
│   └── wcAnalyticsStore.ts
├── views/
│   └── WcAnalyticsView.vue
└── components/wc/analytics/
    ├── MyAnalyticsPanel.vue
    ├── CommunityPanel.vue
    ├── ComparePanel.vue
    ├── AccuracyTimelineChart.vue
    ├── BetTypeChart.vue
    ├── TrendingTeamsChart.vue
    └── PredictionDistributionChart.vue
```

---

## Implementation Notes

### Core Features

#### Accuracy calculation
```go
// In repository: accuracy = wins / (wins + losses), void excluded
// payout > 0 AND status = 'settled' → win
// payout = 0 AND status = 'settled' → loss
// status = 'void' → excluded entirely

SELECT
    COUNT(*) FILTER (WHERE status = 'settled' AND payout > 0) AS wins,
    COUNT(*) FILTER (WHERE status = 'settled' AND payout = 0) AS losses,
    COUNT(*) FILTER (WHERE status = 'pending') AS pending
FROM wc_bets
WHERE user_id = $1
```

#### Favorite teams (handicap bets only)
```go
// Join wc_bets with wc_matches, map selection to team name
SELECT
    CASE
        WHEN b.selection = 'home' THEN m.home_team
        WHEN b.selection = 'away' THEN m.away_team
    END AS team,
    COUNT(*) AS bet_count
FROM wc_bets b
JOIN wc_matches m ON m.id = b.match_id
WHERE b.user_id = $1
  AND b.type = 'handicap'
  AND (b.selection = 'home' OR b.selection = 'away')
GROUP BY team
ORDER BY bet_count DESC
LIMIT 5
```

#### Streak computation
```go
// Walk settled bets newest-first, count consecutive wins/losses
func computeStreak(bets []WcBetRow) (currentWin, currentLose, longestWin int) {
    // bets ordered DESC by created_at, status='settled'
    // iterate: if payout > 0 → win; else → loss
    // stop incrementing current streak on first opposite outcome
}
```

#### Profile classification
Apply in order (first match wins):
```go
func ClassifyProfile(stats MyBetStats) string {
    if stats.UnderdogRate > 0.60 { return "underdog_lover" }
    if stats.AvgGoalsPredicted > 3.0 { return "goal_hunter" }
    if stats.DrawRate > 0.35 { return "draw_master" }
    if stats.ExactScoreRate > 0.50 { return "aggressive_predictor" }
    if stats.HandicapRate > 0.60 { return "conservative_predictor" }
    return "balanced_predictor"
}
```

For underdog detection (simple proxy): when `selection = 'away'` in matches where `home_handicap < 0` (home team gives handicap → home is favourite → away is underdog).

If handicap line data is not consistently populated, fall back to: `away selection` as proxy for underdog.

#### Average goals predicted (exact score bets)
Parse `selection` field for exact score bets: format is `"{home}-{away}"`. Split on `-`, sum goals, average.

```go
func parseGoals(selection string) (int, int, error) {
    parts := strings.SplitN(selection, "-", 2)
    // parse parts[0] and parts[1] as ints
}
```

### Patterns & Best Practices

**Repository struct:** Follow existing WC repository pattern — inject `*gorm.DB` via constructor, implement interface.

**Handler pattern:** Use `ctx.Get("wc_user_id")` (set by `WcAuthMiddleware`) to get the current user's ID. Never trust user-supplied ID.

**DTO types:** Define Go structs for all response bodies in `wc_analytics_handler.go` using json tags. Keep DTOs close to the handler, not in the model layer.

**Vue-chartjs usage:**
```vue
<script setup lang="ts">
import { Line } from 'vue-chartjs'
import { Chart as ChartJS, CategoryScale, LinearScale, LineElement, PointElement, Title, Tooltip } from 'chart.js'
ChartJS.register(CategoryScale, LinearScale, LineElement, PointElement, Title, Tooltip)
</script>
```
Always register only the Chart.js components you use (tree-shaking).

**Pinia store pattern:** Follow existing store pattern (see `wcAuthStore`). Use `ref()` for state, `async` action functions. Don't store raw API responses — map to typed interfaces first.

---

## Integration Points

**Router wiring (backend):**
Add to `backend/internal/api/router.go` inside the authenticated WC group:
```go
wcAuth := wc.Group("", wcAuthMiddleware)
wcAuth.GET("/analytics/my", analyticsHandler.GetMyAnalytics)
wcAuth.GET("/analytics/community", analyticsHandler.GetCommunityAnalytics)
wcAuth.GET("/analytics/compare", analyticsHandler.GetCompareAnalytics)
```

**Router wiring (frontend):**
Add to both the `isSocSite` and non-soc route arrays in `src/router/index.ts`:
```ts
{
  path: '/world-cup/analytics',
  name: 'wc-analytics',
  component: () => import('../views/WcAnalyticsView.vue'),
  meta: { requiresWcAuth: true, requiresGoogleLink: true }
}
```

**Navigation link:**
Find where the WC nav links are rendered (check WcBettingView or the layout component) and add an "Analytics" link using the `wc-analytics` route name.

---

## Error Handling

- Empty analytics (0 bets): return valid response with zero values and `profile_label: null`. Frontend shows empty state.
- DB errors: return `500` with generic error message (follow existing WC handler error pattern).
- Invalid `period` query param: default to `30d` silently.
- Chart component: wrap in `v-if="data && data.length > 0"` to avoid rendering empty charts.

---

## Performance Considerations

All queries run on ~1000 rows max (small friend group). No optimization needed for Phase 1.

If query latency is ever noticed:
- Add composite index: `CREATE INDEX idx_wc_bets_user_status ON wc_bets(user_id, status)`
- Add composite index: `CREATE INDEX idx_wc_bets_created ON wc_bets(created_at DESC)`

Chart.js renders client-side — no SSR concerns. Use `v-once` on static chart configs.

---

## Security Notes

- `GET /wc/analytics/my` uses `wc_user_id` from JWT context — SQL uses `WHERE user_id = $1` parameterized query
- No user can query another user's personal analytics
- Community endpoints aggregate across all users but expose no PII (only `name` + `avatar_url` in top predictors, which are already public on the leaderboard)
- `period` query param is validated to one of `7d` | `30d` | `all` — reject or default anything else
