---
phase: planning
title: Project Planning & Task Breakdown
description: Break down work into actionable tasks and estimate timeline
---

# Project Planning & Task Breakdown — FC25 Head-to-Head Popup

## Milestones
**What are the major checkpoints?**

- [x] Milestone 1: Backend endpoint returns correct H2H aggregation (opponents-only) with tests green. **(8 pure-logic tests pass; repo integration tests compile, skip without TEST_DATABASE_URL.)**
- [x] Milestone 2: Frontend modal renders stats + chart + recent list + form from live API.
- [x] Milestone 3: Dashboard entry point (header "So tài" button → modal with 2 selectors) wired, i18n (vi+en) complete, `vue-tsc` type-check clean.

> **Note (discovered during execution):** the `internal/service` **test package fails to compile on `main`** (pre-existing) — `wc_integration_test.go` + `wc_prediction_penalty_test.go` call `NewWcService` with a stale 4-arg signature (wants 5). Not caused by this feature; verified via clean-`main` stash. H2H service tests were run by temporarily patching those calls, then reverting. **Fixing those stale tests is a separate task.**

## Task Breakdown
**What specific work needs to be done?**

### Phase 1: Backend (foundation)
- [x] Task 1.0: **Inactive-player support** — add `UserRepository.GetByIDIncludingInactive` (used by the H2H service so soft-deleted players load, 404 only if the row truly doesn't exist) and expose an all-players list for the dropdown (`GetAllIncludingInactive` via `GET /users?include_inactive=true`, DTO carries `is_active`). Do **not** change existing callers' active-only default.
- [x] Task 1.1: Add DTOs (`H2HPlayer` incl. `is_active`, `H2HMatch`, `H2HStreak`, `HeadToHeadResponse`) and a lean row struct `H2HRow` (near `UserService` or in `model`).
- [x] Task 1.2: `MatchRepository.GetHeadToHeadMatches(p1, p2)` — the lean opposing-side join query, ordered by `match_date DESC` (drives aggregates). Plus `GetMatchesWithParticipants(ids)` (`Preload("Participants.User")`) for the top-10 recent lineups.
- [x] Task 1.3: `UserService.GetHeadToHead(p1, p2)` — validate `p1 != p2` (sentinel `ErrSamePlayer`), load both users via `GetByIDIncludingInactive` (404 only if a row truly doesn't exist), aggregate counts/win-rate/streak/form over the full set, cap form to 10, fetch lineups for the top-10 recent matches and map to `[]H2HMatch` with `participants`.
- [x] Task 1.4: Cache-aside wrap using `GetOrFetch` with key `esport:h2h:v{matchesVersion}:{p1}:{p2}`, TTL 5m. (No new invalidation code — version key already bumped on match writes.)
- [x] Task 1.5: `UserHandler.GetHeadToHead` — parse `player1`/`player2` query params, map sentinel errors → 400/404, else 200.
- [x] Task 1.6: Register route in `router.go` `/users` group, adjacent to `/leaderboard` (before `/:id`).

### Phase 2: Frontend core
- [x] Task 2.1: Add `HeadToHeadResponse` type (snake_case) to the core types file; add `userService.getHeadToHead(p1, p2)`.
- [x] Task 2.2: Build `src/components/user/HeadToHeadModal.vue` (el-dialog, props `player1Id`/`player2Id`/`v-model:visible`, fetch on open, loading/empty/error states).
- [x] Task 2.3: Stat row (total, per-player wins + win-rate %) + reuse `PlayerTierBadge.vue` in the header.
- [x] Task 2.4: Doughnut chart via `vue-chartjs` (register `ArcElement, Tooltip, Legend`), P1 wins vs P2 wins — mirror `components/wc/analytics/*`.
- [x] Task 2.5: Form badges (W/L, green/red, most-recent first) + recent-matches list showing full lineups grouped by team ("A+C vs B+D"), match-type tag, date, and highlighted winning side (mark P1/P2). No scoreline (schema has none).

### Phase 3: Integration & polish
- [x] Task 3.1: Dashboard "So tài đối đầu" card: two `el-select` listing **all** players incl. soft-deleted (inactive tagged "đã nghỉ" via `is_active`), compare button disabled until two distinct players chosen, opens modal. Source the full list from the new all-players endpoint.
- [x] Task 3.2: i18n keys `dashboard.headToHead.*` in `vi` + `en` locale files; no hardcoded strings.
- [x] Task 3.3: `npm run type-check` clean; `go test ./...` green; manual e2e smoke.

## Dependencies
**What needs to happen in what order?**

- 1.1 → 1.2 → 1.3 → 1.4 → 1.5 → 1.6 (backend chain).
- 2.1 depends on 1.5 (response shape finalized) — can stub type earlier from the design DTO.
- 2.2–2.5 depend on 2.1; 3.1 depends on 2.2.
- No external service or migration dependencies. Charts dependency (chart.js/vue-chartjs) already installed.

## Timeline & Estimates
**When will things be done?**

- Phase 1 (backend + tests): ~0.5 day.
- Phase 2 (modal + chart): ~0.5 day.
- Phase 3 (dashboard wiring + i18n + verify): ~0.25 day.
- Total: ~1.25 days including buffer.

## Risks & Mitigation
**What could go wrong?**

- **Route shadowing** (`/users/head-to-head` swallowed by `/users/:id`): register the static route before/adjacent to other static `/users` sub-routes; add a handler test hitting the path. Mitigation baked into Task 1.6.
- **Cache orientation bug** (A-vs-B vs B-vs-A sharing a slot): key keeps the ordered pair — documented in design; covered by a unit assertion that swapping players flips win counts.
- **Teammate leakage** (2v2/1v2 same-team pairs counted): prevented by the `team_number <>` join condition; explicit test with a teammate-only pair expecting `total_matches == 0`.
- **Streak/form off-by-one or wrong orientation**: unit test with a known sequence asserting form order (most-recent first) and streak owner/count.
- **Empty state**: service returns zeroed response (not error); modal shows friendly empty message.

## Resources Needed
**What do we need to succeed?**

- No new team members, tools, or infra. Existing Go + Vue toolchain.
- Seed data with mixed 1v1/2v2/1v2 matches between the same two players for manual verification.
- Reference: `src/components/wc/analytics/*` (chart pattern), `PlayerTierBadge.vue`, existing `el-dialog` components.
</content>
