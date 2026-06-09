---
phase: design
title: "1v2 Match Type — System Design"
description: Technical architecture for supporting asymmetric 1-vs-2 match format
---

# System Design & Architecture

## Architecture Overview

The change is additive: no new tables or services — only validation logic and scoring calculation are extended.

```mermaid
graph TD
  FE["Frontend\nMatchForm"] -->|POST /api/v1/matches\nmatch_type:1v2| Handler
  Handler -->|CreateMatchRequest| MatchService
  MatchService -->|validate teams| Validation["Team Size\nValidation"]
  MatchService -->|calcPoints()| ScoreCalc["Asymmetric\nScore Calc"]
  ScoreCalc --> DB[(matches\nmatch_participants\nusers)]
  MatchService --> TierService
  MatchService --> SettlementService
```

No schema migration is required. `match_type VARCHAR(10)` already holds `"1v2"`.

## Data Models

### Match (no change)
`match_type` column now accepts a third valid value: `"1v2"`.

### MatchParticipant (no change)
`point_change` already supports any integer — it will store `+2 / -2 / +1 / -1` as needed.

### Scoring rules for `1v2`

| Scenario | team1 (solo) `point_change` | each team2 player `point_change` |
|---|---|---|
| team1 wins | `+2 × base` | `-1 × base` |
| team2 wins | `-2 × base` | `+1 × base` |
| draw (0) | `0` | `0` |

Where `base` = `points_per_win` from config (or request override). Net change per match = 0 (zero-sum).

## API Design

### `POST /api/v1/matches` — updated validation

**New valid request (1v2):**
```json
{
  "match_type": "1v2",
  "team1": ["<uuid-solo>"],
  "team2": ["<uuid-a>", "<uuid-b>"],
  "winner_team": 1,
  "match_date": "2026-06-04T20:00:00Z"
}
```

**Validation changes:**
- `match_type` enum: `"1v1" | "2v2" | "1v2"`
- Team size map:
  - `"1v1"` → team1 = 1, team2 = 1 (unchanged)
  - `"2v2"` → team1 = 2, team2 = 2 (unchanged)
  - `"1v2"` → team1 = 1, team2 = 2 (new)
- Error on mismatch: HTTP 400 `INVALID_TEAM_SIZE`

**Error response update** in `match_handler.go`:
```go
case "match_type must be '1v1', '2v2', or '1v2'":
    statusCode = http.StatusBadRequest
    code = "INVALID_MATCH_TYPE"
```

## Component Breakdown

### Backend: `internal/service/match_service.go`

1. **Validation block** — extend the `MatchType` check and the team-size `expectedSize` logic to a map:
   ```go
   teamSizes := map[string][2]int{
     "1v1": {1, 1},
     "2v2": {2, 2},
     "1v2": {1, 2},
   }
   ```

2. **`calcPointChange(matchType, teamNumber, winnerTeam, base int)`** — extract a helper that returns the correct `point_change` per participant, handling the asymmetric `1v2` case:
   ```go
   // 1v2: solo wins → solo +2*base, each duo -1*base
   //      duo wins  → solo -2*base, each duo +1*base
   ```

3. **Error message** — update the literal string to include `'1v2'`.

### Backend: `internal/api/match_handler.go`

- Update the `switch err.Error()` block to match the new error message string.

### Frontend: `frontend/src/types/match.ts`

```ts
export type MatchType = '1v1' | '2v2' | '1v2'
```

### Frontend: `MatchForm` component — team slot limits

The form uses `multiple-limit` per team. For `1v2`, team1 = 1 slot, team2 = 2 slots. Update the computed limits:

```ts
const team1Limit = computed(() => formData.match_type === '2v2' ? 2 : 1)
// 1v1=1, 1v2=1, 2v2=2

const team2Limit = computed(() => formData.match_type === '1v1' ? 1 : 2)
// 1v1=1, 1v2=2, 2v2=2
```

Bind these to the respective `el-select` `:multiple-limit` props. Also add `1v2` radio button:
```html
<el-radio-button value="1v2">{{ t('matches.types.oneVsTwo') }}</el-radio-button>
```

Show a helper note when `1v2` is selected: "Solo vs Duo — solo thắng: +2đ, duo thắng: +1đ mỗi người".

### Frontend: `MatchList` component

Two concrete updates:

**1. Type filter** — add `1v2` option:
```html
<el-option :label="t('matches.types.oneVsTwo')" value="1v2" />
```

**2. Match type CSS class** — current ternary only handles 1v1/2v2, defaults non-1v1 to 2v2 styling. Replace with:
```ts
const matchTypeClass = (type: string) => ({
  'match-type--1v1': type === '1v1',
  'match-type--2v2': type === '2v2',
  'match-type--1v2': type === '1v2',
})
```

## Design Decisions

| Decision | Chosen approach | Alternative considered |
|---|---|---|
| Scoring formula | Zero-sum: solo×2 = duo×1×2 | Non-zero-sum custom points |
| Solo always on team1 | Convention, not enforced by label | Add `is_solo_side` flag |
| Base multiplier | `base × 2` for solo win, `base × 1` for duo win | Hard-code 2/1 ignoring base |
| Schema migration | None — VARCHAR(10) fits `"1v2"` | Add `team1_size`/`team2_size` columns |

## Security & Performance

- No new attack surface — same endpoint, same auth
- No performance impact — score calc is O(n) on participants (n ≤ 3)
