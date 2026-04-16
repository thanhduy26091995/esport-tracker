---
phase: implementation
title: Random Tournament - Implementation Guide
description: Technical notes for implementing player tier and random tournament
feature: random-tournament
created: 2026-04-15
---

# Implementation Guide

## Development Setup

- Go 1.21+ backend with GORM + PostgreSQL
- Vue 3 + TypeScript + Pinia frontend
- Run `go run ./cmd/...` from `backend/` to start backend
- Run `npm run dev` from `frontend/` to start frontend

## Code Structure

### Modified: `matches` table  
Add `tournament_match_id` field to link a regular match back to its tournament match origin.

### New files to create
```
backend/
  internal/model/tournament.go
  internal/repository/tournament_repository.go
  internal/service/round_robin.go
  internal/service/team_assigner.go
  internal/service/handicap.go
  internal/service/tournament_service.go
  internal/api/tournament_handler.go
  migrations/XXXX_add_tier_and_tournaments.sql

frontend/src/
  types/tournament.ts
  services/tournamentService.ts
  stores/tournamentStore.ts
  components/PlayerTierBadge.vue
  components/TournamentMatchCard.vue
  components/TournamentBracket.vue
  views/TournamentsView.vue
  views/TournamentDetailView.vue
  views/CreateTournamentView.vue
```

## Implementation Notes

### Round-Robin Algorithm (Polygon Rotation)
```go
// GenerateRoundRobin returns list of rounds, each round is list of (i, j) matchup indices
// Uses standard "circle method": fix player[0], rotate rest
func GenerateRoundRobin(n int) [][]MatchPair {
    players := makeRange(0, n)
    if n%2 == 1 {
        players = append(players, -1) // -1 = bye
        n++
    }
    rounds := make([][]MatchPair, n-1)
    for round := 0; round < n-1; round++ {
        pairs := []MatchPair{}
        for i := 0; i < n/2; i++ {
            p1 := players[i]
            p2 := players[n-1-i]
            if p1 != -1 && p2 != -1 {
                pairs = append(pairs, MatchPair{p1, p2})
            }
        }
        rounds[round] = pairs
        // rotate: fix players[0], rotate players[1:]
        last := players[n-1]
        copy(players[2:], players[1:n-1])
        players[1] = last
    }
    return rounds
}
```

### Team Assigner (2v2 Tier Balancing)
```go
// AssignTeams takes flat player list and returns list of Team pairs []{Team1, Team2}
// Each Team = [player1, player2]
// Rules:
//   - Pro player must not be paired with another Pro
//   - Pro prefers Noop partner; else Normal; else any available
func AssignTeams(participants []TournamentParticipant) ([][2][2]uuid.UUID, error) {
    pros := filter(participants, tier == "pro")
    noops := shuffle(filter(participants, tier == "noop"))
    normals := shuffle(filter(participants, tier == "normal"))
    pool := append(noops, normals...)

    teams := [][2]uuid.UUID{}
    for _, pro := range pros {
        if len(pool) == 0 {
            // fallback: pair pros together (warn)
            return nil, errors.New("not enough non-pro players for balanced assignment")
        }
        partner := pool[0]
        pool = pool[1:]
        teams = append(teams, [2]uuid.UUID{pro.UserID, partner.UserID})
    }
    // pair remaining randomly
    shuffle(pool)
    for i := 0; i+1 < len(pool); i += 2 {
        teams = append(teams, [2]uuid.UUID{pool[i].UserID, pool[i+1].UserID})
    }
    return pairTeams(teams), nil // pair teams into matchups
}
```

