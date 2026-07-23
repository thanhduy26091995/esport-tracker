---
phase: implementation
title: ASEAN Cup 2026 — Implementation Guide
description: Technical notes for the tournament_type discriminator migration and ASEAN Cup route wiring
---

# Implementation Guide

## Development Setup

No new dependencies required. All existing Go and Vue tooling applies.

```bash
# Backend
cd backend && go run cmd/server/main.go

# Frontend
cd frontend && npm run dev

# Type check
cd frontend && npm run type-check
```

## Code Structure

### New Files

```
backend/internal/middleware/tournament.go      # TournamentMiddleware
frontend/src/config/tournaments.ts             # TournamentConfig + TOURNAMENTS map
frontend/src/composables/useTournament.ts      # Route-derived tournament context
```

### Modified Files (Backend)

```
backend/internal/model/wc_*.go                 # Add TournamentType field
backend/internal/repository/wc_*_repository.go # Add tournamentType param
backend/internal/service/wc_service.go         # Add tournamentType param
backend/internal/service/wc_custom_bet_service.go
backend/internal/api/wc_handler.go             # Read from Gin context
backend/internal/api/wc_custom_bet_handler.go
backend/internal/api/wc_champion_handler.go
backend/internal/api/router.go                 # New /ac groups + middleware
```

### Modified Files (Frontend)

```
frontend/src/router/index.ts                   # /asean-cup/* routes + guard
frontend/src/services/wcService.ts             # Accept apiPrefix
frontend/src/stores/wcAuthStore.ts             # Tournament-aware config fetch
frontend/src/components/layout/NavSidebar.vue  # ASEAN Cup nav section
frontend/src/locales/vi.json
frontend/src/locales/en.json
```

## Implementation Notes

### Core Features

#### 1. TournamentMiddleware (Go)

```go
// backend/internal/middleware/tournament.go
package middleware

import "github.com/gin-gonic/gin"

func TournamentMiddleware(tournamentType string) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Set("tournament_type", tournamentType)
        c.Next()
    }
}
```

#### 2. Router wiring pattern

In `router.go`, apply middleware to all existing WC groups, then duplicate for AC:

```go
// Existing WC groups — add middleware
wcPublic := v1.Group("/wc")
wcPublic.Use(middleware.TournamentMiddleware("world_cup"))

wcAuth := v1.Group("/wc")
wcAuth.Use(middleware.TournamentMiddleware("world_cup"), WcAuthMiddleware())

wcAdmin := v1.Group("/wc/admin")
wcAdmin.Use(middleware.TournamentMiddleware("world_cup"), WcAuthMiddleware(), WcAdminMiddleware())

// New ASEAN Cup groups
acPublic := v1.Group("/ac")
acPublic.Use(middleware.TournamentMiddleware("asean_cup"))

acAuth := v1.Group("/ac")
acAuth.Use(middleware.TournamentMiddleware("asean_cup"), WcAuthMiddleware())

acAdmin := v1.Group("/ac/admin")
acAdmin.Use(middleware.TournamentMiddleware("asean_cup"), WcAuthMiddleware(), WcAdminMiddleware())

// Register identical handler functions on both groups
registerWcPublicRoutes(wcPublic, wcHandler)
registerWcPublicRoutes(acPublic, wcHandler)
registerWcAuthRoutes(wcAuth, wcHandler, customBetHandler, championHandler)
registerWcAuthRoutes(acAuth, wcHandler, customBetHandler, championHandler)
registerWcAdminRoutes(wcAdmin, wcHandler, customBetHandler, championHandler)
registerWcAdminRoutes(acAdmin, wcHandler, customBetHandler, championHandler)
```

Extract route registration into `registerWcPublicRoutes(g *gin.RouterGroup, h *WcHandler)` etc. to avoid duplication.

#### 3. Handler pattern

```go
func (h *WcHandler) ListMatches(c *gin.Context) {
    tournamentType := c.MustGet("tournament_type").(string)
    // pass to service
    matches, err := h.wcService.ListMatches(c.Request.Context(), tournamentType, ...)
}
```

