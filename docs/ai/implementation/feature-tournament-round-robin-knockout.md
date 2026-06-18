---
phase: implementation
title: Round Robin + Top 4 Knockout Tournament — Implementation Guide
description: Technical notes, patterns, and code guidelines for implementing the round-robin top-4 knockout tournament format
feature: tournament-round-robin-knockout
created: 2026-06-18
---

# Implementation Guide

## Development Setup

No new dependencies required. Uses the existing Go/GORM/PostgreSQL backend and Vue 3/TypeScript/Pinia frontend stack.

```bash
# Backend
cd backend && go run cmd/server/main.go

# Frontend
cd frontend && npm run dev
cd frontend && npm run type-check
```

## Code Structure

```
backend/internal/
  model/tournament.go              ← TournamentTeam struct; Format on Tournament; Stage + TeamIDs on TournamentMatch
  repository/tournament_repository.go  ← CreateTeams, GetTeamsByTournament, GetMatchesByStage
  service/
    tournament_service.go          ← CreateRoundRobinTop4, ComputeStandings, GenerateKnockouts, maybeCreateFinalMatches
    round_robin_teams.go           ← GenerateTeamSchedule (polygon rotation)
  api/
    tournament_handler.go          ← GenerateKnockoutsHandler; extended Create + Get handlers
    router.go                      ← POST /tournaments/:id/generate-knockouts

frontend/src/
  types/tournament.ts              ← TournamentTeam, TeamStanding, stage enum
  services/tournamentService.ts    ← generateKnockouts()
  views/
    CreateTournamentView.vue       ← format selector + team assignment step
    TournamentDetailView.vue       ← standings + bracket integration
  components/
    TournamentStandingsTable.vue   ← group standings table
    TournamentKnockoutBracket.vue  ← 4-team knockout bracket
    TournamentFormatBadge.vue      ← format chip
```

## Implementation Notes

### Core Features

**Format discriminator:**
- Add `Format string` to `Tournament` Go struct and update `CreateTournamentRequest` DTO
- All existing handlers check `t.Format == "round_robin_top4"` before calling new logic; otherwise fall through to existing classic path
- Default value `"classic"` ensures zero migration cost for existing rows

**GenerateTeamSchedule (polygon rotation):**
- Input: `[]TournamentTeam` of length 5 (odd)
- Internally appends a ghost `TournamentTeam{}` (zero UUID) to make length 6
- Fixes slot[0], rotates slots[1..5] left by one each round
- Skips any matchup involving the ghost team (the bye)
- Output: 10 `MatchSlot` structs with `Round`, `Order`, `Team1`, `Team2`
- This is pure in-memory logic with no DB access — place in `round_robin_teams.go`

**ComputeStandings:**
- Only considers `stage == "group"` and `status == "completed"` matches
- Builds a `map[uuid.UUID]*TeamStanding` keyed on `team_id`
- Goals from `actual_score1` / `actual_score2` (raw goals, not effective)
- W/D/L determined by `effective_winner` (0=draw, 1=team1 wins, 2=team2 wins)
- After aggregation, sort slice by pts DESC → gd DESC → gf DESC
- Assign seeds 1–4 to the top 4 entries; others get seed 0

**GenerateKnockouts:**
- Called by `GenerateKnockoutsHandler`
- Validates: all 10 group matches completed; no existing knockout matches in this tournament
- Calls `ComputeStandings` to get sorted standings
- Creates 2 `TournamentMatch` rows with `stage="semi"`:
  - Match 1: `Team1TeamID = standings[0].TeamID` (seed 1), `Team2TeamID = standings[3].TeamID` (seed 4)
  - Match 2: `Team1TeamID = standings[1].TeamID` (seed 2), `Team2TeamID = standings[2].TeamID` (seed 3)
- Populates `Team1Player1ID/Team1Player2ID` by looking up team members for existing handicap logic

**maybeCreateFinalMatches:**
- Called from `RecordResult` when the recorded match has `stage="semi"`
- Queries all semi matches for this tournament; if both are `status="completed"`, creates:
  - `stage="final"`: winner of semi1 vs winner of semi2 (use `effective_winner` to determine)
  - `stage="third_place"`: loser of semi1 vs loser of semi2
- Guards against double-creation: check `count(*) WHERE stage IN ('final','third_place')` before inserting

### Patterns & Best Practices

- Follow existing `repository → service → handler` layering — no direct DB calls from handlers
- Use existing `MatchService.CreateMatch` for creating linked regular matches when `affects_score=true` (unchanged)
- All i18n strings in Vue components go through `vue-i18n` — no hardcoded Vietnamese/English text
- Pinia store: add `generateKnockouts(tournamentId)` action; refresh tournament detail on success

## Integration Points

**Create tournament (POST /api/v1/tournaments):**
1. Parse `format` from request
2. If `"round_robin_top4"`: validate 10 players, parse or auto-assign teams, persist `TournamentTeam` rows, call `GenerateTeamSchedule`, persist matches
3. Else: existing classic path unchanged

**Get tournament detail (GET /api/v1/tournaments/:id):**
1. Fetch tournament + participants + matches (existing)
2. If `format == "round_robin_top4"`: also fetch `tournament_teams`; call `ComputeStandings`; append to response
3. Else: `teams` and `standings` are omitted from response

**Record result (POST /api/v1/tournaments/:id/matches/:matchId/result):**
1. Existing logic runs unchanged
2. After saving: if `match.Stage == "semi"`, call `maybeCreateFinalMatches`

## Error Handling

| Scenario | HTTP Status | Message |
|----------|-------------|---------|
| `round_robin_top4` with ≠ 10 players | 400 | "round_robin_top4 format requires exactly 10 players (5 teams of 2)" |
| Duplicate player across teams | 400 | "each player must appear in exactly one team" |
| Generate knockouts — group matches incomplete | 400 | "all group stage matches must be completed before generating knockouts" |
| Generate knockouts — knockouts already exist | 409 | "knockout matches already generated for this tournament" |
| Team not found for knockout match | 500 | internal error — should never happen if data consistent |

## Performance Considerations

- `ComputeStandings` aggregates 10 rows — no caching needed
- `GenerateTeamSchedule` operates on 5 items in-memory — negligible
- `GET /tournaments/:id` fetches at most 14 matches (10 group + 4 knockout) — single query with preload is fine
- No new indexes needed beyond existing `tournament_id` foreign keys

## Security Notes

- No auth system; follows existing no-auth pattern
- `generate-knockouts` endpoint should validate tournament ownership/status server-side (not rely on client)
- Input validation: player IDs must be valid UUIDs referencing existing users; team composition checked for duplicates
