# WC2026 Core System

## Overview

Standalone prediction/betting platform for World Cup 2026, running alongside (but isolated from) the core esport tracker. Has its own users, config, and feature flag.

## Isolation Principle

WC tables are prefixed `wc_`. WC backend files are prefixed `wc_`. WC routes are all under `/api/v1/wc/`. WC never writes to core tables (`users`, `matches`, etc.) and vice versa. The only shared infrastructure is the Go server, database connection, and JWT secret.

## Tables (WC Core)

| Table | Purpose |
|-------|---------|
| `wc_users` | WC players: name, password_hash (nullable), google_id (nullable), avatar_url, is_admin |
| `wc_config` | Single row (id=1): feature flag `is_enabled`, updatedBy, updatedAt |
| `wc_matches` | Match schedule synced from external API: teams, date, status, stage, group, odds fields |
| `wc_wallets` | One wallet per player: balance NUMERIC(10,2) |
| `wc_wallet_logs` | Audit trail for all wallet changes |

## Feature Flag

`wc_config.is_enabled` (boolean, id=1) controls the entire WC feature:
- `WcFeatureMiddleware` blocks non-essential routes → 404 when disabled
- Frontend route guard fetches `/api/v1/wc/config` and redirects to schedule when off
- Public routes (`/wc/config`, `/wc/matches`, `/wc/schedule`) always accessible regardless of flag

## WC Match Statuses

`scheduled` → `live` → `completed` | `cancelled`

## WC Match Stages

`group`, `r32`, `r16`, `qf`, `sf`, `final`

## Wallet System

- Wallet auto-created when a new WC user is registered (Google signup or admin creation)
- Balance type: `NUMERIC(10,2)` — always use `shopspring/decimal` in Go
- Admin can top-up wallets via `POST /admin/wallets/:userId/topup`
- Every wallet change writes an audit row to `wc_wallet_logs`

## Match Sync (StatsAPI)

WC matches are synced from `thestatsapi.com`:
- Background cron every 30 minutes in `router.go`
- Upsert on `external_id` (ON CONFLICT clause) — never duplicates
- One-time admin mapping: `POST /admin/setup-statsapi-mapping` maps statsapi fixture IDs to `wc_matches`

## Key API Endpoints (Public / Core)

```
GET  /api/v1/wc/config            → feature flag status
GET  /api/v1/wc/matches           → match list (filter: status, stage, group, date_from, date_to, player_id)
GET  /api/v1/wc/schedule          → public schedule (no auth)
GET  /api/v1/wc/leaderboard       → net P&L from settled bets
```

## Admin Endpoints (Core)

```
PUT  /api/v1/wc/admin/config          → toggle is_enabled
GET  /api/v1/wc/admin/users           → list WC users
POST /api/v1/wc/admin/users           → create WC user
PUT  /api/v1/wc/admin/wallets/:userId/topup
POST /api/v1/wc/admin/setup-statsapi-mapping
```

## Auth System

See `docs/ai/knowledge/wc-auth-system.md` for full details on JWT, Google OAuth, middleware stack, and route guards.

## Upcoming Matches Dashboard Widget

Horizontal scrollable widget on the main dashboard (not WC-specific page):
- Fetches `/wc/matches?date_from=now-4h&date_to=now+48h`
- Filters client-side to `scheduled` + `live` statuses
- Auto-refreshes every 5 minutes using a separate `wcPublicApi` instance (no auth interceptor — fails silently)
- Shows "LIVE" badge or VN timezone time (`Asia/Ho_Chi_Minh`)
