---
phase: requirements
title: World Cup 2026 Tracker & Betting
description: Match schedule/results tracker + in-team point betting for FIFA World Cup 2026
---

# Requirements & Problem Understanding

## Problem Statement

The team wants to track all FIFA World Cup 2026 matches (schedule, live status, final scores) in one place, and run a fun internal betting competition using a dedicated point wallet — separate from the main ranking system.

**Affected users:** All team members who want to follow WC2026 together and compete in a betting pool.

**Current workaround:** Following results on external sites with no shared tracking, no internal betting.

---

## Goals & Objectives

### Primary goals
1. **WC Match Tracker**: Browse all ~104 WC2026 matches — upcoming, in-progress, completed — filtered by group, stage, or date.
2. **Point Betting — Handicap (Có chấp)**: Bet on 90-min result with admin-set handicap and per-side odds multiplier (Asian handicap style, e.g., home ×0.85 / away ×0.95).
3. **Point Betting — Exact Score (Dự đoán tỉ số)**: Admin publishes a list of bettable scorelines for each match, each with its own odds multiplier (e.g., 1:0 ×5, 0:0 ×3.5, 0:1 ×7). User picks one or more distinct scorelines and enters a stake per scoreline; correct prediction pays `stake × odds`.
4. **Separate betting wallet**: Each user has a dedicated WC betting balance starting at **0**, completely independent of `current_score` (BXH remains unaffected). Balance can go negative (user owes points). No balance check on bet placement.
5. **Admin control**: Admin syncs match data from a free football API, sets handicap/odds per match, and triggers settlement after each match.
6. **Tournament Settlement (Tất toán giải)**: Admin creates a settlement record at any point (mid-tournament or end of tournament) that snapshots every user's balance **from settled bets only** (pending bets on upcoming/unsettled matches are excluded), calculates money owed/due at a configurable point-to-money rate, and resets all wallets back to 0. Full settlement history is preserved.

### Secondary goals
- Betting leaderboard showing who is up/down during the tournament
- Bet history per user
- Lock bets automatically at scheduled kickoff time (`utcDate` from API), or manually by admin
- Settlement history: full list of past settlement events with per-user breakdown (lời/lỗ)
- **Feature flag**: Admin can turn the entire WC2026 feature on/off from the admin UI; state stored in DB; when off, all WC pages and API endpoints are inaccessible to everyone (including logged-in users)

### Non-goals
- Auth for the existing app (dashboard, ranking pages stay public)
- Email / OAuth login (name + password is sufficient for an internal tool)
- Email / SMS-based password recovery (self-service UI resets to `{name}_@123` using name only)
- Live real-time score updates (API polled on-demand or by admin, not pushed)
- Parlay / accumulator bets (single-match bets only)
- Cash payouts (points only)
- Unlimited scoreline options (admin can add as many as needed, but in practice 6–10 per match is enough)
- In-play/live betting
- Mobile push notifications

---

## User Stories & Use Cases

### Schedule & Results
- As a **team member**, I want to see all WC2026 matches grouped by date/round so I can follow the tournament timeline.
- As a **team member**, I want to see the current score or final result of any match at a glance.
- As a **team member**, I want to filter matches by group (A–L) or stage (Group / R32 / R16 / QF / SF / Final).

### Betting — Place
- As a **team member**, I want to see my current WC betting wallet balance (can be negative if I've lost more than I've won).
- As a **team member**, I want to place a handicap bet (Có chấp) on an upcoming match — choose home or away team, enter stake — before the match locks.
- As a **team member**, I want to place exact score bets (Dự đoán tỉ số) on a match — see a card grid of available scorelines each showing its odds multiplier (e.g., 1:0 ×5.0, 0:0 ×3.5), select one or more scorelines, enter stake per scoreline, see live payout preview for each, before the match locks.
- As a **team member**, I want to see my open bets and their status (pending / won / lost / push).

### Authentication
- As a **team member**, I want to register my WC account on a dedicated registration page (independent of the FC25 player list) by choosing a display name and setting a password.
- As a **team member**, I want to log in with my name and password to access the betting page.
- As a **team member**, I can still browse the match schedule and view handicap kèo + odds without logging in.
- As a **team member**, I want to reset my password to `{name}_@123` using just my name via a self-service UI — no email or admin needed.
- As a **team member**, I want to see all bets placed on a match by everyone at any time (including while betting is still open), to compare strategies with teammates.

