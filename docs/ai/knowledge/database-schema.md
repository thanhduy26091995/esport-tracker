# Database Schema

PostgreSQL, managed via GORM AutoMigrate. All PKs are UUID (`gen_random_uuid()`), timestamps are `time.Time`.

---

## Core Esport System

### `users`
FC25 players (separate from WC users).

| Column | Type | Notes |
|--------|------|-------|
| id | uuid PK | |
| name | varchar(100) | unique where `is_active = true` |
| current_score | int | default 0 |
| is_active | bool | default true; soft-delete pattern |
| tier | varchar(10) | `normal` \| `pro` \| `noob`; default `normal` |
| handicap_rate | float64 | default 0.0 |
| avatar_url | varchar(255) | nullable |
| favorite_club | varchar(50) | nullable |
| created_at, updated_at | timestamp | |

### `matches`
FC25 match results.

| Column | Type | Notes |
|--------|------|-------|
| id | uuid PK | |
| match_type | varchar(10) | `1v1` \| `2v2` \| `1v2` |
| winner_team | int | 1 or 2 |
| match_date | timestamp | default now() |
| recorded_by | varchar(100) | |
| is_locked | bool | default false; locked matches can't be edited |
| tournament_match_id | uuid | FK → `tournament_matches`, nullable |
| created_at | timestamp | |

### `match_participants`
Links users to matches.

| Column | Type | Notes |
|--------|------|-------|
| id | uuid PK | |
| match_id | uuid | FK → `matches` CASCADE |
| user_id | uuid | FK → `users` |
| team_number | int | 1 or 2 |
| point_change | int | +1 (winner) or -1 (loser) |

### `score_bonuses`
Manual bonus/penalty points awarded outside matches.

| Column | Type | Notes |
|--------|------|-------|
| id | uuid PK | |
| user_id | uuid | FK → `users` |
| points | int | positive = bonus, negative = penalty |
| description | text | |
| recorded_by | varchar(100) | |
| bonus_date | timestamp | default now() |
| created_at | timestamp | |

### `debt_settlements`
Triggered when a player's score hits `debt_threshold` (default −6).

| Column | Type | Notes |
|--------|------|-------|
| id | uuid PK | |
| debtor_id | uuid | FK → `users` |
| debt_amount | int | negative (e.g. −7) |
| money_amount | int | total VND paid |
| fund_amount | int | VND portion → fund |
| winner_distribution | int | VND portion → winners |
| original_debt_points | int | absolute value of debt |
| to_fund | numeric(12,2) | legacy column — use `fund_amount` |
| to_winners | numeric(12,2) | legacy column — use `winner_distribution` |
| settlement_date | timestamp | |
| created_at | timestamp | |

### `settlement_winners`
Winners who receive a share of a debt settlement.

| Column | Type | Notes |
|--------|------|-------|
| id | uuid PK | |
| settlement_id | uuid | FK → `debt_settlements` |
| winner_id | uuid | FK → `users` |
| money_amount | int | VND received |
| points_deducted | int | points removed from winner's score |

### `fund_transactions`
Tracks the group fund balance (deposits from settlements, withdrawals).

| Column | Type | Notes |
|--------|------|-------|
| id | uuid PK | |
| amount | int | VND |
| transaction_type | varchar(20) | `deposit` \| `withdrawal` |
| description | text | |
| related_settlement_id | uuid | FK → `debt_settlements`, nullable |
| transaction_date | timestamp | |
| created_at | timestamp | |

### `config`
Key-value app configuration.

| Key | Default | Meaning |
|-----|---------|---------|
| `debt_threshold` | −6 | Score that triggers settlement |
| `point_to_vnd` | 22000 | 1 point = X VND |
| `fund_split_percent` | 50 | % of debt money → fund (rest → winners) |
| `auto_settlement` | false | Auto-settle on threshold hit |
| `min_matches_for_tier` | 5 | Min matches before tier applies |
| `pro_win_rate_threshold` | 0.60 | Win rate ≥ this → Pro tier |
| `normal_win_rate_threshold` | 0.40 | Win rate < this → Noob tier |

---

## Tournament System

### `tournaments`

