---
phase: requirements
title: Bug Fix Batch — 21 June 2026
description: 8 bugs found across the WC2026 prediction/betting system covering UX, responsive layout, data correctness, champion prediction multi-pick, smart cron scheduling, and settlement user-name display
---

# Requirements & Problem Understanding

## Problem Statement

**What problem are we solving?**

Eight bugs discovered in the WC2026 system during active use. The bugs span multiple layers (frontend UX, backend logic, data mapping):

1. `/world-cup` page requires a manual click to enter the prediction page even when the feature flag is already enabled.
2. Users on the schedule page must manually scroll to find the next upcoming match.
3. On `/world-cup/predict`, the "Tất cả dự đoán" toggle inside each match collapses all other matches, breaking independent expand/collapse per match.
4. The admin "Tính điểm toàn bộ" preview dialog shows house P&L aggregated across all matches including already-settled ones — should only reflect the matches being previewed.
5. The "Vô địch" (Champion) tab in `/world-cup/predict` breaks layout on mobile/small screens.
6. Champion prediction is limited to one team per user; the system should support multiple simultaneous predictions.
7. The cron job syncs every 30 minutes unconditionally; sync should be frequent when live matches are active and sparse when none are scheduled.
8. The settlement history table (admin "Tất toán" section) never shows user names — `user_name` field is missing from API response.

**Who is affected?**

- All users: bugs 1–3, 5, 6 (predict page UX, champion panel).
- Admin only: bugs 4, 8 (finalize preview, settlement table).
- Backend/infra: bug 7 (cron).

**Current situation / workaround:**

- Bug 1: Users must click the CTA button manually even when feature is on.
- Bug 2: Users scroll through the entire match list to find upcoming games.
- Bug 3: Viewing all predictions for one match always collapses others.
- Bug 4: Admin sees inflated P&L numbers including already-finalized matches.
- Bug 5: Champion tab is unusable on phones.
- Bug 6: Users can only back one champion team.
- Bug 7: Cron fires 30 min even when no matches are happening (unnecessary API calls) and too infrequently during live games (score updates delayed).
- Bug 8: Settlement table shows an empty name column — admin cannot identify users.

---

## Goals & Objectives

**Primary goals:**

- Fix all 8 bugs so the system behaves as expected for both players and admins.
- Do not introduce regressions in the WC betting, prediction, or score-calculation flows.

**Non-goals:**

- No new features beyond what is described in each bug.
- No changes to the core esport system tables or WC wallet settlement logic.
- No UI redesigns; fixes should be minimal and targeted.

---

## User Stories & Use Cases

1. **As a logged-in user**, when I navigate to `/world-cup` and the feature flag is enabled, I am automatically taken to `/world-cup/predict` without clicking a button.
2. **As a user** on the `/world-cup` schedule page, the view auto-scrolls to the next upcoming match so I don't have to search.
3. **As a user** viewing the predict tab, I can expand multiple match sections to view all predictions simultaneously; expanding one does not collapse others.
4. **As an admin** running "Tính điểm toàn bộ", the house P&L summary in the preview dialog reflects only the matches being finalized now (not already-settled ones).
5. **As a user on mobile**, the "Vô địch" champion tab renders correctly — no horizontal overflow, readable card layout.
6. **As a user**, I can place champion predictions on multiple teams simultaneously, each with a separate points wager.
7. **As the system**, the match-sync cron job fires every 5 minutes when there is a live match and resumes the normal low-frequency schedule when idle.
8. **As an admin** reviewing the settlement history, every row in the details table shows the user's name.

---

## Success Criteria

| # | Criterion |
|---|-----------|
| 1 | Navigating to `/world-cup` while logged in + feature enabled redirects to `/world-cup/predict` immediately |
| 2 | Page auto-scrolls to the first card of the next upcoming match group |
| 3 | Toggling "Tất cả dự đoán" on match A while match B is expanded leaves both expanded |
| 4 | House P&L in finalize-all preview = sum across only non-already-settled matches |
| 5 | Champion tab renders with no overflow on 375px viewport |
| 6 | User can place 2+ champion predictions on different teams; all appear in "Tất cả dự đoán" |
| 7 | Cron interval changes to 5 min when any match has status `live`; reverts to ≥30 min when idle |
| 8 | Settlement history detail rows show `user_name` correctly |

---

## Constraints & Assumptions

- **Bug 1**: Admin users redirect to `wc-predict` (same as regular users) when feature is enabled. No special admin redirect.
- **Bug 2**: The `/world-cup` schedule page will have a link/button from `/world-cup/predict`. Scroll-to-next-match applies every time the schedule page loads, for all visitors.
- **Bug 6** (multi-pick champion) requires removing the UNIQUE constraint on `wc_champion_predictions.user_id`. This is a migration — existing data is safe (one row per user is still valid; the constraint is simply relaxed to allow multiple).
- **Bug 6**: No cap on number of champion predictions per user — user can predict as many different teams as they want (one prediction per team enforced by composite UNIQUE `(user_id, team_id)`).
- **Bug 6**: Points are **not** deducted from wallet at placement time. Settlement is deferred: when admin announces winner, correct predictions receive `points × odds` added to wallet; incorrect predictions have their `points` subtracted from wallet.
- **Bug 6**: Cancel is per-prediction by ID (not bulk cancel all).
- **Bug 7** (smart cron) must not increase total StatsAPI API calls beyond what the free tier allows; smart scheduling with long idle intervals is acceptable. All interval values configurable via env vars.
- **Bug 4**: HouseSummary guard (exclude already-settled matches) applies to `finalize-match` and `finalize-all` only. `refinalize-all` intentionally includes all matches in HouseSummary since the purpose is to see the recalculated total P&L across everything.
- All monetary/points arithmetic continues to use `shopspring/decimal` on the backend.
- i18n: no new UI strings required beyond what's needed; champion multi-pick reuses existing keys where possible.

---

## Questions & Open Items

All questions resolved during requirements review session on 21 June 2026:

| Question | Decision |
|----------|----------|
| Bug 1: Admin redirect destination? | Admin → `wc-predict` (same as regular user) |
| Bug 2: Who does scroll apply to after Bug 1 redirect? | All visitors to `/world-cup`; the predict page will have a link back to schedule |
| Bug 6: Max predictions per user? | No cap — composite (user_id, team_id) unique is the only constraint |
| Bug 6: Are points deducted from wallet at placement? | No — settled on champion announcement only |
| Bug 4: Apply HouseSummary guard to refinalize-all? | No — only finalize-match and finalize-all; refinalize-all shows full totals |
