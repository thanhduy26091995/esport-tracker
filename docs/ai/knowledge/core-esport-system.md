# Core Esport System

## Overview

The original FC25 match tracker — separate from the WC2026 feature. Tracks match results, scores, debt settlement, and a shared fund.

## Tables

| Table | Purpose |
|-------|---------|
| `users` | Player info: name, score, tier, avatar_url, favorite_club |
| `matches` | Match records: type (1v1/2v2/1v2), teams, timestamp |
| `match_participants` | Per-player point changes for a match |
| `debt_settlements` | Debt payment records when a player hits threshold |
| `fund_transactions` | Shared fund income (from debt) and expenses |
| `score_bonuses` | External bonus points (e.g., cookie losses) |
| `config` | System config: debt threshold, point conversion rate, fund split |

## Match Types

| Type | Team 1 | Team 2 | Scoring Rule |
|------|--------|--------|-------------|
| `1v1` | 1 player | 1 player | winner +1, loser -1 |
| `2v2` | 2 players | 2 players | each winner +1, each loser -1 |
| `1v2` | 1 player | 2 players | solo win: solo +2, each duo -1; duo win: each duo +1, solo -2 (net zero) |

Team size validation map: `{ "1v1": {1,1}, "2v2": {2,2}, "1v2": {1,2} }`.

## Debt Settlement

Trigger: player score ≤ `config.debt_threshold` (default -6).

1. Calculate debt in VND: `abs(score) × config.point_conversion_rate`
2. Split by `config.fund_split_ratio` (default 50/50):
   - 50% → `fund_transactions` (fund income)
   - 50% → distributed to recent winning opponents
3. Debtor's score resets to 0
4. Deduct points from creditor winners (money received ÷ conversion rate)
5. Lock all related matches (`locked = true`) — cannot be edited after settlement

Always use `shopspring/decimal` for all VND calculations.

## Score Bonuses (`score_bonuses`)

Award or deduct points from external events (e.g., cookie losses):

- `POST /score-bonuses` → insert record + update `users.score` + recalculate tier
- `DELETE /score-bonuses/:id` → reverts score change
- Merged into unified activity feed returned by `GET /matches` with discriminator `type: "match" | "bonus"`

## User Tiers

Computed via `TierService` after every match result commit and after score bonus changes.  
Backfill available via `RecalculateAllTiers()` at startup.

| Tier | Condition |
|------|-----------|
| `pro` | win_rate ≥ 60% AND total_matches ≥ 10 |
| `normal` | 40% ≤ win_rate < 60% AND total_matches ≥ 10 |
| `noob` | win_rate < 40% AND total_matches ≥ 10 |
| `normal` | default when total_matches < 10 |

Win rate = wins / (wins + losses). Draws excluded.  
`users.tier` is persisted to DB; win_rate and match counts are computed at query time.

## API Endpoints (Core)

```
# Users
GET    /api/v1/users                  → UserWithStats[] (includes win_rate, tier)
POST   /api/v1/users
GET    /api/v1/users/:id
PUT    /api/v1/users/:id
DELETE /api/v1/users/:id
GET    /api/v1/users/leaderboard      → ranked UserWithStats[]
GET    /api/v1/users/payment-ranking  → users sorted by total debt paid

# Matches (returns merged feed: matches + bonuses)
POST   /api/v1/matches
GET    /api/v1/matches                → ?type=match|bonus, ?player_id=uuid
GET    /api/v1/matches/:id
PUT    /api/v1/matches/:id            → only if not locked
DELETE /api/v1/matches/:id

# Score bonuses
POST   /api/v1/score-bonuses
DELETE /api/v1/score-bonuses/:id

# Settlements
GET    /api/v1/settlements
GET    /api/v1/settlements/:id

# Fund
GET    /api/v1/fund                   → balance + recent transactions
POST   /api/v1/fund/expense

# Config
GET    /api/v1/config
PUT    /api/v1/config/:key
```

## Player Personalization

- **Avatar**: `PUT /users/:id/avatar` (multipart upload), stored at `./uploads/avatars/{uuid}`. Gin static route serves files. Validates MIME (rejects SVG to prevent XSS).
- **Favorite club**: `PUT /users/:id/club` (club slug string). 20 pre-defined clubs in `src/config/clubs.ts`.
- **Dynamic global theme**: `useGlobalTheme` composable watches `leaderboard[0]`, applies club CSS vars to `<html>` element.

## Dashboard Sort Strategies

Three strategies selectable per view (persisted in localStorage):

1. **Debt First** — `GET /users/payment-ranking` (backend SUM of `debt_settlements.money_amount`)
2. **Default** — client-side by current score grouping
3. **Winners First** — client-side sort

Frontend utility `sortByStrategy()` handles strategies 2 and 3.
