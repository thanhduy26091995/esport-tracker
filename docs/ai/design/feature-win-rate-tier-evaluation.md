---
phase: design
title: Win Rate & Tier Evaluation — System Design
description: Architecture and data flow for computing, persisting, and displaying win rate and tier
---

# System Design & Architecture

## Architecture Overview

Win rate is computed at query time and returned as part of the user response. Tier is computed after every match mutation and persisted to the `users.tier` column. This keeps win rate always fresh (no stale cache) while making tier queryable/filterable.

```mermaid
graph TD
    FE["Frontend (Vue 3)"] -->|GET /users or GET /users/leaderboard| API["UserHandler"]
    API --> US["UserService.GetAll / GetLeaderboard"]
    US --> UR["UserRepository.GetAll (win rate JOIN)"]
    UR -->|SQL: users LEFT JOIN match_participants WHERE point_change != 0| DB[(PostgreSQL)]
    DB --> UR
    UR -->|UserWithStats{win_rate, total_matches, tier}| US
    US --> API
    API --> FE

    FE2["Frontend"] -->|POST /matches| MH["MatchHandler"]
    MH --> MS["MatchService.Create"]
    MS --> MR["MatchRepository"]
    MR -->|INSERT match + participants, tx.Commit| DB
    MS -->|"post-commit: RecalculateForUsers(participantIDs)"| TS["TierService"]
    TS --> UR2["UserRepository.GetWinRatesBatch(ids)"]
    UR2 -->|single SQL WHERE user_id IN (...)| DB
    DB --> UR2
    UR2 --> TS
    TS --> UR3["UserRepository.UpdateTier(id, tier) × N"]
    UR3 -->|UPDATE users SET tier=...| DB

    STARTUP["App Startup (SetupRouter)"] -->|"once: RecalculateAllTiers()"| TS
```

## Data Models

### Backend — no schema migration needed

`users.tier` already exists. Win rate and total matches are **computed, not stored**.

New response struct (Go):

```go
// UserWithStats extends User with computed win rate fields
type UserWithStats struct {
    User
    WinRate      float64 `json:"win_rate"`      // 0.0–1.0, e.g. 0.65 = 65%
    TotalMatches int     `json:"total_matches"` // total matches played
    WonMatches   int     `json:"won_matches"`   // matches won
}
```

Tier values stored in `users.tier`:

| Value | Condition |
|-------|-----------|
| `"pro"` | win_rate ≥ 0.60 AND total_matches ≥ 10 |
| `"normal"` | 0.40 ≤ win_rate < 0.60 AND total_matches ≥ 10 |
| `"noob"` | win_rate < 0.40 AND total_matches ≥ 10 |
| `"normal"` | total_matches < 10 (default, not yet evaluated) |

### Frontend — extended type

```typescript
export interface UserWithStats extends User {
  win_rate: number        // 0.0–1.0
  total_matches: number
  won_matches: number
}
```

## API Design

### Updated endpoints (return `UserWithStats` instead of `User`)

**GET /api/v1/users**
```json
[
  {
    "id": "...",
    "name": "Alice",
    "current_score": 12,
    "tier": "pro",
    "win_rate": 0.65,
    "total_matches": 20,
    "won_matches": 13,
    ...
  }
]
```

**GET /api/v1/users/leaderboard** — same shape, sorted by `current_score DESC`

**GET /api/v1/users/:id** — same shape, single user

No new endpoints required.

### SQL for win rate (computed in repository layer)

Draw matches (`point_change = 0`) are **excluded** from both the numerator and denominator — win rate is `wins / (wins + losses)` only.

```sql
SELECT
  u.*,
  COUNT(mp.id) FILTER (WHERE mp.point_change != 0)              AS total_matches,
  COUNT(mp.id) FILTER (WHERE mp.point_change > 0)               AS won_matches,
  CASE
    WHEN COUNT(mp.id) FILTER (WHERE mp.point_change != 0) = 0 THEN 0
    ELSE COUNT(mp.id) FILTER (WHERE mp.point_change > 0)::float
         / COUNT(mp.id) FILTER (WHERE mp.point_change != 0)
  END AS win_rate
FROM users u
LEFT JOIN match_participants mp ON mp.user_id = u.id
GROUP BY u.id
```

