---
phase: requirements
title: WC Prediction Analytics — Personal & Community Insights
description: Analytics dashboard surfacing personal prediction patterns, community trends, and head-to-head comparisons for WC2026 betting
feature: analysis-trending-bet
status: reviewed
---

# Requirements: WC Prediction Analytics

## Problem Statement

**What problem are we solving?**

Users currently place bets on WC2026 matches but have no visibility into their own prediction patterns or how they compare to the broader community. After a match settles, the only feedback loop is their wallet balance — there is no insight into *why* they win or lose, what tendencies they have, or what the crowd is thinking.

**Who is affected?**

All authenticated WC users who actively place bets. Admin users benefit from the community-level view to understand overall platform engagement.

**Current workaround:** None — users manually track their own history by scrolling through settled bets.

**Goal:** Turn the platform from a pure betting tool into a *prediction social analytics platform* where users return not just to place bets, but to see their personal insights, track improvement, and benchmark against the community.

---

## Goals & Objectives

**Primary goals (MVP — Phase 1):**

1. Personal accuracy tracking — visualize win rate over time with Today / 7d / 14d / 30d / Custom range filters
2. Personal prediction profile — classify the user's betting style (Aggressive, Conservative, Draw Master, Goal Hunter, Underdog Lover)
3. Personal tendencies — favorite teams bet on, favorite scorelines, home/away bias, bet type distribution
4. Streak tracking — current winning and losing streaks, computed per match (net result basis)
5. Community trending — which teams/scorelines the community bets on most (last 7 days, including pending bets); overall prediction distribution (home/draw/away)
6. Top predictors ranking by accuracy — shown only within the Analytics page, not a standalone leaderboard
7. Me vs Community comparison table — 10 key stats side by side

**Secondary goals (Phase 2, out of scope now):**

- Prediction heatmap by hour/day of week
- AI-generated personality narrative
- Gamification badges (Goal Prophet, Giant Killer, Night Owl)
- Weekly evolution charts
- Crowd Agreement / Contrarian Index
- Upset Hunter ranking
- Per-bet-type accuracy drill-down

**Non-goals:**

- No changes to betting logic, settlement, or wallet — read-only analytics
- No new database tables required for Phase 1 (all queries aggregate existing `wc_bets` and `wc_matches` data)
- No real-time streaming — analytics are best-effort computed on request
- No integration with the core esport system tables
- No changes to the existing P&L leaderboard at `/world-cup` — accuracy ranking is separate and only inside this analytics page

---

## User Stories & Use Cases

**As an authenticated WC user, I want to:**

1. See my overall accuracy percentage and how it has changed over time (filtered by Today / 7d / 14d / 30d / custom date range), so I can track improvement.
2. Know my betting style label (e.g., "Goal Hunter") so I understand my tendency.
3. See which teams I bet on most, to be aware of potential bias.
4. See my favorite scorelines to understand if I have recurring picks.
5. See my current win/loss streak and longest streak ever (calculated per match on net payout basis).
6. Compare my stats against the community average across 10 key metrics.

**Key workflows:**

- User navigates to `/world-cup/analytics` via nav link
- Personal stats tab (default) shows their own data
- Community tab shows community-wide aggregate data
- Compare tab shows a side-by-side table of 10 metrics

**Edge cases:**

- User has 0 settled matches: show empty state with CTA to place first bet
- User has bets but none settled yet: accuracy = 0%, no profile label shown (need ≥ 3 settled matches)
- Void bets are excluded from all calculations
- Metrics with no applicable data (e.g., home_bias for user with no handicap bets) show "N/A"

---

## Success Criteria

**Acceptance criteria:**

