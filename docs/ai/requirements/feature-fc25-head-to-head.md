---
phase: requirements
title: Requirements & Problem Understanding
description: Clarify the problem space, gather requirements, and define success criteria
---

# Requirements & Problem Understanding — FC25 Head-to-Head Popup

## Problem Statement
**What problem are we solving?**

- In the FC25 Tracker, players and organizers want to quickly see the **head-to-head record between any two players** — who beats whom, how often, and the recent trend.
- Currently there is no way to compare two specific players directly. The Dashboard/Users views show individual scores and tiers, but not the pairwise "so tài" (versus) record.
- Current workaround: manually scanning the match history and mentally counting — tedious and error-prone.

## Goals & Objectives
**What do we want to achieve?**

Primary goals:
- A **popup (modal)**, opened from the **Dashboard**, that compares two selected players' head-to-head record.
- Show, from Player A's perspective vs Player B:
  - **Total matches** played against each other.
  - **Win / loss counts and win-rate %** for each player.
  - A **visual chart** (doughnut) of the win split.
  - A **recent-matches list** of their encounters.
  - **Current streak / recent form** (W/L sequence).
- Head-to-head aggregates **all match types together** (1v1, 1v2, 2v2) — the UI does **not** break stats down by type.

Secondary goals:
- Reusable modal component so it can later be triggered from other views (e.g. Users page) with minimal work.
- Server-computed, cacheable stats (tiny payload).

Non-goals (explicitly out of scope):
- No per-type breakdown of stats (user explicitly said "không cần phân biệt type").
- No stats for matches where the two players were **teammates** — only opposing-side encounters count (see Constraints).
- No changes to WC2026 system, no new DB tables, no schema migration.
- No historical trend chart over time (only current form/streak).

## User Stories & Use Cases
**How will users interact with the solution?**

- As a player, I want to open a "So tài" panel on the Dashboard, pick two players, and see their head-to-head record so that I know my record against a specific rival.
- As an organizer, I want to compare any two players before a match to gauge the matchup.
- Key workflow:
  1. On Dashboard, user sees a "So tài đối đầu" card with two player dropdowns.
  2. User picks Player A and Player B, clicks the compare button.
  3. Modal opens showing totals, win/loss, chart, recent matches, and form.

Edge cases to consider:
- Two players have **never** faced each other on opposing sides → show a friendly "chưa có trận đối đầu nào" empty state.
- Players who have only ever been **teammates** (2v2/1v2) → counts as zero head-to-head matches.
- Same player selected twice → block / disable compare.
- 1v2 matches: solo player vs each duo member are opposing-side encounters; the two duo members are teammates (not counted against each other).
- Every FC25 match always has a decisive `winner_team` (1 or 2) → **no draws** to handle.
- Locked/settled matches are still counted (they are real results).
- Deleted/inactive players **are still selectable** — the dropdowns include soft-deleted players (marked "đã nghỉ") so historical matchups remain viewable. The backend endpoint must therefore accept inactive player IDs (404 only when an ID genuinely does not exist).

## Success Criteria
**How will we know when we're done?**

- Selecting two players who have opposing-side history shows correct totals: `player1_wins + player2_wins == total_matches`.
- Win rates are correct to the counts (draws not possible).
- Chart renders the win split and matches the numbers.
- Recent-matches list shows the most recent encounters with date, match type, and who won.
- Form/streak reflects the actual most-recent results from Player A's perspective.
- Empty state shows when there is no opposing-side history.
- All UI strings go through vue-i18n (vi primary, en parity).
- Backend unit tests cover: no-history, only-teammate-history, mixed types, streak computation.

## Constraints & Assumptions
**What limitations do we need to work within?**

Technical constraints:
- Backend: Go/Gin/GORM, Repository → Service → Handler layering (handlers never call repos).
- Frontend: Vue 3 `<script setup>` + Pinia + Element Plus + Tailwind; charts via already-installed `chart.js` + `vue-chartjs`.
- Money/points not involved here, but any counting must stay integer-correct.
- Reuse the existing matches cache-version counter for invalidation; no new invalidation plumbing.

Assumptions:
- **"Head-to-head" = matches where the two players were on *opposing* teams** (confirmed with user). Win/loss is measured from Player A's perspective; Player B's wins are simply the complement.
- All match types are aggregated; no type filter in the UI.
- Dashboard is the sole entry point for v1 (component built to be reusable later).

## Questions & Open Items
**What do we still need to clarify?**

- [Resolved] H2H definition → opponents-only (different `team_number`).
- [Resolved] Entry point → Dashboard, two dropdowns.
- [Resolved] Content → totals + win/loss + chart + recent matches + streak/form.
- [Resolved] Empty state (never met) → open the popup and show a friendly "chưa có trận đối đầu nào" message.
- [Resolved] Dropdowns list **all** players including soft-deleted (marked "đã nghỉ"), so historical matchups are viewable; endpoint accepts inactive IDs.
- [Default chosen] Recent-matches list length → 10 most recent (revisit if too short/long).
- [Default chosen] Form length → last 10 encounters, most recent first.
- [Default chosen] Dropdown sort order → by name (or score); to confirm during design.
</content>
</invoke>