### Admin
- As an **admin**, I want to turn the WC2026 feature on or off from the admin UI — when off, all WC pages disappear for everyone and all `/api/v1/wc/*` endpoints (except the toggle itself) return 503.
- As an **admin**, I want to trigger a sync from the football API to import/update WC2026 match data.
- As an **admin**, I want to edit a match's handicap, home odds, and away odds before it kicks off, and optionally lock the match manually to stop further bets.
- As an **admin**, I want to enter the final score for a match and then trigger settlement — the system calculates win/loss for all bets and credits/debits wallets automatically.
- As an **admin**, I want to top up or deduct points from a specific user's wallet at any time, with an optional note, and have every adjustment logged.
- As an **admin**, I want to see the full top-up/deduction history for any user.
- As an **admin**, I want to promote or demote any registered user to/from admin role.
- As an **admin**, I want to preview the current settlement state (who owes/is owed how much money) before committing, with a configurable point-to-money rate.
- As an **admin**, I want to create a settlement event that snapshots all wallet balances, records lời/lỗ per user, resets all wallets to 0, and preserves the full history.
- As an **admin**, I want to mark individual users as "đã thu" or "đã chi" within a settlement event as I collect/pay them one by one.
- As an **admin**, I want to view all past settlement events and their per-user breakdown at any time.

### Leaderboard
- As a **team member**, I want to see a WC betting leaderboard ranking everyone by current wallet balance.

---

## Success Criteria

| Criterion | Target |
|---|---|
| All WC2026 group-stage matches importable via sync | 100% |
| Handicap bet placed & settled correctly | P&L correct across 10 test scenarios |
| Exact score bet placed & settled correctly | Payout = stake × exact_score_odds when predicted score matches; 0 otherwise |
| Push (chấp hoà) in handicap returns stake in full | Verified |
| Bets locked once match kicks off | Match `status = live/completed` → no new bets accepted |
| Schedule page loads in < 1s | With 100+ matches in DB |
| Settlement snapshot correct | `final_balance` per user matches wallet at settlement time |
| Wallet resets to 0 after settlement | All `wc_wallets.balance = 0` post-settlement |
| Settlement history preserved | Past settlement events and per-user details remain queryable |

---

## Constraints & Assumptions

### Technical constraints
- **Auth scope**: WC section only; existing app pages are unchanged and public.
- **JWT secret**: `WC_JWT_SECRET` env var required; the first admin must be seeded directly in DB, after which admins can promote others via the UI.
- **football-data.org free tier**: 10 requests/minute; sync is on-demand by admin, not automated/cron.
- **API coverage**: WC2026 competition data available on football-data.org as competition code `WC`; exact team lineup for knockout rounds filled in as tournament progresses.
- **No extra infrastructure**: Same Go + Vue + PostgreSQL stack; no WebSocket, no Redis.

### Business constraints
- Betting wallet is purely for fun — no integration with the existing ranking/settlement system.
- Wallet starts at 0 and can go negative; no balance check on bet placement (users bet on credit).
- Admin is responsible for verifying scores and triggering match settlement; no automatic score-to-settlement pipeline.
- Tournament settlement (tất toán) converts net point balance to money at admin-set rate; actual cash collection/payment is done manually by admin outside the app.

### Assumptions
- All WC group stage fixtures are seeded on Day 1 via API sync; knockout fixtures are seeded as teams advance.
- Handicap is whole or half numbers only (0.5, 1, 1.5 …) — no quarter-ball handicap splits to keep calculations simple.
- Exact score prediction is for 90-min result only (excluding extra time and penalties).
- Admin adds a list of bettable scorelines to `wc_score_odds` for each match before kickoff; each scoreline has its own odds; user can only bet on listed scorelines.
- A user can place both a handicap bet and multiple exact score bets on the same match (no overall cap per match).
- For handicap: a user cannot bet the same side (home/away) twice on the same match.
- For exact score: a user cannot bet the same scoreline twice on the same match, but can bet any number of distinct listed scorelines.
- All wallets start at 0; there is no initial top-up step. Balance goes up when user wins bets and down (including negative) when user loses.
- A user can place a bet regardless of their current balance — no minimum balance required.

---

## Questions & Open Items

| # | Question | Decision |
|---|---|---|
| 1 | What initial wallet balance? | Always 0 — no admin configuration; balance increases on wins, decreases on losses |
| 2 | Can admin edit bets? | No — bets are immutable once placed |
| 3 | What happens to open bets if match is cancelled? | Stake is refunded (push) |
| 4 | API for sync | football-data.org (free tier, Competition `WC`) |
| 5 | Can users view each other's bets? | Yes — visible to everyone at any time, including while betting is open |
| 6 | Handicap payout model? | Admin sets per-side odds multiplier (e.g., home ×0.85 / away ×0.95); payout = `stake × odds` |
| 7 | What triggers bet lockout? | Automatic at scheduled `utcDate` (kickoff time from API), or admin manual lock |
| 8 | Mid-tournament settlement: open bets? | Ignored — only settled bets from completed matches count toward balance snapshot |
| 9 | Registration tied to FC25 player list? | No — WC registration is a separate page; any user can register with any name + password |
| 10 | Self-service password reset UI? | Yes — user enters their name and password resets to `{name}_@123`; non-goal note removed |