### Handicap Calculator
```go
// Returns effective winner: 1, 2, or 0 (draw)
func EffectiveWinner(score1, score2 int, handicap1, handicap2 float64) int {
    eff1 := float64(score1) - handicap1
    eff2 := float64(score2) - handicap2
    switch {
    case eff1 > eff2: return 1
    case eff2 > eff1: return 2
    default: return 0
    }
}

### Recording Result with affects_score
```go
func (s *TournamentService) RecordMatchResult(tournamentID, matchID uuid.UUID, score1, score2 int, recordedBy string) (*TournamentMatch, error) {
    tm := s.repo.GetMatch(matchID)
    tournament := s.repo.GetTournament(tournamentID)
    effectiveWinner := EffectiveWinner(score1, score2, tm.HandicapTeam1, tm.HandicapTeam2)

    // Override: revert previous linked regular match
    if tm.Status == "completed" && tm.MatchID != nil {
        s.matchService.DeleteMatch(*tm.MatchID) // reverses score changes
        tm.MatchID = nil
    }

    // Determine match winner for regular match record
    // affects_score=false OR effective draw → WinnerTeam=0 (no score change)
    matchWinnerTeam := 0
    if tournament.AffectsScore && effectiveWinner != 0 {
        matchWinnerTeam = effectiveWinner
    }

    // Always create regular match (for main history visibility)
    req := &CreateMatchRequest{
        MatchType:        tournament.MatchType,
        Team1:            buildTeam(tm, 1),
        Team2:            buildTeam(tm, 2),
        WinnerTeam:       matchWinnerTeam, // 0 = no score change
        RecordedBy:       recordedBy,
        TournamentMatchID: &tm.ID,
    }
    match, err := s.matchService.CreateMatch(req)
    if err == nil {
        tm.MatchID = &match.ID
    }

    tm.ActualScore1 = &score1
    tm.ActualScore2 = &score2
    tm.EffectiveWinner = effectiveWinner
    tm.Status = "completed"
    s.repo.SaveMatch(tm)
    return tm, nil
}
```

> **Required MatchService change:** `CreateMatchRequest` needs `TournamentMatchID *uuid.UUID` field. `WinnerTeam=0` must be supported (record match with no point changes).

## Integration Points

- **MatchService.CreateMatch**: reused as-is for regular score tracking when `affects_score=true`
- **Settlement**: runs automatically via existing hook in MatchService when score threshold reached — no special handling needed
- **UserRepository**: `GetByID` already returns User; just needs `Tier` and `HandicapRate` fields in model

## Error Handling

| Case | HTTP Status | Message |
|------|-------------|---------|
| 2v2 with odd player count | 400 | "2v2 requires even number of players" |
| Less than 3 players | 400 | "Tournament requires at least 3 players" |
| Not enough non-Pro for balancing | 422 | "Cannot balance teams: too many Pro players" (proceed anyway with warning flag in response) |
| Record result on completed match | 409 | "Match result already recorded" |
| Invalid tier value | 400 | "tier must be one of: pro, normal, noop" |

## Frontend Implementation Notes

### PlayerTierBadge.vue
```vue
<!-- Color coding: pro=gold, normal=blue, noop=gray -->
<template>
  <span :class="tierClass">{{ tierLabel }}</span>
</template>
```

### CreateTournamentView.vue Flow
1. Multi-select players from existing user list (checkboxes)
2. Configure: name, match_type (1v1/2v2), affects_score toggle, entry_fee
3. Submit → POST /api/v1/tournaments → redirect to TournamentDetailView

### TournamentDetailView.vue
- Display matches grouped by round
- Each match shows: team names, tier badges, handicap info, score input (if pending)
- Standings table: player | W | D | L | Points (calculated client-side from completed matches)
- "Mark Complete" button when all matches done

## Migration SQL
```sql
-- Migration: add tier and handicap to users
ALTER TABLE users 
  ADD COLUMN tier VARCHAR(10) NOT NULL DEFAULT 'normal',
  ADD COLUMN handicap_rate FLOAT NOT NULL DEFAULT 0.0;

-- Tournaments
CREATE TABLE tournaments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(200) NOT NULL,
  match_type VARCHAR(10) NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'active',
  affects_score BOOLEAN NOT NULL DEFAULT true,
  entry_fee INT NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Tournament participants
CREATE TABLE tournament_participants (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id),
  tier_snapshot VARCHAR(10) NOT NULL DEFAULT 'normal',
  handicap_rate_snapshot FLOAT NOT NULL DEFAULT 0.0
);

-- Tournament matches
CREATE TABLE tournament_matches (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
  round INT NOT NULL,
  match_order INT NOT NULL,
  team1_player1_id UUID NOT NULL REFERENCES users(id),
  team1_player2_id UUID REFERENCES users(id),
  team2_player1_id UUID NOT NULL REFERENCES users(id),
  team2_player2_id UUID REFERENCES users(id),
  handicap_team1 FLOAT NOT NULL DEFAULT 0.0,
  handicap_team2 FLOAT NOT NULL DEFAULT 0.0,
  status VARCHAR(20) NOT NULL DEFAULT 'pending',
  actual_score1 INT,
  actual_score2 INT,
  effective_winner INT,
  match_id UUID REFERENCES matches(id),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```
