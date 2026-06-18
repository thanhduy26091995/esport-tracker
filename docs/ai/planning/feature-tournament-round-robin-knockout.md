---
phase: planning
title: Round Robin + Top 4 Knockout Tournament — Planning
description: Task breakdown and implementation order for the round-robin group + top-4 knockout tournament format
feature: tournament-round-robin-knockout
created: 2026-06-18
---

# Project Planning & Task Breakdown

## Milestones

- [ ] M1: Backend data model and DB migration
- [ ] M2: Backend group-stage schedule generation + standings
- [ ] M3: Backend knockout generation and auto-create final/3rd-place
- [ ] M4: Frontend creation flow (format selector + team assignment)
- [ ] M5: Frontend detail view (standings table + knockout bracket)

## Task Breakdown

### Phase 1: Backend — Data Model & Migration

- [ ] **1.1** Write migration: add `format VARCHAR(30) NOT NULL DEFAULT 'classic'` to `tournaments`
- [ ] **1.2** Write migration: create `tournament_teams` table (id, tournament_id, player1_id, player2_id, created_at)
- [ ] **1.3** Write migration: add `stage VARCHAR(20) NOT NULL DEFAULT 'group'`, `team1_team_id UUID`, `team2_team_id UUID` to `tournament_matches`
- [ ] **1.4** Update `Tournament` struct: add `Format` field and `Teams []TournamentTeam` association
- [ ] **1.5** Create `TournamentTeam` Go struct with GORM tags
- [ ] **1.6** Update `TournamentMatch` struct: add `Stage`, `Team1TeamID`, `Team2TeamID` fields
- [ ] **1.7** Update `TournamentRepository`: AutoMigrate for new table; add `CreateTeams`, `GetTeamsByTournament`, `GetMatchesByStage` methods

### Phase 2: Backend — Group Stage Logic

- [ ] **2.1** Create `round_robin_teams.go`: implement `GenerateTeamSchedule(teams []TournamentTeam) []MatchSlot` using polygon rotation with ghost bye team; unit-test output for 5 teams (10 matches, each team 4 games, all pairs covered)
- [ ] **2.2** Implement `ComputeStandings(teams, matches) []TeamStanding` in `tournament_service.go`; sort by pts→GD→GF; assign seeds 1-4
- [ ] **2.3** Extend `CreateTournament` handler: detect `format="round_robin_top4"`, validate 10 player IDs, parse `teams` array (or auto-assign via existing tier-balanced logic), persist `TournamentTeam` rows, call `GenerateTeamSchedule`, persist group-stage matches with `stage="group"` + `team1_team_id`/`team2_team_id`
- [ ] **2.4** Extend `GetTournament` handler response: include `teams []TournamentTeam` and `standings []TeamStanding` arrays (only populated for `round_robin_top4`)

### Phase 3: Backend — Knockout Logic

- [ ] **3.1** Implement `GenerateKnockouts(tournamentID)` in `tournament_service.go`: validate all group matches completed + no existing knockout matches; call `ComputeStandings`; create 2 semi `TournamentMatch` rows (`stage="semi"`, `round=6`)
- [ ] **3.2** Wire `POST /api/v1/tournaments/:id/generate-knockouts` in `tournament_handler.go` and `router.go`
- [ ] **3.3** Implement `maybeCreateFinalMatches(tx, tournamentID)` helper: called after recording any semi result; if both semis completed, creates `stage="final"` (winner vs winner) and `stage="third_place"` (loser vs loser) with `round=7`
- [ ] **3.4** Call `maybeCreateFinalMatches` from the existing `RecordResult` handler when `stage="semi"`

### Phase 4: Frontend — Create Flow

- [ ] **4.1** Add `format` field to tournament TypeScript types; add `TournamentTeam` and `TeamStanding` interfaces to `tournament.ts`
- [ ] **4.2** Extend `tournamentService.ts`: pass `format` + `teams` in `createTournament()`; add `generateKnockouts(tournamentId)` API call
- [ ] **4.3** Update `CreateTournamentView.vue`: add format selector radio/toggle ("Classic" | "Round Robin + Top 4"); conditionally show team-assignment step for `round_robin_top4` (5 row player-pair dropdowns)
- [ ] **4.4** Validate in `CreateTournamentView.vue`: exactly 10 players required for `round_robin_top4`; each player used in exactly one team; no duplicate players across teams
- [ ] **4.5** Create `TournamentFormatBadge.vue`: small inline chip displayed on list and detail views

### Phase 5: Frontend — Detail View

- [ ] **5.1** Create `TournamentStandingsTable.vue`: columns Pos / Team (player1 + player2 names) / P / W / D / L / GF / GA / GD / Pts; top-4 rows visually highlighted (green tint or trophy icon); shows "—" for groups with no results yet
- [ ] **5.2** Create `TournamentKnockoutBracket.vue`: 4-slot bracket layout — Semi1 (seed1 vs seed4), Semi2 (seed2 vs seed3), 3rd-place, Final; pending matches show "TBD"; completed matches show score + winner highlighted
- [ ] **5.3** Update `TournamentDetailView.vue`: for `round_robin_top4`, render `TournamentStandingsTable` above matches; add "Group Stage" and "Knockout" tabs (or sections) to organise matches; add "Generate Knockouts" button (visible when all group matches done and no knockout matches exist yet)
- [ ] **5.4** Add `TournamentKnockoutBracket` to `TournamentDetailView` in the Knockout section; call `generateKnockouts()` on button click and refresh detail
- [ ] **5.5** Add `TournamentFormatBadge` to `TournamentsView` list rows and `TournamentDetailView` header

## Dependencies

```
Phase 1 → Phase 2 (models must exist before service logic)
Phase 2 → Phase 3 (standings used in knockout generation)
Phase 1 → Phase 4 (format field needed in TS types)
Phase 2+3 → Phase 5 (standings + knockout data must exist in API response)
Phase 4 and Phase 5 are independent (can be parallelised)
```

## Timeline & Estimates

| Phase | Description | Est. Effort |
|-------|-------------|-------------|
| 1 | Backend data model + migration | ~1 h |
| 2 | Backend group stage logic | ~2 h |
| 3 | Backend knockout logic | ~1.5 h |
| 4 | Frontend create flow | ~2 h |
| 5 | Frontend detail view | ~3 h |
| **Total** | | **~9.5 h** |

## Risks & Mitigation

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Polygon-rotation schedule has off-by-one in bye handling | Medium | Unit-test `GenerateTeamSchedule` for 5 teams — assert 10 matches, all pairs exactly once |
| Standings tie: 2 teams equal on pts+GD+GF, wrong team advances | Low | Document v1 limitation; admin can manually reorder if needed (head-to-head deferred) |
| `maybeCreateFinalMatches` called twice (race or double-submit) | Low | Check match count for stage="final" before inserting; idempotent guard |
| Team assignment UX with dropdowns for 10 players is tedious | Medium | Use simple per-slot dropdown pairs for v1; drag-drop is a v2 enhancement |
| Classic tournament creation broken by format validation | Medium | Default `format` to `"classic"` on backend; frontend omits field for classic → zero change to existing flow |

## Resources Needed

- `docs/ai/knowledge/tournament-system.md` — existing patterns and existing schema
- `docs/ai/knowledge/backend-patterns.md` — repository/service conventions
- `docs/ai/knowledge/frontend-patterns.md` — Pinia store, service layer patterns
- No new external dependencies
