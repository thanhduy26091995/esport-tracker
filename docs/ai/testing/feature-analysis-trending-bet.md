---
phase: testing
title: WC Prediction Analytics — Testing Strategy
description: Test cases, manual validation checklist, and quality assurance for the analytics feature
feature: analysis-trending-bet
status: draft
---

# Testing Strategy: WC Prediction Analytics

## Test Coverage Goals

- Unit tests: analytics service profile classification + streak computation logic
- Integration tests: all 3 API endpoints with seeded bet data
- Manual: UI rendering, chart display, empty states, mobile layout

---

## Unit Tests

### `wc_analytics_service.go` — Profile Classification

- [ ] Returns `"underdog_lover"` when underdog_rate > 60%
- [ ] Returns `"goal_hunter"` when avg_goals > 3.0 (even if underdog_rate < 60%)
- [ ] Returns `"draw_master"` when draw_rate > 35%
- [ ] Returns `"aggressive_predictor"` when exact_score_rate > 50%
- [ ] Returns `"conservative_predictor"` when handicap_rate > 60% and avg_stake below community avg
- [ ] Returns `"balanced_predictor"` as default fallback
- [ ] Returns `null`/empty when fewer than 3 settled bets

### `wc_analytics_service.go` — Streak Computation

- [ ] Returns `current_win_streak = 3` for user with 3 consecutive settled wins
- [ ] Returns `current_lose_streak = 2` for user with 2 consecutive losses after a win
- [ ] Stops counting at first opposite outcome
- [ ] `longest_win_streak` tracks all-time best, not just current
- [ ] Void bets are excluded (streak continues through void)
- [ ] Empty bet list returns all zeros

### Accuracy Calculation

- [ ] `accuracy = wins / (wins + losses)` — void excluded
- [ ] 0 settled bets → accuracy = 0.0 (not NaN/division-by-zero)
- [ ] All wins → accuracy = 1.0
- [ ] All losses → accuracy = 0.0

---

## Integration Tests

Test data setup: create 2–3 WC users + seed `wc_bets` + `wc_matches` rows before each test.

- [ ] `GET /wc/analytics/my` — returns 200 with correct accuracy for user with known settled bets
- [ ] `GET /wc/analytics/my?period=7d` — timeline limited to last 7 days
- [ ] `GET /wc/analytics/my` — user with 0 bets returns zero-value response (no 500 error)
- [ ] `GET /wc/analytics/my` — unauthenticated request returns 401
- [ ] `GET /wc/analytics/community` — returns aggregated community stats
- [ ] `GET /wc/analytics/community` — `top_predictors` sorted by accuracy DESC, min 3 settled bets
- [ ] `GET /wc/analytics/compare` — returns both `me` and `community` sections
- [ ] `GET /wc/analytics/compare` — `me.accuracy` matches `GET /my` accuracy for same user
- [ ] All endpoints exclude void bets from accuracy calculation

---

## End-to-End Tests (Manual)

### My Analytics tab

- [ ] Navigate to `/world-cup/analytics` while logged in — page loads without error
- [ ] Accuracy % matches what you'd expect from your known bet history
- [ ] Period filter (7d / 30d / all) changes the accuracy timeline chart
- [ ] Profile label is displayed and matches the described behavior
- [ ] Win/loss streak numbers are correct
- [ ] Favorite teams list shows teams you actually bet on most
- [ ] BetType doughnut chart segments sum to total bets

### Community tab

- [ ] Trending teams chart shows team names and counts
- [ ] Prediction distribution doughnut shows home/away/other split
- [ ] Top predictors table is sorted by accuracy, shows name + accuracy + bet count

### Compare tab

- [ ] All 5 metrics shown for both "Me" and "Community"
- [ ] Rows where I'm above community average are visually highlighted

### Empty states

- [ ] New user (0 bets): "My Analytics" shows friendly empty state with CTA
- [ ] Bets placed but none settled: shows "pending settlement" message
- [ ] Community with < 3 qualifying predictors: top predictors shows "not enough data"

### Mobile (375px viewport)

- [ ] Charts are responsive and not clipped
- [ ] Cards stack vertically
- [ ] Tab navigation doesn't overflow horizontally

---

## Test Data

Minimum seed for meaningful tests:

```sql
-- 2 matches (completed)
-- User A: 5 settled bets (3 wins, 2 losses), 2 handicap, 2 exact_score, 1 OU
-- User B: 4 settled bets (1 win, 3 losses)
-- User C: 1 settled bet (1 win) — below threshold, shouldn't appear in top_predictors
```

Streak test: User A bets in this order (newest first): win, win, win, loss → current_win_streak = 3.

---

## Manual Testing Checklist

- [ ] Profile label displays in both VI and EN locales
- [ ] All chart axis labels and tooltips are localized
- [ ] All table headers use i18n keys (no hardcoded English/Vietnamese)
- [ ] Page title in browser tab is set (via `useHead` or document.title)
- [ ] Loading spinner shows while data is being fetched
- [ ] Error state shows if API returns 500 (test by temporarily disabling backend)
- [ ] Chart tooltips are readable (font size, contrast)
- [ ] No console errors in browser DevTools

---

## Bug Tracking

- Report bugs as GitHub issues tagged `analytics`
- Priority: P1 = wrong accuracy calculation, P2 = chart not rendering, P3 = UI alignment