The same query filtered to specific user IDs is used by `GetWinRatesBatch(ids []uuid.UUID)` (adds `WHERE u.id IN (...)`), called during tier recalculation.

## Component Breakdown

### Backend changes
- `model/user.go` — add `UserWithStats` struct
- `repository/user_repository.go` — update `GetAll`, `GetByID`, `GetLeaderboard` to return `UserWithStats` with win rate subquery (draws excluded)
- `repository/user_repository.go` — add `GetWinRatesBatch(ids []uuid.UUID) (map[uuid.UUID]WinRateStats, error)` for bulk tier recalculation
- `repository/user_repository.go` — add `UpdateTier(id uuid.UUID, tier string) error`
- `service/user_service.go` — update service methods to pass through `UserWithStats`
- `service/tier_service.go` *(new)* — three methods:
  - `RecalculateForUsers(ids []uuid.UUID) error`: batch win rate fetch → evaluate → `UpdateTier` per user
  - `RecalculateAllTiers() error`: calls `GetAll` users then `RecalculateForUsers` — used for startup backfill
  - `EvaluateTier(winRate float64, totalMatches int) string`: pure function
- `service/match_service.go` — after `tx.Commit()` in `CreateMatch`/`DeleteMatch`, call `TierService.RecalculateForUsers` with participant IDs (non-fatal)
- `api/router.go` — after service initialization, call `tierService.RecalculateAllTiers()` once at startup

### Frontend changes
- `types/user.ts` — extend `User` → add `win_rate`, `total_matches`, `won_matches` to `UserWithStats`; update stores/services to use new type
- `components/UserTable.vue` — add "Win Rate" column showing `XX%` and tier badge
- `components/WinRateBadge.vue` *(new)* — displays tier chip (pro/normal/noob) with color coding; shows `—` for both win rate and tier when `total_matches < 10`
- `stores/userStore.ts` — no structural change; type update only
- `services/userService.ts` — no structural change; type update only

### UI — tier badge color scheme

| Tier | Color | Condition |
|------|-------|-----------|
| pro | gold / `#f59e0b` | win_rate ≥ 60% AND total_matches ≥ 10 |
| normal | blue / `#3b82f6` | 40% ≤ win_rate < 60% AND total_matches ≥ 10 |
| noob | gray / `#6b7280` | win_rate < 40% AND total_matches ≥ 10 |
| — (no badge) | — | total_matches < 10 → show dash, no chip |

Win rate display: `65%` (formatted as integer %, e.g. `Math.round(win_rate * 100)`). Players with `total_matches < 10` show `—` for both the win rate percentage and the tier badge (no chip rendered).

## Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Store win rate in DB? | No — compute at query time | Avoids sync issues; match history is the source of truth |
| Store tier in DB? | Yes — `users.tier` | Allows sorting/filtering by tier; updated on every match mutation |
| Tier recalculation timing? | Post-commit, outside match transaction | Match mutation never rolls back due to a tier failure; simpler wiring |
| Bulk win rate fetch? | `GetWinRatesBatch(ids)` — single SQL | One query for all match participants; avoids N+1 on 2v2 recalculation |
| Draw match handling? | Excluded from win rate | `win_rate = wins / (wins + losses)`, draws ignored — players not penalized for draws |
| Minimum sample size | 10 matches | Prevents misleading tiers from small samples; show `—` below threshold |
| Win rate basis | All match types (incl. tournament) | Simplest definition; consistent with overall score |
| Backfill on startup? | `RecalculateAllTiers()` in `SetupRouter` | Ensures tiers are correct immediately on first deploy without a separate migration step |

## Non-Functional Requirements

- **Performance:** Win rate query uses a single LEFT JOIN + GROUP BY — acceptable at this scale (< 1000 users). Add index on `match_participants(user_id)` if not already present. Batch recalculation (`GetWinRatesBatch`) uses a single SQL with `WHERE user_id IN (...)`.
- **Correctness:** Tier recalculation runs post-commit — a tier update failure never rolls back a valid match. Draws are excluded from win rate so they don't artificially lower a player's tier.
- **Scalability:** If user count grows significantly, win rate could be stored (cached) and invalidated on match events — deferred for now.
- **i18n:** Tier labels (`tier.pro`, `tier.normal`, `tier.noob`) must be added to both `en` and `vi` locale files. No `tier.unranked` key needed — the `—` dash is rendered directly in the component.