#### 4. `wc_config` — retire id=1 singleton

Before:
```go
db.Where("id = ?", 1).First(&config)
db.Where("id = ?", 1).Updates(&config)
```

After:
```go
db.Where("tournament_type = ?", tournamentType).First(&config)
db.Where("tournament_type = ?", tournamentType).Updates(&config)
```

Search for all `id = 1` and `id: 1` references in the WC config code and update.

#### 5. Frontend `useTournament` composable

```typescript
// src/composables/useTournament.ts
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { TOURNAMENTS, type TournamentConfig } from '@/config/tournaments'

export function useTournament(): { tournamentConfig: ComputedRef<TournamentConfig> } {
  const route = useRoute()
  const tournamentConfig = computed<TournamentConfig>(() => {
    if (route.path.startsWith('/asean-cup')) return TOURNAMENTS.asean_cup
    return TOURNAMENTS.world_cup
  })
  return { tournamentConfig }
}
```

#### 6. `wcService.ts` — apiPrefix injection

Two options (prefer option A):

**Option A — Factory function** (cleaner, reusable):
```typescript
export function createWcService(apiPrefix: string) {
  return {
    getMatches: (params) => api.get(`${apiPrefix}/matches`, { params }),
    placeBet: (matchId, data) => api.post(`${apiPrefix}/matches/${matchId}/bet`, data),
    // ... all methods using apiPrefix
  }
}
```

Views instantiate: `const wcService = createWcService(tournamentConfig.value.apiPrefix)`

**Option B — Pass apiPrefix per call** (simpler but verbose)

#### 7. Route guard — ASEAN Cup config fetch

```typescript
// router/index.ts — extend existing beforeEach
router.beforeEach(async (to) => {
  if (to.path.startsWith('/asean-cup')) {
    const config = await fetch('/api/v1/ac/config').then(r => r.json())
    if (!config.is_enabled && !to.path.endsWith('/schedule')) {
      return '/asean-cup/schedule'
    }
  }
  // ... existing WC guard logic
})
```

### Patterns & Best Practices

- **Never hardcode `"world_cup"` strings** in handler bodies — always read from context
- **Migration safety**: add `tournament_type` with DEFAULT first, backfill, then add NOT NULL constraint in a separate step
- **grep audit before starting**: `grep -rn "id = 1\|id: 1\|wc_config" backend/internal/` to find all config singleton references
- **Test backward compatibility**: after migration, smoke-test all existing `/wc/*` endpoints to confirm they still return correct data

## Integration Points

### GORM AutoMigrate

Add `TournamentType string \`gorm:"column:tournament_type;not null;default:world_cup"\`` to relevant model structs. AutoMigrate will add the column (safe, non-destructive). The DEFAULT handles existing rows automatically.

### Background Cron (match sync)

The existing WC match sync cron passes `statsapi_fixture_id` matches as `"world_cup"`. A new ASEAN Cup sync (if StatsAPI covers it) would use `"asean_cup"`. If not available, skip the cron — admin creates matches manually via the admin API.

## Error Handling

- If `tournament_type` key is missing from Gin context (middleware not applied), `MustGet` panics — this is intentional (programming error, not a user error)
- If `wc_config` row for a tournament doesn't exist, return a clear error: `"tournament config not found"` with 404 — don't fall back to WC config

## Performance Considerations

- Add all 5 `tournament_type` indexes (Phase 1.2) **before** wiring the new `/ac` routes — queries will be fast from day one
- The existing WC queries gain a `WHERE tournament_type = 'world_cup'` clause — this uses the index and is no slower than before (the index reduces scan size)

## Security Notes

- Same JWT secret, same `WcAuthMiddleware` — no new auth surface
- `TournamentMiddleware` is server-controlled (not user-supplied) — no injection risk
- Admin endpoints on `/ac/admin/*` are guarded by the same `WcAdminMiddleware` as WC
