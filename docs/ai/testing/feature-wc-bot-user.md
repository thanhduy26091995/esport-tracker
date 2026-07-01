---
phase: testing
title: WC Bot User Flag — Testing Strategy
description: Test cases for is_bot flag, leaderboard projection, and banner filtering
---

# Testing Strategy

## Unit Tests

### `computeTop3Real` (banner logic)

- [ ] All 3 top entries are real → returns positions 1, 2, 3 unchanged
- [ ] Position 1 is bot → skips it, returns positions 2, 3, 4
- [ ] Positions 1 and 2 are bots → returns positions 3, 4, 5
- [ ] All top 3 are bots → returns positions 4, 5, 6
- [ ] Fewer than 3 real users exist → returns however many are available (1 or 2)
- [ ] Empty leaderboard → returns empty array

### `SetUserBot` service

- [ ] Setting `is_bot = true` persists to DB
- [ ] Setting `is_bot = false` reverts the flag
- [ ] Non-existent user returns error

## Integration Tests

- [ ] `GET /wc/leaderboard` — response includes `is_bot` on each entry
- [ ] `PUT /admin/users/:id/set-bot` — non-admin returns 403
- [ ] `PUT /admin/users/:id/set-bot` — admin toggles flag, subsequent leaderboard fetch reflects change
- [ ] `GET /wc/leaderboard` — newly marked bot has `is_bot: true`

## Manual Testing Checklist

- [ ] Mark a top-1 user as bot via admin panel → honor banner now shows positions 2, 3, 4 as top 1, 2, 3
- [ ] Mark top 1, 2, 3 as bots → banner shows positions 4, 5, 6
- [ ] Unmark bot → banner reverts to include that user
- [ ] Leaderboard tab: bot user has "Bot" badge
- [ ] Bot user can still place predictions and earn points normally

## Test Data

- Create 6+ users with varying point balances
- Mark users at positions 1 and 3 as bots
- Verify banner shows users at positions 2, 4, 5 as top 1, 2, 3
