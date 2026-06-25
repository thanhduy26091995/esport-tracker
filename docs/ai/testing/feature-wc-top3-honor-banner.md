---
phase: testing
title: WC Top-3 Honor Banner — Testing Strategy
description: Manual and integration test plan for the animated top-3 honor banner
---

# Testing Strategy

## Test Coverage Goals

- Unit tests: not required (pure presentational components with no business logic)
- Integration: verify store data flows into banner correctly
- Manual: visual + behavioral checklist below

## Manual Testing Checklist

### Core Rendering

- [ ] Banner appears on `/wc/predict` when logged in
- [ ] Banner appears on `/wc/schedule` (authenticated view) when logged in
- [ ] Banner does NOT appear on the public `/wc/schedule` (unauthenticated)
- [ ] Banner does NOT appear on `/wc/login`
- [ ] Banner shows 🥇🥈🥉 medals for positions 1/2/3
- [ ] Each card shows: avatar (or default avatar), name (truncated if > 16 chars), net_points with +/- sign and correct color (green = positive, red = negative)
- [ ] Banner is hidden when leaderboard returns 0 entries

### Animation

- [ ] Cards scroll continuously left in a loop without visible jump/reset
- [ ] Hover pauses the animation
- [ ] Mouse-out resumes the animation
- [ ] Animation loops smoothly (no gap between last and first card)

### Current-User Highlight

- [ ] If logged-in user is in top 3, their card has a distinct gold ring/glow
- [ ] If logged-in user is NOT in top 3, no special highlight on any card

### Data Refresh

- [ ] Banner data updates after 5 minutes (simulate by manually calling `wcStore.fetchLeaderboard()` in console)
- [ ] After a new bet is settled and leaderboard changes, banner reflects new order within next refresh cycle

### Responsive / Mobile

- [ ] Banner is legible on 375px wide viewport (iPhone SE)
- [ ] Banner height is consistent (does not cause layout shift) at 56px on mobile and desktop
- [ ] Text does not overflow cards on mobile

### Accessibility

- [ ] `prefers-reduced-motion: reduce` → animation stops, cards displayed statically
- [ ] Banner region has `aria-label`
- [ ] Screen reader can announce top-3 player names and scores

### Edge Cases

- [ ] Only 1 or 2 players in leaderboard — banner shows what is available (no crash)
- [ ] Very long player name (20+ chars) — truncated with ellipsis
- [ ] Negative net_points — displayed as `-12.5` in red
- [ ] Zero net_points — displayed as `+0` in green (or neutral)
- [ ] Avatar URL fails to load — falls back to default SVG avatar

## Test Data

Use existing dev/staging WC users. Ensure at least 3 WC users have settled bets to populate leaderboard.

## Regression Checklist

- [ ] Existing `WcLeaderboard.vue` component still works correctly on the predict page
- [ ] No duplicate API calls to `/wc/leaderboard` on initial page load (banner should reuse already-fetched store data)
- [ ] WC betting flow (place bet, cancel bet) unaffected
- [ ] Page scroll not hijacked by banner animation

## Manual Sign-Off

- [ ] Tested on Chrome desktop
- [ ] Tested on Chrome mobile (DevTools device emulation or real device)
- [ ] No console errors during normal usage