| Column | Type | Notes |
|--------|------|-------|
| id | uuid PK | |
| name | varchar(200) | |
| match_type | varchar(10) | `1v1` \| `2v2` |
| format | varchar(30) | `classic` \| `round_robin_top4` |
| knockout_size | int | default 4; 2 = final only; 4 = semis + final + 3rd place |
| status | varchar(20) | `active` \| `completed` |
| affects_score | bool | default true; whether matches count toward `current_score` |
| entry_fee | int | default 0 |
| champion_team_id | uuid | FK → `tournament_teams`, nullable |
| created_at, updated_at | timestamp | |

### `tournament_participants`
Snapshot of each player's tier/handicap at enrollment time.

| Column | Type | Notes |
|--------|------|-------|
| id | uuid PK | |
| tournament_id | uuid | FK → `tournaments` CASCADE |
| user_id | uuid | FK → `users` |
| tier_snapshot | varchar(10) | snapshot at enrollment |
| handicap_rate_snapshot | float64 | snapshot at enrollment |

### `tournament_teams`
2v2 teams. Each row = one team of two players.

| Column | Type | Notes |
|--------|------|-------|
| id | uuid PK | |
| tournament_id | uuid | FK → `tournaments` CASCADE |
| player1_id | uuid | FK → `users` |
| player2_id | uuid | FK → `users` |
| player1_handicap_snapshot | float64 | |
| player2_handicap_snapshot | float64 | |
| created_at | timestamp | |

### `tournament_matches`
Scheduled fixtures within a tournament.

| Column | Type | Notes |
|--------|------|-------|
| id | uuid PK | |
| tournament_id | uuid | FK → `tournaments` CASCADE |
| round | int | round number |
| match_order | int | position within round |
| stage | varchar(20) | `group` \| `semi` \| `final` \| `third_place` |
| team1_team_id, team2_team_id | uuid | FK → `tournament_teams`, nullable (2v2 only) |
| team1_player1_id, team2_player1_id | uuid | FK → `users`, required |
| team1_player2_id, team2_player2_id | uuid | FK → `users`, nullable (null for 1v1) |
| handicap_team1, handicap_team2 | float64 | applied handicap per side |
| status | varchar(20) | `pending` \| `completed` |
| actual_score1, actual_score2 | int | nullable until played |
| effective_winner | int | 0 = pending/draw, 1 = team1, 2 = team2 |
| match_id | uuid | FK → `matches`, nullable — links to actual match record |
| created_at, updated_at | timestamp | |

---

## WC2026 System

### `wc_users`
WC2026 app accounts — completely separate from `users`.

| Column | Type | Notes |
|--------|------|-------|
| id | uuid PK | |
| name | varchar(100) | globally unique |
| password_hash | varchar(255) | nullable (null = Google-only account) |
| google_id | varchar(100) | unique where not null; nullable |
| email | varchar(255) | nullable |
| avatar_url | varchar(500) | nullable |
| is_admin | bool | default false |
| is_blocked | bool | default false |
| created_at, updated_at | timestamp | |

### `wc_config`
Singleton row (id = 1). Global feature flag.

| Column | Type | Notes |
|--------|------|-------|
| id | int PK | always 1 |
| is_enabled | bool | default false; gates the entire WC feature |
| updated_at | timestamp | |
| updated_by | uuid | FK → `wc_users`, nullable |

### `wc_matches`
Real WC2026 match data, synced from StatsAPI.

| Column | Type | Notes |
|--------|------|-------|
| id | uuid PK | |
| external_id | varchar(64) | unique; StatsAPI match ID |
| home_team, away_team | varchar(100) | full team names |
| home_team_code, away_team_code | char(3) | e.g. `ARG` |
| match_date | timestamp | |
| group_name | varchar(30) | e.g. `Group A` |
| stage | varchar(30) | `group` \| `r32` \| `r16` \| `qf` \| `sf` \| `final` \| `third_place` |
| venue | varchar(100) | |
| home_score, away_score | int | nullable until played |
| status | varchar(20) | `scheduled` \| `live` \| `completed` \| `cancelled` |
| handicap_team | varchar(5) | `home` \| `away` (which side gives the handicap) |
| handicap_value | numeric(5,2) | nullable |
| odds_handicap_home, odds_handicap_away | numeric(5,2) | nullable |
| predictions_open | bool | default false |
| predictions_locked_at | timestamp | nullable |
| bets_locked_at | timestamp | nullable |
| settled_at | timestamp | nullable |
| statsapi_fixture_id | varchar(64) | unique; nullable |
| ou_line | numeric(4,1) | Over/Under line, nullable |
| odds_over, odds_under | numeric(5,2) | nullable |
| ou_synced_at, odds_synced_at, poisson_synced_at | timestamp | nullable |
| created_at, updated_at | timestamp | |

