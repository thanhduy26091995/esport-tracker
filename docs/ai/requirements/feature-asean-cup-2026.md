---
phase: requirements
title: ASEAN Cup 2026 — Requirements & Problem Understanding
description: Extend the WC2026 betting platform to support ASEAN Cup 2026 with the same UI/UX and bet types
---

# Requirements & Problem Understanding

## Problem Statement

World Cup 2026 is over. The friend group wants to continue the betting game for **ASEAN Cup 2026** (ASEAN Hyundai Cup, AFF Championship successor, July 24 – August 26 2026). The existing WC2026 platform already has everything needed — handicap bets, tài xỉu (O/U), kèo phụ (custom bets), champion predictions, wallets, leaderboard, house P&L — but it is hard-coded to "World Cup". We need to make the platform **tournament-agnostic** so the same codebase and same users can participate in a new tournament, while preserving the WC2026 historical data as a read-only archive.

**Who is affected:** All existing `wc_users` (same friend group). Admin manages both tournaments.

**Current workaround:** None — a new tournament is impossible without this feature.

## Goals & Objectives

### Primary Goals
- Add ASEAN Cup 2026 as a second tournament on the same platform with the same feature set as WC2026
- Preserve WC2026 data intact and accessible in read-only mode (feature flag off, UI still browsable)
- Share user accounts and wallet across tournaments (one balance, carry-over from WC)

### Secondary Goals (all confirmed in-scope for launch)
- Champion prediction for ASEAN Cup (same as WC champion prediction feature)
- Group standings table (W/D/L/GD/Points) on ASEAN Cup schedule page
- Prediction analytics page (personal accuracy, streaks, trending) for ASEAN Cup
- Activity feed WebSocket toasts when bets are placed (shared channel with WC)
- Top-3 honor banner scoped to active tournament (shows ASEAN Cup top-3 when AC is active)
- Leaderboard scoped per-tournament (net P&L for ASEAN Cup bets only)
- Admin can toggle ASEAN Cup on/off independently of WC flag

### Non-Goals
- No new registration flow — existing `wc_users` log in the same way
- No separate wallet per tournament — a single shared wallet balance (carry-over from WC)
- No wallet reset at ASEAN Cup start — users continue with existing balance
- No migration of WC bet history into ASEAN Cup context
- No changes to the core esport (FC25) system
- Not a generic multi-tournament framework for arbitrary future tournaments (just the `asean_cup` discriminator)
- No aggregate-score betting on 2-legged ties — each leg is bet independently

## User Stories & Use Cases

### Player
- As a player, I want to see upcoming ASEAN Cup 2026 matches and place handicap/O/U/kèo phụ bets on them
- As a player, I want my ASEAN Cup bet history and predictions shown separately from my WC history
- As a player, I want to browse my old WC2026 bets and predictions as a read-only archive
- As a player, I want my existing wallet balance (from WC) to carry over into ASEAN Cup seamlessly
- As a player, I want to pick my ASEAN Cup champion before the tournament starts
- As a player, I want to see ASEAN Cup group standings (W/D/L/GD/Pts) on the schedule page
- As a player, I want to see real-time activity feed toasts when others place ASEAN Cup bets

### Admin
- As an admin, I want to **manually create ASEAN Cup matches** with date, teams, venue, group, and stage (primary path — no confirmed external API)
- As an admin, I want to **manually enter the final score** of a completed match via an "Enter Score" dialog (backend endpoint exists; frontend form is new)
- As an admin, I want to set handicap and O/U lines, define kèo phụ, and settle bets for ASEAN Cup
- As an admin, I want separate feature flags for WC (read-only) and ASEAN Cup (active)
- As an admin, I want a house P&L dashboard scoped to ASEAN Cup
- As an admin, I want to run tất toán (settlement cycle) for ASEAN Cup independently

### Key Workflows

