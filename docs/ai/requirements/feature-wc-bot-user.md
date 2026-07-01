---
phase: requirements
title: WC Bot User Flag — Requirements
description: Add is_bot flag to wc_users; bots appear on leaderboard with tag but are excluded from honor banner top-3
---

# Requirements & Problem Understanding

## Problem Statement

The friend group uses bot accounts to simulate activity or hold placeholder positions. Currently there is no way to distinguish bot accounts from real users, so:
- The honor banner (top-3 marquee) may celebrate a bot instead of a real player.
- Admins cannot visually tell which accounts are bots in the admin panel.

## Goals & Objectives

**Primary**
- Add `is_bot` boolean flag to `wc_users`.
- Admin can toggle the flag for any user via the admin panel.
- Leaderboard tab shows bots with a visible "Bot" badge.
- Honor banner (top-3 marquee) always shows 3 real users — bots are skipped and the list extends until 3 non-bot entries are found.

**Non-goals**
- Bots do not behave differently from real users (wallet, predictions, bets all unchanged).
- No automated bot behaviour or scripting.
- No separate bot-only endpoints.

## User Stories & Use Cases

- As an **admin**, I want to mark an account as a bot so that it is excluded from the honor banner.
- As an **admin**, I want to unmark a bot so that it becomes a regular user again.
- As a **viewer** of the honor banner, I want to see the top 3 real players, not bots.
- As a **leaderboard viewer**, I still want to see bots ranked with their actual scores, but clearly labelled.

**Edge cases**
- All top 3 are bots → banner extends to positions 4, 5, 6 (up to 3 real users found).
- Fewer than 3 real users exist → banner shows however many real users are available (1 or 2).
- Bot is ranked outside top 3 → no change to banner behaviour.

## Success Criteria

- `is_bot` column added via idempotent migration (`ADD COLUMN IF NOT EXISTS`).
- Admin can toggle `is_bot` from the admin panel user list.
- `GET /wc/leaderboard` returns `is_bot` on each entry.
- `WcTop3Banner` displays exactly the top-3 non-bot entries (or fewer if <3 real users exist).
- Leaderboard tab shows a "Bot" badge next to bot entries.

## Constraints & Assumptions

- `is_bot` defaults to `false` — no existing accounts are affected.
- The leaderboard query already returns all users ranked by `net_points`; we just need to add `is_bot` to the projection.
- The banner fetches the full leaderboard from `wcStore.leaderboard`; filtering happens client-side — no extra API call needed.

## Questions & Open Items

- None — all clarified in conversation.
