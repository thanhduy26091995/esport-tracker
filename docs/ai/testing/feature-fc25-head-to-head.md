---
phase: testing
title: Testing Strategy
description: Define testing approach, test cases, and quality assurance
---

# Testing Strategy — FC25 Head-to-Head Popup

## Test Coverage Goals
**What level of testing do we aim for?**

- Unit-test all new backend aggregation logic (service) — 100% of new branches.
- Repository query verified via the service tests against a test DB (or a seeded fixture).
- Frontend: type-check clean + manual UI verification (project has no component test harness for core views).
- Alignment: every requirement acceptance criterion maps to a test below.

## Unit Tests
**What individual components need testing?**

### UserService.GetHeadToHead
- [ ] Test case: **no shared history** → `total_matches == 0`, both win counts 0, win rates 0, empty form/recent, streak `PlayerID == nil`.
- [ ] Test case: **teammate-only history** (2v2/1v2 same team every time) → `total_matches == 0` (opponents-only proven).
- [ ] Test case: **mixed types** (1v1 + 2v2 + 1v2 opposing) → `player1_wins + player2_wins == total_matches`, aggregated across types.
- [ ] Test case: **win-rate math** → rates equal counts/total, sum ≈ 1.0 (no draws).
- [ ] Test case: **orientation** → swapping `(p1,p2)` → `(p2,p1)` flips win counts and form W↔L; totals identical.
- [ ] Test case: **form order** → most-recent first, capped at 10.
- [ ] Test case: **streak** → leading run of identical results; correct owner (P1 on "W", P2 on "L") and count.
- [ ] Test case: **same player** (`p1 == p2`) → `ErrSamePlayer`.
- [ ] Test case: **truly-missing player** (nonexistent UUID) → not-found error.
- [ ] Test case: **soft-deleted player** (`is_active = false`) → loads fine, `H2HPlayer.is_active == false`, stats returned (NOT a 404).
- [ ] Additional coverage: recent list capped at 10, ordered date-DESC.
- [ ] Test case: **lineups** — for a 2v2/1v2 recent match, `H2HMatch.participants` includes all players with correct `team` numbers and names (both sides present).

### MatchRepository.GetHeadToHeadMatches
- [ ] Test case: returns only rows where the two users are on different `team_number`.
- [ ] Test case: excludes matches where only one of the two participated.
- [ ] Test case: rows ordered `match_date DESC`; `p1_team` and `winner_team` populated.

### UserHandler.GetHeadToHead
- [ ] Test case: valid params → 200 + JSON body.
- [ ] Test case: `player1 == player2` → 400.
- [ ] Test case: invalid UUID → 400.
- [ ] Test case: unknown player → 404.
- [ ] Test case: **route not shadowed** by `/users/:id` (path resolves to the H2H handler).

## Integration Tests
**How do we test component interactions?**

- [ ] Seed matches (1v1/2v2/1v2, opposing + teammate) between two players → call endpoint → assert full response shape and counts.
- [ ] Cache behavior: second call served from cache; after a new match write (version bump) the next call reflects the new match.
- [ ] API endpoint smoke via the running server.

## End-to-End Tests
**What user flows need validation?**

- [ ] Dashboard → pick two players → open modal → stats/chart/list/form render correctly.
- [ ] Empty state: pick two players who never met → friendly empty message, no chart error.
- [ ] Compare button disabled until two distinct players selected.
- [ ] Regression: Dashboard's existing widgets and other `/users/*` endpoints unaffected.

## Test Data
**What data do we use for testing?**

- Fixtures: two users with a mix of opposing-side and teammate matches across all three types, plus a decisive winner per match.
- A third/fourth user to build 2v2/1v2 teams.

## Test Reporting & Coverage
**How do we verify and communicate test results?**

- Backend: `cd backend && go test ./...` (add `-cover` for the service/repository packages).
- Frontend: `cd frontend && npm run type-check`.
- Record manual UI sign-off (screenshots of populated + empty states).

## Manual Testing
**What requires human validation?**

- Doughnut colors/legend readable; numbers match chart.
- Form badges color-correct (W green / L red), most-recent first.
- i18n: switch vi/en, confirm no hardcoded strings leak.
- Responsive: modal usable on mobile width.

## Performance Testing
**How do we validate performance?**

- Not load-critical (friend-group scale). Sanity-check the endpoint responds < 100ms warm-cache; single join cold.

## Bug Tracking
**How do we manage issues?**

- Track defects against the acceptance criteria in the requirements doc; add a regression test for any bug found before closing.
</content>
