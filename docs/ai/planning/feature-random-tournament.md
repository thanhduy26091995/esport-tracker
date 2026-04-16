---
phase: planning
title: Random Tournament - Project Planning
description: Task breakdown for player tier system and random tournament feature
feature: random-tournament
created: 2026-04-15
---

# Project Planning & Task Breakdown

## Milestones

- [ ] **M1: Player Tier** — User model has tier + handicap_rate; API and UI updated
- [ ] **M2: Tournament Backend** — Tournament CRUD, schedule generation, result recording with handicap
- [ ] **M3: Tournament Frontend** — Create tournament, view schedule, record results, tournament leaderboard

## Task Breakdown

### Phase 1: Player Tier System

- [ ] **1.1 DB Migration** — Add `tier VARCHAR(10) DEFAULT 'normal'` and `handicap_rate FLOAT DEFAULT 0.0` to `users` table
- [ ] **1.2 User Model** — Add `Tier string` and `HandicapRate float64` fields to `model.User` in Go
- [ ] **1.3 User Repository** — Ensure `UpdateUser` passes through new fields; no other changes needed (GORM handles)
- [ ] **1.4 User Service** — Add `tier` and `handicap_rate` to `UpdateUserRequest`; validate tier is one of `pro|normal|noop`
- [ ] **1.5 User Handler** — No change needed (already passes request → service)
- [ ] **1.6 Frontend Types** — Add `tier` and `handicap_rate` to `User` TypeScript interface
- [ ] **1.7 Frontend User List** — Show `PlayerTierBadge` component on Users table/list
- [ ] **1.8 Frontend Edit User** — Add tier dropdown (Pro/Normal/Noop) and handicap_rate input to edit user form
- [ ] **1.9 Frontend Leaderboard** — Show tier badge next to player name

### Phase 2: Tournament Backend

- [ ] **2.0 MatchService — WinnerTeam=0 + TournamentMatchID** — Extend `CreateMatchRequest` with `TournamentMatchID *uuid.UUID` field and support `WinnerTeam=0` (create match record but apply zero point changes). Add `TournamentMatchID` to `Match` model.
- [ ] **2.1 Tournament Models** — Create `model/tournament.go` with `Tournament`, `TournamentParticipant`, `TournamentMatch`
- [ ] **2.2 DB Migration** — Create `tournament_matches`, `tournament_participants`, `tournaments` tables
- [ ] **2.3 Round-Robin Algorithm** — Implement `service/round_robin.go` — given ordered list of players (or teams for 2v2), generate `(N*(N-1))/2` matchups
- [ ] **2.4 Team Assigner** — Implement `service/team_assigner.go` — tier-balanced 2v2 team assignment (Pro paired with Noop > Normal > any)
- [ ] **2.5 Handicap Calculator** — Implement `service/handicap.go` — pure function: `EffectiveWinner(score1, score2 int, h1, h2 float64) *int`
- [ ] **2.6 Tournament Repository** — CRUD for Tournament, list TournamentMatches, update match result
- [ ] **2.7 Tournament Service — Create** — Accept player IDs, validate, snapshot tiers, run team assigner (2v2) or just shuffle (1v1), run round-robin, persist
- [ ] **2.8 Tournament Service — RecordResult** — Accept actual scores, compute handicap winner, optionally call `MatchService.CreateMatch` with effective winner, update tournament match
- [ ] **2.9 Tournament Service — Complete** — Mark tournament completed when all matches done (auto-trigger after last result, or manual endpoint)
- [ ] **2.10 Tournament Service — Delete** — Delete tournament: iterate all tournament_matches, delete linked regular matches (reverting scores via DeleteMatch), then delete tournament record (cascades participants + matches)
- [ ] **2.11 Tournament Handler** — Implement `api/tournament_handler.go` for all CRUD + result endpoints
- [ ] **2.12 Router** — Register tournament routes in `router.go`

### Phase 3: Tournament Frontend

- [ ] **3.1 TypeScript Types** — `types/tournament.ts`: `Tournament`, `TournamentMatch`, `TournamentParticipant`
- [ ] **3.2 Tournament Service** — `services/tournamentService.ts`: API wrappers for all tournament endpoints
- [ ] **3.3 Tournament Pinia Store** — `stores/tournamentStore.ts`: state, actions for fetch/create/result
- [ ] **3.4 PlayerTierBadge Component** — Small badge showing Pro/Normal/Noop with color coding
- [ ] **3.5 TournamentMatchCard Component** — Display a single match with team names, handicap info, and result input form
- [ ] **3.6 TournamentBracket Component** — Round-robin table: rows = rounds, cards per match
- [ ] **3.7 CreateTournamentView** — Wizard: select players → configure (type, affects_score, entry_fee, name) → submit
- [ ] **3.8 TournamentDetailView** — Full schedule, standings table, mark complete button
- [ ] **3.9 TournamentsView** — List of tournaments with status and quick stats
- [ ] **3.10 Router** — Add `/tournaments`, `/tournaments/:id`, `/tournaments/create` routes
- [ ] **3.11 Navigation** — Add Tournaments link to sidebar/nav

## Dependencies

```
1.1 → 1.2 → 1.3 → 1.4   (DB first, then model, repo, service)
1.6 → 1.7, 1.8, 1.9      (Types before components)
2.0 → 2.1                  (MatchService changes needed before tournament model uses them)
2.1 → 2.2                  (Model before migration)
2.3, 2.4, 2.5              (Independent algorithms)
2.6 → 2.7 → 2.8 → 2.9 → 2.10  (Repo before services, delete service after create)
2.3, 2.4, 2.5 → 2.7       (Algorithms used by service)
1.2, 2.0 → 2.7             (Need tier on User + WinnerTeam=0 for snapshot and result recording)
2.11 → 2.12                (Handler before router)
3.1 → 3.2 → 3.3            (Types → service → store)
3.3 → 3.5, 3.6, 3.7, 3.8  (Store before views)
3.4 → 3.5, 3.7             (Badge used in match card and create view)
3.7, 3.8, 3.9 → 3.10 → 3.11
```

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| 2v2 player count not even | Schedule impossible | Validate on API: return 400 if 2v2 and odd player count |
| Pro-only group (no Normal/Noop to pair with) | Team assigner fails | If not enough non-Pro for pairing, allow Pro-Pro pairing with a warning |
| Existing matches when user tier changes | Handicap inconsistency | Snapshots in `TournamentMatch` prevent retroactive changes |
| `affects_score=true` + settlement trigger mid-tournament | Interrupts flow | Existing settlement logic runs as usual; no special handling needed |

## Implementation Order (Recommended)
Phase 1 → Phase 2 → Phase 3 (sequential, each milestone is independently testable)
