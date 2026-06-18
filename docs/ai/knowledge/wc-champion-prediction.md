# WC Champion Prediction

## Overview

Tournament-level prediction separate from match betting. Each user picks one team to win the World Cup and wagers 1–5 points (virtual). Settled when the admin crowns the champion.

## Tables

| Table | Purpose |
|-------|---------|
| `wc_champion_teams` | List of available teams with their current odds |
| `wc_champion_config` | Singleton row (id=1, CHECK id=1): betting open/closed flag, settled_at |
| `wc_champion_predictions` | One per user (UNIQUE constraint on user_id): team_id, points_wagered, odds_snapshot, settled_at, payout |

## Key Rules

- **One prediction per user** — enforced by UNIQUE constraint on `wc_champion_predictions.user_id`
- **Odds snapshot at prediction time** — `odds_snapshot` stores the odds at the moment the user bet. Admin changing odds later does not affect existing predictions.
- **Settlement is idempotent** — checks `settled_at IS NOT NULL` before processing; safe to call again without double-paying
- **Config singleton** — `wc_champion_config` enforces `CHECK (id = 1)` at DB level; only one config row can ever exist

## Payout on Win

```
payout = points_wagered × odds_snapshot
```

Written to `wc_wallets` on settlement.

## API Endpoints

### User-Facing

```
GET    /wc/champion/teams            → available teams + current odds
GET    /wc/champion/config           → is betting open?
GET    /wc/champion/predictions      → all users' predictions (visible to everyone)
POST   /wc/champion/predict          → { team_id, points } — place or replace prediction
DELETE /wc/champion/predict          → cancel prediction (only if config.betting_open = true)
```

### Admin

```
PUT  /wc/admin/champion/teams/:id/odds   → update team odds
PUT  /wc/admin/champion/config           → open/close betting
POST /wc/admin/champion/settle           → { winner_team_id } — settle all predictions
```

## Frontend Views

- Champion prediction UI integrated into WC predict page or dedicated champion section
- Shows current odds per team, user's own pick with locked odds, countdown to close
