---
phase: requirements
title: WC Custom Bet (Kèo Phụ)
description: Admin-defined proposition bets for WC matches — N options with custom odds, player bets on one option, admin settles by selecting the winner
---

# Requirements & Problem Understanding

## Problem Statement

The current WC betting system supports three fixed bet types (handicap, exact score, over/under) that are tied to structured match data (team scores, handicap lines, O/U lines). There is no way for the admin to create ad-hoc proposition bets — arbitrary questions about a match that players can wager points on.

Examples of bets players want to make:
- "Có 6 quả phạt góc trong trận này không?" (Yes/No)
- "Đội nào ghi bàn trước?" (Argentina / Brazil / Không ai)
- "Ronaldo có ghi bàn không?" (Có / Không)
- "Tổng số thẻ vàng?" (< 2 / 2–3 / > 3)

Without this feature, these side bets are tracked informally (group chat, spreadsheet), creating friction and no integration with the existing wallet system.

**Who is affected:** Admin (creates and settles bets), all WC players (place bets).

**Current workaround:** Side bets tracked in group chat with no automated settlement.

## Goals & Objectives

### Primary Goals
- Admin can create a custom bet for a specific WC match with a title and N options (each with a label and odds).
- Players can place a stake (in điểm) on one option per custom bet.
- After the match, admin selects the winning option; system automatically settles all entries (payout = stake × odds for winners, 0 for losers).
- Custom bets appear alongside other bet types on the match page.

### Secondary Goals
- Admin can open/close betting on a custom bet independently of match start.
- Admin can void a custom bet (all entries refunded).
- Settled custom bets appear in the player's bet history.
- House P&L dashboard includes custom bet settlement.

### Non-Goals
- Automatic result detection — admin always enters the result manually.
- Custom bets unlinked from a match (always match-scoped for V1).
- Player-created bets (admin only).
- Live (in-play) bets.

## User Stories & Use Cases

**Admin:**
- As an admin, I want to create a "Có 6 quả phạt góc?" bet for Argentina vs Brazil with options [Có (odds 1.8), Không (odds 2.0)] so that players can wager on it alongside the regular handicap.
- As an admin, I want to close betting before the match starts and settle with the correct answer after the match ends.
- As an admin, I want to void a custom bet if the match is cancelled or the bet question becomes irrelevant.

**Player:**
- As a player, I want to see all custom bets for a match on the match page and place a stake on my chosen option.
- As a player, I want to see my custom bet entries in my betting history with their result (win/lose/pending).
- As a player, I want to be able to cancel my custom bet entry before betting closes.

## Success Criteria

1. Admin creates a custom bet with 3 options; all 3 appear as selectable options on the match page for players.
2. Player bets 3 điểm on "Có"; admin settles with "Có" (odds 1.8) → player receives 3 × 1.8 = 5.4 điểm payout; losers receive 0.
3. Admin voids a custom bet → all player stakes are refunded to their wallets.
4. Custom bet entries appear in bet history with the same label/status UX as regular bets.
5. House P&L for a match includes custom bet net (sum of stakes − sum of payouts).

## Constraints & Assumptions

- Custom bets use the same wallet system (`wc_wallets`) as regular bets — same balance, same transaction log.
- Min/max stake per entry follows the same `wc_config.min_points` / `max_points` as regular bets.
- Each player can only place **one entry per custom bet** (but can change their chosen option before betting closes by cancelling and re-betting).
- A custom bet can have 2–10 options (admin-defined).
- Odds are set per option at creation time; admin can update odds before any player has bet on that option (or before close).
- Settlement selects exactly one winning option; push/half-win are out of scope for V1.
- Custom bets are separate from `wc_bets` and `wc_predictions` tables — new tables.

## Questions & Open Items

- Can admin update odds for an option after players have already bet on it? → Odds are snapshotted at bet time (same as existing bets), so admin can change future odds without affecting existing entries.
- Should custom bets appear in the same leaderboard P&L as regular bets? → Yes, same `wc_wallet_logs` pattern.
- Can a match have multiple custom bets simultaneously? → Yes, unlimited per match.
