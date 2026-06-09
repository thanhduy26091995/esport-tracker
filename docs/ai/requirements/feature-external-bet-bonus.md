---
phase: requirements
title: "External Bet Bonus Points"
description: Allow awarding bonus score points to a registered player for winning an external (off-system) bet, with the equivalent money amount deposited into the fund
---

# Requirements & Problem Understanding

## Problem Statement

Players sometimes make bets outside the game system (e.g., side wagers, friendly bets). When the loser of such a bet pays into the group fund, the winner should also be credited score points — but there is no opponent registered in the system and no match to record.

**Example scenario:**  
Cookie (not a registered player, or simply an external person) loses a side bet to Player A. Cookie deposits the equivalent of 3 points into the fund. Player A should receive +3 score points, and the fund should record a deposit of 3 × point_value (in VND).

Current workaround: admins manually adjust scores via SQL or skip the score credit entirely — neither is auditable.

## Goals & Objectives

**Primary goals:**
- Provide a dedicated UI action and API endpoint to record an external bet win
- Award a specified number of bonus points to a single registered player
- Create an auditable record with a description (e.g., "Cookie thua cược – 3 điểm")
- Show the bonus as a special entry in the global `MatchesView`, mixed with real matches, with label `[Cược]`

**Non-goals:**
- Auto-creating a fund deposit (the fund is topped up manually via the existing Deposit feature on the Fund page)
- Recording the external loser as a player in the system
- Creating a `Match` record for this transaction
- Supporting multi-winner payouts (one recipient per transaction)

## User Stories & Use Cases

- **As an admin**, I want to record that Player A won an external bet worth 3 points so that their score is updated with an auditable description.
- **As an admin**, I want to provide a description (e.g., "Cookie thua cược – 3 điểm") so the bonus is traceable.
- **As a player**, I want to see my score increase and a `[Cược]` entry appear in the Matches page after my external bet win is recorded.
- **As an admin**, I will separately deposit the equivalent money into the fund using the existing Fund → Deposit feature.

**Edge cases:**
- Points must be a positive integer (no negative or zero awards)
- The player (winner) must be a registered user in the system
- The action is reversible: deleting the bonus record reverts the player's score and removes the history entry

## Success Criteria

- `POST /api/v1/score-bonuses` accepts `{ user_id, points, description }` and returns the created record
- The target player's `current_score` increases by the specified points
- The bonus appears in `MatchesView` mixed with real matches, labelled `[Cược]` with the description and `+N pts` for the winner
- A "Cộng điểm cá cược" action is accessible from the UI (Fund page or Dashboard)
- Deleting a bonus reverts the player's `current_score` and removes it from `MatchesView`
- No fund transaction is created automatically — fund is managed separately via the existing Deposit flow

## Constraints & Assumptions

- Only one player (the winner) is credited per bonus record
- No `Match` record is created — this is a separate `score_bonuses` entity
- No fund deposit is auto-created; fund balance is updated manually via the existing Deposit feature
- Tier recalculation runs after the bonus is applied (same hook as after a match, non-fatal)
- There is no role-based auth in the system — any user can record a bonus
- The bonus history label will be "Bet Bonus" (or "Thưởng cược") in the player's score history

## Questions & Open Items

- ~~What is the config key name for the money value of 1 point?~~ → Resolved: no fund auto-deposit needed
- ~~Should bonus points appear in the player's history?~~ → Resolved: Yes, as a special "Bet Bonus" entry
- ~~Should bonus be deletable?~~ → Resolved: Yes, with full score revert
- Where exactly in the player profile is the bonus entry shown — merged into match history timeline, or a separate section?
