---
phase: requirements
title: Refactored WC2026 — Player Filter, Schedule Time & Admin Page
description: Three focused enhancements to the existing WC2026 feature: player-filtered match history, correct VN timezone display + live highlight, and a dedicated admin route.
---

# Requirements & Problem Understanding

## Problem Statement

Three pain-points have emerged after the initial WC2026 release:

1. **`/matches` has no player filter.** Users can't quickly find matches involving a specific player — they have to scroll through the full list.
2. **WC schedule shows match times in the browser's local timezone.** Users viewing on machines set to UTC or other timezones see incorrect kickoff times. Additionally, live matches don't stand out visually — the current green border is too subtle.
3. **The admin panel is buried as a tab inside the betting view.** Any admin action (sync, settle, score entry) navigates away from the betting UI, and there is no dedicated admin-only entry point.

**Affected users:**
- Feature 1: All FC25 app users who review match history.
- Feature 2: All WC2026 schedule viewers (public page, no login required).
- Feature 3: WC admins.

**Current workaround:**
- Feature 1: Manually scan or use browser search on the full match list.
- Feature 2: Users mentally convert UTC times; live matches are easy to miss.
- Feature 3: Admin switches between the "Predictions" and "Admin" tabs inside `WcPredictView`.

---

## Goals & Objectives

### Primary goals
1. **Player filter on `/matches`**: Add a `?player_id=<uuid>` query param to `GET /api/v1/matches` (server-side) and a player dropdown in `MatchesView.vue` that passes the selected player ID.
2. **WC schedule VN time + live glow**: Force `timeZone: 'Asia/Ho_Chi_Minh'` on all `match_date` displays in `WcMatchCard.vue`, and add a glowing border + animated "LIVE" badge (distinct from the current subtle status badge) to cards with `status === 'live'`.
3. **Dedicated `/world-cup/admin` page**: Create `WcAdminView.vue` at route `/world-cup/admin` (admin WC auth required), move the full `WcAdminPanel` there, and remove the Admin tab from `WcPredictView`.

### Non-goals
- Changing how match data is stored or the match model itself.
- Adding new admin capabilities beyond what already exists in `WcAdminPanel`.
- Real-time timezone detection per user preference (UTC+7 is hardcoded for all).
- Player filter on any page other than `/matches`.

---

## User Stories & Use Cases

### Feature 1 — Player filter on /matches
- As a **team member**, I want to select a player from a dropdown so that only matches involving that player are shown.
- As a **team member**, when no player is selected the full match list loads as before (matches + bonuses merged).
- As a **team member**, the filter clears back to "All players" with a single click.
- As a **team member**, when a player is selected, the Total/Today/Locked stat cards are hidden to avoid showing misleading global counts.
- As a **team member**, score bonus entries remain visible in the feed regardless of which player is selected (bonuses are not player-filtered).

### Feature 2 — WC schedule VN time + live highlight
- As a **team member**, I want to see every WC kickoff time in Vietnam time (UTC+7) regardless of what timezone my device is set to.
- As a **team member**, I want live matches to be immediately obvious — a glowing green border around the card and a prominent pulsing "LIVE" chip — so I can find them at a glance without reading badges.

### Feature 3 — Dedicated /world-cup/admin page
- As a **WC admin**, I want to open `/world-cup/admin` to do all admin tasks (sync, score entry, settlement, feature toggle, top-ups) without being on the betting page.
- As a **WC admin**, I want the betting/prediction page (`/world-cup/predict`) to be purely for betting — no admin tab clutter.
- As a **WC admin**, when I'm on the WC schedule page and logged in, I want an "Admin" button/link in the page header so I can navigate to the admin page without typing the URL manually.
- As a **non-admin WC user**, if I navigate to `/world-cup/admin` I am redirected to the schedule page.

---

## Success Criteria

| Criterion | Target |
|---|---|
| Player dropdown filters server-side via `?player_id=` | Matches returned include only that player's matches; bonuses always appear |
| Filtering with no player selected returns full feed | Same result as current — matches + bonuses merged |
| Stat cards hidden when player filter is active | Total/Today/Locked cards not shown while filtering |
| WC match times displayed in Asia/Ho_Chi_Minh (UTC+7) | Correct for matches whose `match_date` is stored in UTC |
| Live match card has glowing border + prominent LIVE chip | Visually distinct from scheduled/completed cards |
| Admin link visible in WcScheduleView header for WC admins | Logged-in WC admin sees "Admin" link; non-admins and guests do not |
| `/world-cup/admin` accessible only to WC admins | Non-admin WC users and unauthenticated users are redirected |
| `WcPredictView` has no admin tab | Admin tab is fully removed from betting view |

---

## Constraints & Assumptions

### Technical constraints
- **No new tables**: Player filter uses the existing `match_participants` JOIN already implemented in `GetByUserID()`.
- **Same Go + Vue + Postgres stack**: no new services.
- **VN timezone is a constant**: `Asia/Ho_Chi_Minh` is hardcoded; no user preference needed.
- **WC admin auth**: `/world-cup/admin` must reuse the existing `requiresWcAuth` + `is_admin` check pattern.

### Assumptions
- `match_participants.user_id` correctly maps to `users.id` and GORM Preload works as in `GetByUserID()`.
- The football-data.org API stores `utcDate` in UTC; the current `match_date` column in `wc_matches` is stored in UTC and served as RFC3339/ISO8601.
- All three changes are self-contained; they do not affect each other's implementations.

---

## Questions & Open Items

| # | Question | Decision |
|---|---|---|
| 1 | Should the player filter dropdown include inactive players? | No — only active players (`is_active = true`). |
| 2 | Should `/world-cup/admin` redirect to login or schedule when unauthenticated? | Redirect to `/world-cup/login` if no WC token; redirect to `/world-cup` if token exists but user is not admin. The WC feature flag is NOT checked — admins must be able to reach this page even when the feature is disabled (e.g., to re-enable it). |
| 3 | Should the "LIVE" glow chip replace or supplement the existing status badge? | Supplement — existing badge remains; the glow is an additional card-level CSS class. |
| 4 | Pagination on filtered match list? | Keep the same page-size as the current unfiltered list; no additional pagination UI needed. |
| 5 | Client-side vs server-side player filter? | Server-side (`?player_id=` on `GET /api/v1/matches`). Bonus entries are not filtered — all bonuses appear regardless. Matches in the response are filtered by `match_participants.user_id`. |
| 6 | Should stat cards be hidden or recomputed when filter is active? | Hidden — collapse Total/Today/Locked stat cards when a player is selected. |
| 7 | How do WC admins discover `/world-cup/admin`? | An "Admin" button/link appears in the WcScheduleView page header for logged-in WC admins only. |
| 8 | Should bonuses be filtered when a player filter is active? | No — all bonuses appear regardless of the player filter. |
