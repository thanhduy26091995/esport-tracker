---
phase: requirements
title: WC Top-3 Honor Banner — Requirements
description: Continuously animated banner on WC pages celebrating the top 3 leaderboard players to boost engagement and return rate
---

# Requirements & Problem Understanding

## Problem Statement

The World Cup betting page has a leaderboard section, but it's static and buried — users who are not already on the leaderboard tab never notice the standings. There is no persistent social signal that keeps the competition top-of-mind.

- **Who is affected:** All WC players
- **Current situation:** Leaderboard is a scrollable list on the predict/schedule page — visible only when explicitly navigated to
- **Desired outcome:** A continuously running, always-visible banner that celebrates the top 3 players and creates ambient competitive pressure, increasing both return visits and betting volume

## Goals & Objectives

**Primary goals:**
- Show top 3 players (rank 🥇🥈🥉 + name + net points) in a continuously animating banner visible on all WC pages
- Refresh data automatically so it stays current without a page reload
- Increase the sense of competition and motivation to climb the leaderboard

**Secondary goals:**
- Highlight the current logged-in user if they appear in top 3 (subtle "you are here" signal)
- Keep the banner visually distinct but non-intrusive — must not block betting controls

**Non-goals:**
- Changes to the leaderboard API or backend (reuse `GET /wc/leaderboard`)
- Full podium page or separate route
- Push notifications for rank changes
- Animation beyond CSS (no canvas/WebGL)

## User Stories & Use Cases

- As a **WC player**, I want to glance at the top 3 at any time while betting, so I know how close I am to the leaders without navigating away.
- As a **top-3 player**, I want my name to be visibly celebrated, so I feel recognized and motivated to stay on top.
- As a **player outside top 3**, I want to see the gap to #3, so I know exactly how much I need to earn to break into the podium.
- As a **casual returner**, I want to immediately see who is winning when I open the WC page, so the competition grabs my attention.

## Success Criteria

- Banner is visible on all WC authenticated pages (predict, schedule/upcoming, wallet, custom bet)
- Banner auto-scrolls through top 3 continuously (loops when reaching the end)
- Data refreshes every 5 minutes (same cadence as the upcoming-matches widget)
- On mobile (< 640 px) the banner is still legible and not cut off
- Banner highlights the logged-in user's card if they are in top 3
- No layout shift — banner occupies a fixed height slot in the page

## Constraints & Assumptions

- **Technical:** Reuse `GET /api/v1/wc/leaderboard` — no new backend work
- **Reuse:** Leaderboard API already returns `net_points`, `name`, `avatar_url`, `wc_user_id`
- **i18n:** All new strings must use `vue-i18n` (no hardcoded text)
- **Auth:** Banner only shown when the user is logged in (WC auth required to view P&L)
- **Performance:** Banner must not block the main thread; use CSS `animation` not JS timers for scroll motion
- **Assumption:** Leaderboard always returns entries sorted by `net_points` descending

## Questions & Open Items

- Should the banner appear on public (unauthenticated) WC pages like `/wc/schedule`? → Assume **no** for now; authenticated-only pages first.
- How many characters can a player name be? (impacts marquee layout) → Cap display name at 16 chars with ellipsis.
- Should clicking a banner card navigate anywhere? → No navigation for now — banner is purely decorative/informational.
