---
phase: implementation
title: Win Rate & Tier Evaluation — Implementation Guide
description: Key technical patterns and integration notes for implementing this feature
---

# Implementation Guide

## Development Setup

No new dependencies required. The feature uses existing:
- GORM + PostgreSQL (backend)
- Vue 3 + Element Plus + vue-i18n (frontend)

## Code Structure

```
backend/internal/
  model/user.go                    ← add UserWithStats struct
  service/tier_service.go          ← NEW: tier evaluation logic
  service/match_service.go         ← inject TierService, call after match mutations
  service/user_service.go          ← update return types
  repository/user_repository.go    ← win rate JOIN query + UpdateTier method
  migrations/                      ← optional: index on match_participants(user_id)

frontend/src/
  types/user.ts                    ← add UserWithStats interface
  stores/userStore.ts              ← update type references
  services/userService.ts          ← update return types
  components/UserTable.vue         ← add Win Rate column
  components/WinRateBadge.vue      ← NEW: tier chip + win rate display
  locales/en.json, vi.json         ← add tier label keys
```

## Implementation Notes

### Core Features

**Win rate SQL (repository layer):**
```go
type userWinRateRow struct {
    model.User
    TotalMatches int     `gorm:"column:total_matches"`
    WonMatches   int     `gorm:"column:won_matches"`
    WinRate      float64 `gorm:"column:win_rate"`
}

// Use raw query or GORM with Select + Joins
db.Table("users u").
    Select(`u.*,
        COUNT(mp.id) AS total_matches,
        COUNT(mp.id) FILTER (WHERE mp.point_change > 0) AS won_matches,
        CASE WHEN COUNT(mp.id) = 0 THEN 0
             ELSE COUNT(mp.id) FILTER (WHERE mp.point_change > 0)::float / COUNT(mp.id)
        END AS win_rate`).
    Joins("LEFT JOIN match_participants mp ON mp.user_id = u.id").
    Group("u.id").
    Find(&rows)
```

**Tier evaluation (pure function in tier_service.go):**
```go
const (
    TierPro    = "pro"
    TierNormal = "normal"
    TierNoob   = "noob"

    MinMatchesForTier = 10
    ProThreshold      = 0.60
    NormalThreshold   = 0.40
)

func EvaluateTier(winRate float64, totalMatches int) string {
    if totalMatches < MinMatchesForTier {
        return TierNormal // default until enough matches
    }
    switch {
    case winRate >= ProThreshold:
        return TierPro
    case winRate >= NormalThreshold:
        return TierNormal
    default:
        return TierNoob
    }
}
```

**Tier recalculation trigger in MatchService:**
```go
func (s *MatchService) CreateMatch(req CreateMatchRequest) (*model.Match, error) {
    match, err := s.matchRepo.Create(req)
    if err != nil {
        return nil, err
    }
    participantIDs := extractUserIDs(match.Participants)
    _ = s.tierService.RecalculateForUsers(participantIDs) // non-fatal
    return match, nil
}
```

**WinRateBadge.vue (key logic):**
```vue
<script setup lang="ts">
const props = defineProps<{ tier: string; winRate: number; totalMatches: number }>()
const isUnranked = computed(() => props.totalMatches < 10)
const displayRate = computed(() => `${Math.round(props.winRate * 100)}%`)
const tierColor = { pro: '#f59e0b', normal: '#3b82f6', noob: '#6b7280' }
</script>

<template>
  <div class="win-rate-cell">
    <el-tag :color="isUnranked ? '#e5e7eb' : tierColor[tier]" size="small">
      {{ isUnranked ? t('tier.unranked') : t(`tier.${tier}`) }}
    </el-tag>
    <span class="win-rate-pct">{{ isUnranked ? '—' : displayRate }}</span>
  </div>
</template>
```

### Patterns & Best Practices

- `EvaluateTier` is a pure function — keep it free of DB calls so it's easily unit-tested
- Tier recalculation errors in `RecalculateForUsers` are logged but not fatal — a failed tier update should not roll back the match creation
- Win rate column in `UserTable` should be conditionally hidden on mobile using Element Plus `v-if="!isMobile"` or responsive hide classes

## Integration Points

- `MatchService.CreateMatch` and `MatchService.DeleteMatch` → `TierService.RecalculateForUsers`
- `UserRepository.GetAll` / `GetLeaderboard` / `GetByID` → new SQL with win rate JOIN
- `UserTable` → `WinRateBadge` (reused in both UsersView and DashboardView)

## Error Handling

- If `RecalculateForUsers` fails, log the error and continue — match creation still succeeds
- If win rate query returns no rows for a user (e.g. LEFT JOIN yields null), treat as `total_matches=0, win_rate=0.0`

## Performance Considerations

- LEFT JOIN on `match_participants(user_id)` — ensure this index exists (migration task 1.6)
- At current scale (< 100 users, < 10k matches) the GROUP BY query is fast enough without caching
- If performance degrades: add `win_rate` and `total_matches` as stored columns, updated via trigger or on match events

## Security Notes

- No new input surface — win rate is computed server-side from existing match data
- `tier` update via `UpdateTier` only accepts the three valid values; validate with a whitelist before writing