1. **Player bets on a group match** — navigates to `/asean-cup/bet` → picks a match → places handicap/O/U bet → wallet balance visible
2. **Admin enables ASEAN Cup** — toggles ASEAN Cup flag on → `/asean-cup` becomes active for all users
3. **Admin enters match score** — opens "Enter Score" dialog on completed match → types home/away score → saves → clicks "Tính kết quả" → bets settled, wallets updated
4. **Admin settles kèo phụ** — after match score saved, manually picks winning option for each custom bet
5. **Admin runs tất toán** — settlement cycle records net P&L per user for ASEAN Cup; wallet balance adjusts but history preserved
6. **Player views WC archive** — visits `/world-cup` → sees historical WC data, leaderboard, bet history; no new bets possible (feature flag off)
7. **Two-legged knockout** — admin creates two separate matches (Leg 1 + Leg 2) for each SF/Final tie; each leg is bet on independently; admin enters score and settles each leg after it is played

## Success Criteria

- [ ] Users can navigate to `/asean-cup` and see ASEAN Cup 2026 matches
- [ ] All three bet types work for ASEAN Cup: handicap, tài xỉu, kèo phụ
- [ ] Champion prediction works for ASEAN Cup
- [ ] Admin can manually create matches and enter scores via the admin panel
- [ ] WC2026 pages remain accessible with all historical data intact (feature flag off)
- [ ] Shared wallet balance carries over from WC; no reset needed
- [ ] House P&L dashboard filters correctly by tournament
- [ ] Leaderboard is scoped per tournament
- [ ] Admin can toggle ASEAN Cup on/off without affecting WC flag
- [ ] Group standings table shows correct W/D/L/GD/Pts for ASEAN Cup groups
- [ ] Activity feed toasts appear when ASEAN Cup bets are placed
- [ ] Top-3 honor banner shows ASEAN Cup leaderboard top-3 (not WC top-3)

## Constraints & Assumptions

### Technical Constraints
- Must use `tournament_type` discriminator on existing `wc_*` tables — no new table prefix
- `wc_config` must become multi-row (one per tournament) — id=1 singleton pattern is retired
- All money arithmetic uses `shopspring/decimal` — no float64
- Route wiring: all new API routes go in `backend/internal/api/router.go`
- All new UI strings through `vue-i18n` — no hardcoded text
- Backend `PUT /admin/matches/:id` already accepts `home_score`, `away_score`, `status` as free-form JSON — only frontend form is missing

### Business Constraints
- WC2026 data must never be modified as a side effect of ASEAN Cup work
- All WC bets and kèo phụ are fully settled before ASEAN Cup launches (admin pre-condition)
- Same `wc_users` table — no new registration system needed
- One wallet per user, shared across all tournaments; ASEAN Cup starts with carry-over WC balance

### Confirmed Assumptions (resolved)
- **Match data source:** odds-api.io (current "statsapi") — unconfirmed for ASEAN Cup; football-data.org confirmed NOT covered. Admin manual entry is the primary path. API sync is a bonus if odds-api.io coverage is confirmed.
- **Score entry:** Admin manually enters match scores via a new "Enter Score" dialog in the admin panel; existing backend endpoint supports this already.
- **Knockout format:** ASEAN Cup SF and Final use 2 legs each. Each leg is a separate `wc_matches` row and bet independently. No aggregate-score betting.
- **Stages in use:** `group`, `sf`, `final` only — no `r32`, `r16`, `qf`, `third_place`.
- **Teams:** 10 teams across 2 groups of 5: Group A (Vietnam, Singapore, Indonesia, Cambodia, Timor-Leste), Group B (Thailand, Malaysia, Philippines, Myanmar, Laos).
- **WC flag:** admin-controllable; WC flag defaults off after all bets are settled; ASEAN Cup flag defaults off until admin enables it.
- **Group standings, analytics, activity feed, honor banner:** all confirmed in-scope for ASEAN Cup launch.

## Questions & Open Items

1. **odds-api.io ASEAN Cup coverage** — Verify by calling `GET /v3/leagues?sport=football&apiKey=<KEY>` and grepping for `aff` or `asean`. If found, add `acOddsApiLeague` constant and cron — score/odds sync becomes automatic and manual entry is no longer needed. *(Admin task, not blocking development.)*
2. **Champion prediction lock timing** — At what point should ASEAN Cup champion picks lock? Suggested: admin manually locks before first match (July 24 kickoff). Same mechanism as WC champion prediction.
3. **Timor-Leste home venue** — Timor-Leste plays "home" matches at Chonburi, Thailand (no FIFA-approved home stadium). The match card displays the venue string from `wc_matches.venue` — no special handling needed; admin just enters the correct venue when creating the match.