1. `/world-cup/analytics` route accessible to all WC-authenticated users only (no public access)
2. Personal accuracy uses formula: **win = settled match where total payout > total stake; lose = total payout ≤ total stake** (void excluded). Correctly handles win_half/lose_half outcomes.
3. Profile classification label displayed only with ≥ 3 settled matches
4. Accuracy timeline filter options: Today / 7d / 14d / 30d / Custom date range (date picker)
5. Favorite teams derived correctly from matched team names in `wc_matches`
6. Community trending = teams/scorelines with most bets in last 7 days (includes pending bets, excludes void)
7. Top predictors shown only within the Analytics page (not a separate leaderboard), ranked by accuracy, min 3 settled matches to qualify
8. Compare table shows user stat vs community average for all 10 metrics, with N/A for metrics lacking applicable bets
9. Empty states handled gracefully for new users
10. All strings go through `vue-i18n`
11. Page renders correctly on mobile (responsive)

---

## Compare Table — 10 Metrics

| # | Metric | Source | Notes |
|---|--------|--------|-------|
| 1 | Accuracy | All settled matches | win = net payout > stake |
| 2 | Home Bias | Handicap bets only | % picks on home team |
| 3 | Avg Goals Predicted | Exact score bets only | Parse selection field |
| 4 | Exact Score Rate | All bets | % that are exact score type |
| 5 | Underdog Rate | Handicap bets | % picks on away team (proxy) |
| 6 | Avg Stake | All bets | Risk appetite indicator |
| 7 | Over Preference Rate | O/U bets only | % "Over" picks |
| 8 | Exact Score Hit Rate | Exact score bets settled | % that won |
| 9 | Bet Frequency | All bets | Bets per available match |
| 10 | Last-Minute Rate | All bets | % placed < 2h before kickoff |

Metrics 2, 7, 8 show N/A if user has no applicable bet type.

---

## Constraints & Assumptions

**Technical constraints:**

- No new DB tables — Phase 1 derives everything from `wc_bets` + `wc_matches` + `wc_users`
- No chart library currently installed — Chart.js + vue-chartjs to be added
- **Accuracy formula:** per match (group all bets on a match), win = sum(payout) > sum(stake), lose = sum(payout) ≤ sum(stake). Void bets excluded before summing.
- **Asian handicap outcomes:** win_half results in payout > stake (profit), lose_half results in payout = stake/2 < stake (loss) — formula handles both correctly
- For handicap bets: `selection = "home"` → home_team, `selection = "away"` → away_team (team name from `wc_matches`)
- For exact score bets: `selection = "{home}-{away}"` — parse by splitting on `-`

**Business constraints:**

- Friend-group scale: ~10–20 active users. No caching infrastructure needed.
- All analytics are retrospective only; no predictive models in Phase 1.
- WC2026 runs ~35 days — "30d" filter effectively covers the entire tournament.

**Assumptions:**

- `wc_bets.payout` is correctly set at settlement: 0 for full loss, stake/2 for lose_half, partial for win_half, stake×odds for full win
- Void bets have `status = 'void'` and are excluded from all metric calculations
- `wc_matches.home_team` and `away_team` are canonical team name strings
- `wc_matches.match_date` is in UTC; frontend converts to VN timezone for display
- "Underdog" is approximated as the away team pick (simple proxy; true underdog detection requires handicap line direction which may not always be populated)

---

## Decisions Log

| # | Question | Decision |
|---|----------|----------|
| 1 | Community analytics auth | WC auth required for all analytics |
| 2 | Accuracy timeline filter | Today / 7d / 14d / 30d / Custom date range |
| 3 | "Trending" definition | Most bets placed in last 7 days (simple count) |
| 4 | Compare metrics | 10 metrics (original 5 + Avg Stake, Over Preference, Exact Score Hit Rate, Bet Frequency, Last-Minute Rate) |
| 5 | N/A for missing bet types | Show N/A in compare table |
| 6 | Streak calculation unit | Per match, net result (sum payout > sum stake = win) |
| 7 | Pending bets in trending | Included (excludes only void) |
| 8 | Accuracy leaderboard | Accuracy ranking inside Analytics page only, no standalone leaderboard |
| 9 | win_half / lose_half | win = payout > stake; lose = payout ≤ stake |

---

## Open Questions

*(All resolved — see Decisions Log above.)*
