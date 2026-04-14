---
phase: requirements
title: FC25 Esport Score Tracker - Requirements
description: Quick scoring system for FC25 matches with debt settlement and fund management
feature: esport-score-tracker
created: 2026-04-14
---

# Requirements & Problem Understanding

## Problem Statement
**What problem are we solving?**

- **Problem:** Manually tracking FC25 match scores, calculating debt settlements, and managing a shared fund is time-consuming and error-prone
- **Who is affected:** 10-30 FC25 players who play daily matches and need to track wins/losses and money owed
- **Current situation:** Likely using pen and paper or spreadsheets, manually calculating who owes whom, and tracking fund contributions

## Goals & Objectives
**What do we want to achieve?**

### Primary Goals
- Quick match result entry for 1v1 and 2v2 matches
- Automatic point calculation (winner +1, loser -1)
- Automatic debt settlement when a player reaches the threshold (default: -6 points)
- Point-to-money conversion (configurable: 1 point = 22,000 VND)
- Fund management (50% of debt payments go to shared fund)
- Real-time leaderboard showing all players' current scores

### Secondary Goals
- View match history for audit trail
- Configure debt threshold and point conversion rate
- Track fund balance and usage

### Non-Goals (Out of Scope)
- User authentication/login (simple user list without passwords)
- Team/tournament management beyond basic matches
- Advanced statistics/analytics
- Mobile app (web responsive is sufficient)
- Integration with FC25 game API

## User Stories & Use Cases

### User Management
- **As an admin**, I want to add a new player with their name, so they can participate in matches
- **As an admin**, I want to edit a player's name if they want to change it
- **As an admin**, I want to delete a player who no longer participates
- **As a player**, I want to see my current score and debt status

### Match Recording
- **As a match recorder**, I want to create a 1v1 match result (Player A vs Player B) with the outcome, so points are automatically updated
- **As a match recorder**, I want to create a 2v2 match result (Team A vs Team B) with the outcome, so all 4 players' points are updated
- **As a match recorder**, I want to edit a match result if I made a mistake
- **As a match recorder**, I want to delete an incorrect match

### Debt Settlement System
- **As a player in debt**, when my score reaches -6 or below, I want the system to automatically trigger debt settlement
- **During debt settlement**:
  - 50% of the debt amount goes to the shared fund
  - 50% is distributed to the winning player(s) who caused the debt
  - My score is reset (added back to 0 or positive)
  - The winner's score is reduced by the amount paid to them
- **As an admin**, I want to see a history of all debt settlements

### Leaderboard & Reports
- **As a player**, I want to see a leaderboard ranked by current score
- **As a player**, I want to see how much money each score point represents (e.g., +5 points = +110,000 VND)
- **As an admin**, I want to see the current fund balance
- **As an admin**, I want to record fund expenditures (e.g., "Bought new controller - 500,000 VND")

### Edge Cases
- **Tie matches**: Should record as draw with no point changes (0/0)
- **Multiple players at -6**: Each settles independently
- **Negative fund balance**: Should warn but allow (debt paid via IOU)
- **Player deletion with outstanding matches**: Warn and require confirmation

## Success Criteria
**How will we know when we're done?**

### Functional Criteria
- ✅ Can manage (add/edit/delete) at least 30 users
- ✅ Can record 1v1 and 2v2 match results
- ✅ Points automatically update after each match
- ✅ Debt settlement triggers automatically at configurable threshold
- ✅ Debt payment splits correctly (50% fund, 50% to winner)
- ✅ Leaderboard updates in real-time
- ✅ Can view match history with date/time
- ✅ Can configure debt threshold and conversion rate

### Performance Benchmarks
- Page load < 2 seconds
- Match entry saves within 1 second
- Supports 30 concurrent users without lag
- Can handle 1000+ matches in history

### Usability Criteria
- Match entry takes < 30 seconds (quick workflow)
- Leaderboard visible on main page
- Mobile responsive (works on phones/tablets)
- Vietnamese language support for UI

## Constraints & Assumptions

### Technical Constraints
- Must work on modern browsers (Chrome, Firefox, Safari, Edge)
- No native mobile app required (web responsive is sufficient)
- Must be deployable to cloud for internet access

### Business Constraints
- Simple enough for non-technical admin to use
- No budget for paid services initially (use free tiers)
- Must support Vietnamese currency (VND) and language

### Time/Budget Constraints
- Target: 2-3 weeks for MVP
- Single developer
- Free hosting initially (Vercel + Supabase or similar)

### Assumptions
- Players trust the admin to record matches accurately
- Internet connection available during match entry
- No need for real-time notifications (refreshing page is acceptable)
- Debt settlement is voluntary enforcement (no payment integration)
- Fund withdrawals are manual (not tracked in system initially)

## Questions & Open Items

### Resolved
- ✅ Game type: FC25Resolved
- ✅ Point calculation: Winner +1, Loser -1
- ✅ Debt threshold: Configurable (default: -6)
- ✅ Currency: VND with configurable conversion (default: 1 point = 22,000 VND)
- ✅ User count: 10-30 players
- ✅ Authentication: Not needed

### Unresolved
- ❓ Should tied matches (draw) be recorded? → Suggest: Yes, record as 0/0 with no point change
- ❓ Can a match be edited after debt settlement? → Suggest: No, lock matches after settlement
- ❓ Who can record matches? → Suggest: Any player, but consider adding "recorder" name for audit
- ❓ Fund withdrawal tracking needed in MVP? → Suggest: Yes, simple add/subtract transactions
- ❓ Should we show negative fund balance? → Suggest: Yes, with warning indicator
- ❓ Delete player: What happens to their match history? → Suggest: Archive player instead of delete

### Items Requiring Stakeholder Input
- Confirm debt settlement flow is clear to all players
- Verify 50/50 split calculation is correct
- Confirm fund usage tracking requirements

### Research Needed
- Best practice for handling money calculations in JavaScript (avoid floating point errors)
- Optimal database schema for match history queries
- Vietnamese localization libraries for Vue 3
