# WC Betting System

## Overview

Match-level betting for WC2026 with three bet types: handicap, exact score, and over/under. Virtual wallet balances. Admin-controlled settlement. House P&L dashboard.

## Tables

| Table | Purpose |
|-------|---------|
| `wc_bets` | Individual bet records: user, match, type, stake, odds_snapshot, selection, payout NUMERIC(10,2), status |
| `wc_settlements` | One settlement record per admin tất toán (reset) cycle |
| `wc_settlement_details` | Per-user breakdown for each settlement |
| `wc_score_odds` | Admin-defined exact score options with odds per match |

## Bet Types

### Handicap

- Admin sets `wc_matches.home_handicap` and `wc_matches.away_handicap` (imported from StatsAPI or manually set)
- Player picks Home or Away
- Effective score = actual score ± handicap (direction depends on which side gives handicap)
- Result: win, lose, push (exact tie after handicap), win_half, lose_half (common in Asian handicap)

### Exact Score

- Admin defines score options via `POST /admin/matches/:id/score-odds`
- Each option has `home_goals`, `away_goals`, `odds`
- `odds_snapshot` stored at bet time — admin can later change odds without affecting existing bets
- Result: win if match final score matches selection exactly, else lose

### Over/Under

- Admin imports via `POST /admin/matches/:id/import-ou` from StatsAPI
- `wc_matches.ou_line` (e.g., 2.5), `odds_over`, `odds_under`
- Player picks Over or Under total goals

## Payout Calculation

```
payout = round(stake × odds, 2)  // banker's rounding, 2 decimal places
```

Payout type: `NUMERIC(10,2)` in DB and Go (`*float64`). Never use integer for payout.

## Bet Statuses

`pending` → `settled` | `void`

Void bets: match cancelled, or admin manually voids. Void bets excluded from house P&L profit.

## Settlement (Tất Toán)

Admin-triggered cycle to reset all wallets:
1. Calculate net P&L per user across all settled bets since last settlement
2. Create `wc_settlements` record
3. Create `wc_settlement_details` per user (collect or pay)
4. Reset `wc_wallets.balance` to 0
5. Write `wc_wallet_logs` for each reset

Leaderboard (`GET /wc/leaderboard`) shows net P&L from **all settled bets** — survives multiple tất toán cycles.

## House P&L Dashboard

Admin view at `GET /admin/house-pnl`:
- **Total aggregate**: sum(stake) - sum(payout) for all settled bets; pending and void shown separately
- **Per-match breakdown**: match-level stake/payout totals

Dependency: requires `wc_bets.payout` to be `NUMERIC(10,2)` (Betting Refinements feature).

## StatsAPI Odds Import

Admin endpoints to pull handicap and O/U odds:
```
POST /admin/matches/:id/import-handicap   → sets home_handicap, away_handicap, odds_synced_at
POST /admin/matches/:id/import-ou         → sets ou_line, odds_over, odds_under, ou_synced_at
POST /admin/matches/:id/generate-poisson  → compute exact-score probabilities via Poisson model
GET  /admin/sync-logs                     → audit trail in wc_sync_logs
```

**Cron behaviour**: fills blank odds only (does not overwrite admin edits).  
**Manual import**: always overwrites (preview-then-confirm flow in admin UI).

## Poisson Odds Engine

Input: admin-defined `house_margin` (e.g., 0.05) and `min_prob` threshold.  
Computes scoreline probabilities from historical goal rates; converts to odds with margin applied.  
Results written to `wc_score_odds` for the match.

## User-Facing Betting Endpoints

```
GET  /wc/matches/:id/bets       → user's own bets for a match
POST /wc/matches/:id/bet        → place bet { type, selection, stake }
DELETE /wc/bets/:id             → cancel pending bet (refunds stake)
GET  /wc/wallet                 → current balance + recent logs
```

## Admin Betting Endpoints

```
POST /admin/matches/:id/settle           → settle all bets for a match
PUT  /admin/bets/:id/void                → void a specific bet
POST /admin/matches/:id/score-odds       → define exact score options
GET  /admin/house-pnl                    → house profit/loss dashboard
```
