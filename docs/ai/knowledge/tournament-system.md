# Tournament System

## Overview

Random 2v2 round-robin tournament generator with tier-balanced pairings, handicap snapshots, and optional score effects on the main leaderboard.

## Tables

| Table | Purpose |
|-------|---------|
| `tournaments` | Tournament header: name, status, affects_score flag |
| `tournament_participants` | Snapshot of player tier + handicap at creation time |
| `tournament_matches` | Individual round-robin matches; links to `matches.id` if affects_score |

## Tier-Balanced Pairing Rules

1. **Pros pair with Non-Pros** whenever possible (Pro + Normal/Noob team)
2. Remaining players randomly assigned
3. Handicap snapshot stored at tournament creation — not recomputed later

## Dynamic 2v2 Scheduler Algorithm

Generates match slots so every opponent pair plays at least once. Greedy approach:

- **Priority**: maximize uncovered opponent pairs
- **Penalties**: repeated teammates, repeated opponents, sit-out imbalance
- **Tie-break**: prefer Pro+NonPro pairing when penalty scores are equal

Runs in-memory; result persisted to `tournament_matches`.

## Match Scoring

- `EffectiveWinner` = actual score minus handicap
- `WinnerTeam = 0` means no-effect match (draw or admin decision)
- If `affects_score = true`: each tournament match creates a linked `matches` record that updates the main leaderboard

## API Endpoints

```
POST /api/v1/tournaments                          → create + generate schedule
GET  /api/v1/tournaments                          → list all
GET  /api/v1/tournaments/:id                      → full detail (matches, participants)
POST /api/v1/tournaments/:id/matches/:matchId/result → record match result
PUT  /api/v1/tournaments/:id/complete             → mark tournament complete
```

## Cascade Delete

Deleting a tournament unlinks (not deletes) any connected `matches` records — the regular match history is preserved.

## Frontend

- `CreateTournamentView.vue` — player picker with inline player creation (opens `UserForm` modal)
- `TournamentDetailView.vue` — shows schedule, allows recording results per round slot
- `TournamentsView.vue` — list page

## Inline Player Creation

Available from both `MatchForm` and `CreateTournamentView`. A "quick-add" button opens a `UserForm` modal. After save, the parent refreshes the player list. No auto-select after creation.
