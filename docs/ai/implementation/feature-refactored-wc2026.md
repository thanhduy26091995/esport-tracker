---
phase: implementation
title: Refactored WC2026 — Implementation Guide
description: File-level implementation notes for the three WC2026 enhancements.
---

# Implementation Guide

## Development Setup

No new dependencies or env vars required. Standard dev workflow:
```bash
# backend
cd backend && go run ./cmd/server

# frontend
cd frontend && npm run dev
```

---

## Code Structure

Changes are confined to:
```
backend/internal/
  repository/match_repository.go   ← new GetAllFiltered method
  service/match_service.go          ← updated GetAll signature
  api/match_handler.go              ← parse player_id query param

frontend/src/
  services/matchService.ts          ← optional playerID param
  stores/matchStore.ts              ← forward playerID to service
  views/MatchesView.vue             ← player dropdown UI
  components/wc/WcMatchCard.vue     ← timezone fix + live glow
  views/WcAdminView.vue             ← new file (admin page shell)
  router/index.ts                   ← new route + guard
  views/WcPredictView.vue           ← remove admin tab
```

---

## Implementation Notes

### Feature 1 — Player filter

**Repository (`match_repository.go`)**

```go
func (r *MatchRepository) GetAllFiltered(limit, offset int, playerID *uuid.UUID) ([]*model.Match, error) {
    var matches []*model.Match
    query := r.db.Preload("Participants.User").Order("match_date DESC, created_at DESC")
    if playerID != nil {
        query = query.
            Joins("JOIN match_participants mp ON mp.match_id = matches.id").
            Where("mp.user_id = ?", *playerID).
            Distinct()
    }
    if limit > 0 {
        query = query.Limit(limit).Offset(offset)
    }
    return matches, query.Find(&matches).Error
}
```

The existing `GetAll` can delegate to this method with `nil`:
```go
func (r *MatchRepository) GetAll(limit, offset int) ([]*model.Match, error) {
    return r.GetAllFiltered(limit, offset, nil)
}
```

**Service (`match_service.go`)** — update `GetAllMatches(limit, offset int)` to `GetAllMatches(limit, offset int, playerID *uuid.UUID)`. Delegate to `GetAllFiltered`. All existing callers that pass `(0, 0)` now pass `(0, 0, nil)`.

**Handler (`match_handler.go`)** — in the `GetAll` handler:
```go
playerIDStr := c.Query("player_id")
var playerID *uuid.UUID
if playerIDStr != "" {
    parsed, err := uuid.Parse(playerIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_UUID", "message": "invalid player_id"})
        return
    }
    playerID = &parsed
}
// pass playerID to GetAllMatches; bonuses fetch and buildFeed are unchanged
matches, err := h.matchService.GetAllMatches(0, 0, playerID)
```

**Frontend types** (`types/api.ts`) — add:
```ts
export interface MatchFilterParams extends PaginationParams {
  player_id?: string
}
```

**Frontend service** — update `getAll(params?: PaginationParams)` to `getAll(params?: MatchFilterParams)`. Axios forwards all keys in `params` as query string params automatically.

**Frontend store** — update `fetchMatches` params type to `MatchFilterParams`; forward `player_id` field.

**Frontend view** — use `el-select` with an initial "All players" option (`value: ''`). On change re-fetch with `{ player_id: selectedPlayerID.value || undefined }`. Wrap the 3 stat cards in `v-if="!selectedPlayerID"`.

---

### Feature 2 — Timezone + live glow

**`WcMatchCard.vue` computed**
```ts
const matchDateStr = computed(() =>
  matchDate.value.toLocaleDateString('vi-VN', {
    day: '2-digit', month: '2-digit', timeZone: 'Asia/Ho_Chi_Minh'
  })
)
const matchTimeStr = computed(() =>
  matchDate.value.toLocaleTimeString('vi-VN', {
    hour: '2-digit', minute: '2-digit', timeZone: 'Asia/Ho_Chi_Minh'
  })
)
```

**`WcMatchCard.vue` CSS** — replace the existing `.wc-match-card--live` block:
```css
.wc-match-card--live {
  border-color: #16a34a;
  background: linear-gradient(135deg, rgba(22, 163, 74, 0.05), var(--surface-card));
  animation: glow-live 2s ease-in-out infinite alternate;
}
@keyframes glow-live {
  from { box-shadow: 0 0 0 2px rgba(22, 163, 74, 0.20), 0 0 8px  rgba(22, 163, 74, 0.12); }
  to   { box-shadow: 0 0 0 2px rgba(22, 163, 74, 0.40), 0 0 20px rgba(22, 163, 74, 0.28); }
}
```

The existing `.wc-badge--live` pulsing badge is unchanged — the glow is additive.

---

### Feature 3 — Admin page

**`WcAdminView.vue`** structure (mirrors `WcPredictView` header pattern):
```vue
<template>
  <div class="page-wrapper">
    <div class="page-container">
      <div class="page-header">
        <h1 class="page-title">🏆 World Cup 2026 — Admin</h1>
        <div class="wc-user-header">
          <span class="wc-user-name">{{ authStore.userName }}</span>
          <span class="wc-admin-badge">Admin</span>
          <el-button size="small" text @click="handleLogout">Đăng xuất</el-button>
        </div>
      </div>
      <WcAdminPanel />
    </div>
  </div>
</template>
```

Import `WcAdminPanel`, `useWcAuthStore`, and the router. Implement `handleLogout` identically to `WcPredictView`.

**Router guard snippet** — reads `wc_user` (already JSON, stored by `useWcAuthStore`), not the raw JWT:
```ts
if (to.meta.requiresWcAdmin) {
  try {
    const raw = localStorage.getItem('wc_user')
    const user = raw ? JSON.parse(raw) : null
    if (!user?.isAdmin) return { name: 'wc-schedule' }
  } catch {
    return { name: 'wc-schedule' }
  }
}
```

**`WcPredictView.vue` cleanup** — remove:
- `<el-tab-pane label="..." name="admin">` block and its content.
- Any ref/computed only used inside that tab (e.g., `adminSearch`, `adminFilter`, `adminFilterOptions`, `adminFiltered`, `handleSync`, `handleSetOdds`, etc.). Confirm each is not used elsewhere before removing.

---

## Integration Points

- `matchService.ts` calls `GET /api/v1/matches` — add `player_id` as an optional URL param.
- WC admin panel calls `POST /api/v1/wc/admin/*` — no changes, the panel component is moved not rewritten.

---

## Error Handling

- Invalid `player_id` UUID → backend returns `400 { "error": "invalid player_id" }`; frontend should show a toast/alert (use existing error handling pattern in the store).
- `beforeEach` JWT parse error → treat as unauthenticated, redirect to `wc-schedule`.

---

## Security Notes

- `/world-cup/admin` is guarded at the router (client-side JWT check). This is UX-only. The real security is the existing `WcAdminMiddleware` on all `/api/v1/wc/admin/*` endpoints — no change needed there.
- `player_id` is parsed as UUID server-side; SQL injection is not possible via GORM parameterised queries.