### `wc_score_multipliers`
Point multipliers for **free** exact-score predictions (per scoreline per match).

| Column | Type | Notes |
|--------|------|-------|
| id | uuid PK | |
| match_id | uuid | FK → `wc_matches` |
| home_score, away_score | int | unique composite with match_id |
| multiplier | numeric(5,2) | e.g. 5.0× |
| created_at, updated_at | timestamp | |

### `wc_score_odds`
Money odds for **bet** exact-score lines (separate from prediction multipliers).

| Column | Type | Notes |
|--------|------|-------|
| id | uuid PK | |
| match_id | uuid | FK → `wc_matches` |
| home_score, away_score | int | unique composite with match_id |
| odds | numeric(5,2) | |
| created_at, updated_at | timestamp | |

### `wc_wallets`
One wallet per WC user.

| Column | Type | Notes |
|--------|------|-------|
| id | uuid PK | |
| wc_user_id | uuid | FK → `wc_users`, unique |
| balance | numeric(10,2) | default 0 |
| created_at, updated_at | timestamp | |

### `wc_wallet_logs`
Audit trail of admin balance adjustments.

| Column | Type | Notes |
|--------|------|-------|
| id | uuid PK | |
| wc_user_id | uuid | FK → `wc_users` |
| admin_id | uuid | FK → `wc_users` (who made the change) |
| delta | numeric(10,2) | positive = credit, negative = debit |
| balance_before, balance_after | numeric(10,2) | |
| note | varchar(255) | |
| created_at | timestamp | |

### `wc_predictions`
Free/point-based match predictions.

| Column | Type | Notes |
|--------|------|-------|
| id | uuid PK | |
| wc_user_id | uuid | FK → `wc_users` |
| match_id | uuid | FK → `wc_matches` |
| prediction_type | varchar(15) | `handicap` \| `exact_score` \| `over_under` |
| points | int | stake (points wagered) |
| multiplier_snapshot | numeric(5,2) | multiplier at time of prediction |
| prediction_choice | varchar(5) | `home`/`away` (handicap) or `over`/`under` (O/U); nullable for exact_score |
| handicap_snapshot | numeric(4,1) | nullable |
| handicap_team_snapshot | varchar(5) | nullable |
| predicted_home_score, predicted_away_score | int | nullable (exact_score only) |
| result | varchar(10) | `correct` \| `incorrect` \| `void` \| `win_half` \| `lose_half`; nullable |
| points_earned | numeric(10,2) | nullable |
| created_at | timestamp | |

Unique constraints prevent duplicate predictions:
- Handicap/O/U: `(wc_user_id, match_id, prediction_type, prediction_choice)`
- Exact score: `(wc_user_id, match_id, predicted_home_score, predicted_away_score)`

### `wc_bets`
Real-money bets deducted from wallet balance.

| Column | Type | Notes |
|--------|------|-------|
| id | uuid PK | |
| wc_user_id | uuid | FK → `wc_users` |
| match_id | uuid | FK → `wc_matches` |
| bet_type | varchar(15) | `handicap` \| `exact_score` \| `over_under` |
| bet_choice | varchar(5) | `home`/`away`/`over`/`under`; nullable for exact_score |
| stake | int | |
| odds_snapshot | numeric(5,2) | odds at bet placement |
| handicap_snapshot | numeric(5,2) | nullable |
| handicap_team_snapshot | varchar(5) | nullable |
| predicted_home_score, predicted_away_score | int | nullable (exact_score only) |
| result | varchar(10) | `win` \| `lose` \| `push` \| `win_half` \| `lose_half`; nullable |
| payout | numeric(10,2) | nullable; set on settlement |
| created_at, updated_at | timestamp | |

