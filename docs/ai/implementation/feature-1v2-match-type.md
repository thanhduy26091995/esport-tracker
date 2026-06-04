---
phase: implementation
title: "1v2 Match Type — Implementation Guide"
description: Exact code locations and change instructions for 1v2 match support
---

# Implementation Guide

## Key Files to Change

| File | Change |
|---|---|
| `backend/internal/service/match_service.go` | Validation map + asymmetric scoring |
| `backend/internal/api/match_handler.go` | Error string in switch-case |
| `frontend/src/types/match.ts` | Add `'1v2'` to `MatchType` |
| `frontend/src/views/CreateMatchView.vue` (or MatchForm) | Dynamic team slots |
| `frontend/src/locales/vi.json` + `en.json` | i18n label |

## Implementation Notes

### Validation map (`match_service.go`)

Replace:
```go
expectedSize := 1
if req.MatchType == "2v2" {
    expectedSize = 2
}
if len(req.Team1) != expectedSize || len(req.Team2) != expectedSize {
    return nil, fmt.Errorf("each team must have exactly %d player(s) for %s", expectedSize, req.MatchType)
}
```

With:
```go
type teamSize struct{ t1, t2 int }
sizes := map[string]teamSize{
    "1v1": {1, 1},
    "2v2": {2, 2},
    "1v2": {1, 2},
}
sz, ok := sizes[req.MatchType]
if !ok {
    return nil, errors.New("match_type must be '1v1', '2v2', or '1v2'")
}
if len(req.Team1) != sz.t1 || len(req.Team2) != sz.t2 {
    return nil, fmt.Errorf("for %s, team1 needs %d player(s) and team2 needs %d player(s)", req.MatchType, sz.t1, sz.t2)
}
```

Remove the old separate match_type string check (it's now covered by the map lookup).

### Asymmetric scoring (`match_service.go`)

Add helper before `CreateMatch`:
```go
func calcPointChange(matchType string, teamNumber, winnerTeam, base int) int {
    if winnerTeam == 0 {
        return 0
    }
    won := (teamNumber == winnerTeam)
    if matchType == "1v2" {
        if teamNumber == 1 { // solo
            if won { return 2 * base } else { return -2 * base }
        } else { // duo member
            if won { return base } else { return -base }
        }
    }
    // symmetric: 1v1, 2v2
    if won { return base } else { return -base }
}
```

Replace the inline `pointChange` calculations in the team1 and team2 loops with:
```go
pointChange := calcPointChange(req.MatchType, 1, req.WinnerTeam, basePoints) // team 1 loop
pointChange := calcPointChange(req.MatchType, 2, req.WinnerTeam, basePoints) // team 2 loop
```

### Handler error update (`match_handler.go`)

Change:
```go
case "match_type must be '1v1' or '2v2'":
```
To:
```go
case "match_type must be '1v1', '2v2', or '1v2'":
```

### Frontend type (`match.ts`)

```ts
export type MatchType = '1v1' | '2v2' | '1v2'
```

### Frontend form (CreateMatchView.vue / MatchForm)

Add to match type options:
```ts
const matchTypeOptions = [
  { value: '1v1', label: '1v1', hint: '' },
  { value: '2v2', label: '2v2', hint: '' },
  { value: '1v2', label: '1v2', hint: 'Solo vs Duo — solo thắng: +2đ, duo thắng: +1đ mỗi người' },
]
```

Make team slot counts reactive:
```ts
const team1Size = computed(() => form.matchType === '2v2' ? 2 : 1)
const team2Size = computed(() => form.matchType === '1v2' ? 2 : (form.matchType === '2v2' ? 2 : 1))
```

## Integration Points

- `TierService.RecalculateForUsers` — called post-commit, already handles any set of user IDs; no change needed
- `SettlementService.CheckAndTriggerSettlement` — same, no change needed
- `DeleteMatch` — reverting stored `point_change` values is already generic; no change needed
