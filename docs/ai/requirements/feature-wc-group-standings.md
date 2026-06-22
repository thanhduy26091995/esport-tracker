---
phase: requirements
title: WC Group Standings Table
description: Show live group-stage standings (W/D/L, GF, GA, GD, Points, Form) on the /world-cup schedule page
---

# Requirements & Problem Understanding

## Problem Statement

On the `/world-cup` schedule page, visitors can see the match schedule but have no way to know the current group standings. This makes it hard to follow which teams are on track to qualify for the knockout rounds — the core reason people watch group-stage football.

**Affected users**: Everyone viewing the WC schedule, authenticated or not.

**Current workaround**: Users have to go to an external site (FIFA/ESPN) to check standings, which is friction.

## Goals & Objectives

**Primary goals**:
- Display current group standings for each WC group on the `/world-cup` page
- Show enough stats to understand team position: Played, W/D/L, GF, GA, GD, Points
- Show last-5 form to indicate recent momentum
- Works for unauthenticated visitors (schedule page is public)

**Secondary goals**:
- Standings update automatically as match results are synced
- Visible in context — shown for the currently-selected group filter

**Non-goals**:
- Head-to-head tiebreaker details (show only GD and GF as tiebreakers)
- Knockout stage bracket standings
- Historical standings from previous tournaments
- Separate dedicated standings page (embed in existing schedule page)

## User Stories & Use Cases

- As a **visitor**, I want to see Group A's standings table when I select "Group A" in the filter, so I can tell which teams are leading the group.
- As a **visitor**, I want to see each team's W/D/L record and goal stats, so I can understand why they're in their current position.
- As a **visitor**, I want to see the last 5 match results (form) as coloured badges (W/D/L), so I can gauge each team's recent momentum.
- As a **visitor**, after all 3 group-stage rounds are complete, I want to see the final standings to confirm which 2 teams from each group advance.

**Key workflows**:
1. User visits `/world-cup` with default filter ("All matches") → sees all 12 group standings in a **2-column grid** above the match list
2. User selects "Group A" from the filter → only Group A's standings table is shown above Group A's matches
3. Table shows rank, flag/code, team name, P/W/D/L, GF/GA/GD, Points, and last-5 form
4. Row highlight: **rows 1–2 green** (direct qualification), **row 3 yellow** (may qualify as one of the 8 best third-place teams — WC2026 rule), row 4 no highlight

**Edge cases**:
- Group with no completed matches yet → show all 4 teams with zeroed stats. Team roster is built from **all** group-stage matches (any status: scheduled/live/completed), not just completed ones.
- Team with fewer than 5 completed matches → show form only for matches played (1–4 badges)
- Ties on points/GD/GF → sort alphabetically as a stable fallback
- Live match in progress → its score is not yet completed; don't count it in standings stats

## Success Criteria

- [ ] Standings table appears when a group filter (e.g., "Group A") is selected
- [ ] All 48 WC teams appear across 12 groups (4 teams per group in WC2026)
- [ ] Points, GD, and form are correct after match results are synced
- [ ] Rows 1–2 have a green highlight (direct qualification); row 3 has a yellow hint (potential best-3rd qualification)
- [ ] Form badges are correct colour (W=green, D=grey, L=red)
- [ ] Works on mobile (responsive table or card layout)
- [ ] Accessible without login (public endpoint)

## Constraints & Assumptions

**Technical constraints**:
- Must compute standings purely from existing `wc_matches` table — no new database tables
- Standings are only meaningful for `stage = 'group'` matches
- Only count matches with `status = 'completed'` in the standings (not `live` or `scheduled`)
- Must follow the WC isolation principle: no cross-table writes between WC and core esport

**Assumptions**:
- `wc_matches` is kept in sync via the StatsAPI cron (already running)
- WC2026 group stage has 12 groups (A–L), 4 teams each, 3 matches per team
- The backend endpoint is public (`/api/v1/wc/standings`) — no auth required

## Questions & Open Items (All Resolved)

- **Q: Show standings for ALL groups at once, or only the selected group?**
  A: **Both** — when no group filter is selected ("All matches"), show all 12 groups in a 2-column grid. When a specific group is selected (e.g., "Group A"), show only that group's table.

- **Q: WC2026 has 12 groups — confirm group naming convention in the DB?**
  A: `group_name` field in `wc_matches`, e.g. `"Group A"` through `"Group L"`. The filter component already uses this convention.

- **Q: Should standings be shown on the knockout stage views (r16, qf, etc.)?**
  A: No — standings only make sense for the group stage. When a knockout stage filter is active, no standings are shown.

- **Q: Should form include the CURRENT live match?**
  A: No — only `completed` matches count.

- **Q: Pre-tournament — how to show 4 teams when no matches are completed?**
  A: Build the team roster from all group-stage matches (any status), but count stats from completed matches only. Show all 4 teams with zeroed stats.

- **Q: Row highlight — how to handle WC2026 third-place qualification?**
  A: Rows 1–2 = green (direct). Row 3 = yellow (potentially qualifies as one of the 8 best third-place teams). Row 4 = no highlight.
