---
phase: requirements
title: Dashboard Player Sort Strategy
description: Add multiple player sort strategies to the Dashboard leaderboard so the game host can switch between debt-focused, winner-focused, and default views
---

# Requirements & Problem Understanding

## Problem Statement
**What problem are we solving?**

The Dashboard leaderboard currently displays players in a fixed order (`current_score DESC`) — highest positive score first. This single order is not always the most useful view:

- When settling debts, you want the **worst debtors first** (most negative score) so you know who needs to pay.
- When celebrating winners, you want **positive scores first**.
- The game host needs a quick way to switch between these views without leaving the dashboard.

The API endpoint `GET /api/v1/users` currently returns users ordered by `current_score DESC, name ASC` from the backend, and `Leaderboard.vue` re-sorts client-side in the same direction. There is no user-controlled sort strategy.

**Who is affected?**
- Game hosts and players viewing the Dashboard leaderboard.

**Current workaround:** None — users must mentally reorder the list.

## Goals & Objectives

**Primary goals:**
- Add three sort strategies selectable on both the **Dashboard** leaderboard and the **Users** (`/users`) page.
- Dashboard defaults to **"Debt First"** (calls `GET /users/payment-ranking` on mount).
- Strategy 1 — **"Who pays more first" (Historical debt total):** Sort users by their **total money paid across all historical debt settlements** (`SUM(money_amount)` from `DebtSettlement` records where they were the debtor), descending. Users with no settlements appear last, ordered by name. This requires reading `settlementStore.settlements` in addition to `userStore.users`.
- Strategy 2 — **Default (Negative → Positive → 0):** Group by current_score sign — debt group (score < 0, most negative first), then winners (score > 0, highest first), then neutral (score = 0, alpha). Zeros come last.
- Strategy 3 — **"Winners first" (Positive → 0 → Negative):** Winners (score > 0, highest first), then neutral (0, alpha), then debtors (score < 0, most negative last).
- Persist the selected strategy in `localStorage` so it survives page refresh.

**Secondary goals:**
- Keep the sort control compact and non-intrusive on mobile.

**Non-goals:**
- Generic backend `?sort=` query parameter on existing endpoints (a dedicated endpoint is used instead).
- Adding new sort axes beyond `current_score` groups and settlement history (e.g., tier, handicap).

> **Note:** Strategy 1 ("Debt First") requires a new backend endpoint `GET /api/v1/users/payment-ranking` — backend is not completely change-free. Strategies 2 and 3 remain client-side only.

## User Stories & Use Cases

- As a **game host**, I want to see the biggest debtors at the top so I know who must pay first.
- As a **player**, I want to see the winners at the top to feel motivated by the leaderboard.
- As a **game host**, I want my chosen sort preference remembered so I don't need to re-select it every time I open the dashboard.

**Key workflows:**
1. Host opens Dashboard → sees leaderboard in "Debt First" order by default (or last-used sort if stored in localStorage).
2. Host clicks a sort button (e.g., "Winners First") → leaderboard re-orders instantly.
3. Host refreshes → same sort is restored from `localStorage`.
4. Host navigates to `/users` → same sort control is available (2 strategies: Default and Winners First; "Debt First" on Users page is out of scope unless the payment ranking is also fetched there).

**Edge cases:**
- All players have score 0: Strategies 2 and 3 produce identical alphabetical results. Strategy 1 still sorts by settlement history.
- Only two players: sort strategies still apply correctly.
- New user added via inline form while a non-default sort is active: list re-sorts with the active strategy.
- User has never been settled (no `DebtSettlement` records): their total is 0 — they appear after all players who have paid, sorted by name.
- Two users have paid the same total: tie broken alphabetically by name.
- Strategy 1 with settlements not yet loaded: fall back to name-alphabetical until settlement data is available.

## Success Criteria

- [ ] Three sort buttons visible in the Dashboard leaderboard card header.
- [ ] Each sort produces the correct grouping and ordering documented above.
- [ ] Selected strategy persists across page refresh (localStorage key: `dashboard-player-sort`).
- [ ] No additional API calls triggered on sort change.
- [ ] Sort control renders correctly on mobile (≤ 375 px width).

## Constraints & Assumptions

- **Technical:** Sort is entirely client-side; `userStore.users` already contains all active users.
- **Strategy 1 data dependency:** `settlementStore.settlements` must be loaded. `DashboardView.vue` already fetches settlements on `onMounted` — no new API call needed. However, we must verify the settlement store loads **all** settlements, not just recent ones (currently `.slice(0, 5)` is used for display, but the store may hold all). If the store only loads recent settlements, Strategy 1 will under-count historical debt.
- **Assumption:** `current_score` is an integer — ties within a group are broken alphabetically by name.
- **Assumption:** The leaderboard `limit` prop on Dashboard is 10 — sort applies before slicing (so the 10 most relevant players for the active sort are shown).
- **i18n:** Sort labels must be added to both `en.json` and `vi.json` (labels: "Debt First" / "Default" / "Winners First").

## Questions & Open Items

- **[RESOLVED]** Strategy 1 ordering: it is a historical sort by total money paid across all settlements (sum of `money_amount`), not a current-score-based sort.
- **[RESOLVED]** Strategy 1 vs Strategy 2 distinction: they are now fundamentally different — Strategy 1 uses settlement history, Strategy 2 uses current score grouping.
- **[RESOLVED]** Button labels: "Debt First" / "Default" / "Winners First".
- **[RESOLVED]** Limit: keep `limit=10`, sort applies before slice.
- **[RESOLVED]** Does `settlementStore` fetch **all** historical settlements? Yes — `DashboardView` calls `fetchSettlements()` with no params → `GET /settlements` with no limit/offset → backend returns all records. Strategy 1 has full historical accuracy with no additional API calls.
- **[OPEN]** Should Sort control also appear on `/users` page in a future iteration? → Out of scope for this feature, but worth noting.
- **[OPEN]** What happens to the sort control if only one strategy is meaningful (e.g., first day of game, no settlements yet)? → Strategy 1 degrades gracefully to alphabetical; UI still shows all 3 buttons.