Same dedup constraints as `wc_predictions` (handicap + exact_score composite).

### `wc_settlements`
WC payout sessions — admin batches all user balances into pay/collect rows.

| Column | Type | Notes |
|--------|------|-------|
| id | uuid PK | |
| name | varchar(100) | display name for this settlement |
| point_rate | numeric(10,2) | conversion rate: 1 point → X VND |
| settled_by | uuid | FK → `wc_users` (admin) |
| note | varchar(255) | |
| created_at | timestamp | |

### `wc_settlement_details`
Per-user line items within a settlement.

| Column | Type | Notes |
|--------|------|-------|
| id | uuid PK | |
| settlement_id | uuid | FK → `wc_settlements` |
| wc_user_id | uuid | FK → `wc_users`; unique per settlement |
| final_balance | numeric(10,2) | wallet balance at settlement time |
| amount | numeric(12,2) | VND amount to pay or collect |
| direction | varchar(10) | `pay` \| `collect` \| `even` |
| status | varchar(20) | `pending` \| `done` |
| completed_at | timestamp | nullable |
| done_note | varchar(255) | |
| created_at | timestamp | |

### `wc_sync_logs`
StatsAPI sync run history.

| Column | Type | Notes |
|--------|------|-------|
| id | uuid PK | |
| trigger | varchar(20) | `manual` \| `cron` |
| sync_type | varchar(20) | type of sync performed |
| triggered_by | uuid | FK → `wc_users`, nullable (null = cron) |
| matches_updated | int | |
| matches_failed | int | |
| error_detail | text | nullable |
| created_at | timestamp | |

---

## WC Champion Prediction

### `wc_champion_teams`
List of 48 WC2026 teams with prediction odds.

| Column | Type | Notes |
|--------|------|-------|
| id | uuid PK | |
| name | varchar(100) | unique (e.g. `Argentina`) |
| code | varchar(10) | e.g. `ARG` |
| flag_emoji | varchar(10) | |
| odds | numeric(6,2) | payout multiplier (e.g. 3.5×) |
| created_at, updated_at | timestamp | |

### `wc_champion_config`
Singleton row (id = 1). Controls open/close and records the winner.

| Column | Type | Notes |
|--------|------|-------|
| id | int PK | always 1 |
| is_open | bool | default false; whether champion bets are accepted |
| winner_id | uuid | FK → `wc_champion_teams`, nullable; set on settlement |
| settled_at | timestamp | nullable |
| created_at, updated_at | timestamp | |

### `wc_champion_predictions`
One row per user — max 1 champion prediction per user.

| Column | Type | Notes |
|--------|------|-------|
| id | uuid PK | |
| wc_user_id | uuid | FK → `wc_users`, unique |
| team_id | uuid | FK → `wc_champion_teams` |
| points | int | stake |
| odds_snapshot | numeric(6,2) | odds at time of prediction |
| result | varchar(20) | `correct` \| `incorrect`; nullable |
| points_earned | int | nullable |
| created_at, updated_at | timestamp | |

---

## Key Relationships

```
users ←── match_participants ──→ matches
users ←── score_bonuses
users ←── debt_settlements (debtor)
users ←── settlement_winners ──→ debt_settlements
fund_transactions ──→ debt_settlements (optional)

tournaments ──→ tournament_participants ──→ users
tournaments ──→ tournament_teams ──→ users (player1, player2)
tournaments ──→ tournament_matches ──→ tournament_teams
tournament_matches ──→ matches (match_id link)

wc_users ──→ wc_wallets (1:1)
wc_users ──→ wc_wallet_logs
wc_users ──→ wc_predictions ──→ wc_matches
wc_users ──→ wc_bets ──→ wc_matches
wc_matches ──→ wc_score_multipliers
wc_matches ──→ wc_score_odds
wc_settlements ──→ wc_settlement_details ──→ wc_users
wc_users ──→ wc_champion_predictions ──→ wc_champion_teams
```

## System Boundaries

- `users` and `wc_users` are **completely independent** — no FK between them.
- Core esport tables (`matches`, `users`, etc.) never reference WC tables, and vice versa.
- `score_bonuses` is the only cross-cutting concept: it belongs to `users` (core system).
